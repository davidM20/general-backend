package services

import (
	"database/sql"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

var notificationDB *sql.DB

// InitializeNotificationService inyecta la dependencia de la BD.
func InitializeNotificationService(db *sql.DB) {
	notificationDB = db
	logger.Info("SERVICE_NOTIFICATION", "NotificationService inicializado con conexión a BD.")
}

// ProcessAndSendNotification crea un evento, lo guarda en la BD y lo envía al usuario si está conectado.
func ProcessAndSendNotification(userIDToNotify int64, eventType string, title string, message string, relatedData map[string]interface{}, manager *customws.ConnectionManager[wsmodels.WsUserData]) error {
	if notificationDB == nil {
		return fmt.Errorf("NotificationService no inicializado")
	}

	event := models.Event{
		EventType:   eventType,
		EventTitle:  title,
		Description: message,
		UserId:      userIDToNotify,
		IsRead:      false,
		// CreateAt se establecerá en queries.CreateEvent o por la BD
	}

	// Asignar OtherUserId y ProyectId desde relatedData si existen
	if otherUserIDVal, ok := relatedData["otherUserId"]; ok {
		if otherUserID, castOk := otherUserIDVal.(int64); castOk {
			event.OtherUserId = sql.NullInt64{Int64: otherUserID, Valid: true}
		}
	}
	if projectIDVal, ok := relatedData["projectId"]; ok {
		if projectID, castOk := projectIDVal.(int64); castOk {
			event.ProyectId = sql.NullInt64{Int64: projectID, Valid: true}
		}
	}

	if err := queries.CreateEvent(notificationDB, &event); err != nil {
		logger.Errorf("SERVICE_NOTIFICATION", "Error creando evento para UserID %d: %v", userIDToNotify, err)
		return fmt.Errorf("error creando evento: %w", err)
	}

	logger.Successf("SERVICE_NOTIFICATION", "Evento (ID: %d, Tipo: %s) creado para UserID %d", event.Id, eventType, userIDToNotify)

	// Construir el payload para wsmodels.NotificationInfo
	wsPayload := make(map[string]interface{})
	if event.OtherUserId.Valid {
		wsPayload["otherUserId"] = event.OtherUserId.Int64
	}
	if event.ProyectId.Valid {
		wsPayload["projectId"] = event.ProyectId.Int64
	}
	for key, value := range relatedData {
		if key != "otherUserId" && key != "projectId" {
			wsPayload[key] = value
		}
	}

	notificationForClient := wsmodels.NotificationInfo{
		ID:        fmt.Sprintf("%d", event.Id),
		Type:      event.EventType,
		Title:     event.EventTitle,
		Message:   event.Description,
		Timestamp: event.CreateAt,
		IsRead:    event.IsRead,
		Payload:   wsPayload,
		// Profile se poblará a continuación si OtherUserId existe
	}

	if event.OtherUserId.Valid {
		otherUserInfo, err := queries.GetUserBaseInfo(notificationDB, event.OtherUserId.Int64)
		if err != nil {
			logger.Warnf("SERVICE_NOTIFICATION", "Error obteniendo UserBaseInfo para OtherUserId %d para notificación en tiempo real: %v", event.OtherUserId.Int64, err)
		} else if otherUserInfo != nil {
			notificationForClient.Profile = wsmodels.ProfileData{
				ID:        otherUserInfo.ID,
				FirstName: otherUserInfo.FirstName,
				LastName:  otherUserInfo.LastName,
				UserName:  otherUserInfo.UserName,
				Picture:   otherUserInfo.Picture,
				// El resto de los campos de ProfileData (Email, RoleName, etc.)
				// no están en models.UserBaseInfo, por lo que quedarán como sus zero values.
			}
		}
	}

	// DEBUG: Loguear la notificación ANTES de enviarla
	logger.Debugf("SERVICE_NOTIFICATION", "Nueva notificación para UserID %d (antes de enviar): ID=%s, Type=%s, Title=%s, ProfileID=%d, ProfileName=%s, ProfilePic=%s, Payload=%+v",
		userIDToNotify, notificationForClient.ID, notificationForClient.Type, notificationForClient.Title, notificationForClient.Profile.ID, notificationForClient.Profile.FirstName+" "+notificationForClient.Profile.LastName, notificationForClient.Profile.Picture, notificationForClient.Payload)

	if manager.IsUserOnline(userIDToNotify) {
		serverMessage := types.ServerToClientMessage{
			PID:     manager.Callbacks().GeneratePID(),
			Type:    types.MessageTypeNewNotification,
			Payload: notificationForClient,
		}
		if err := manager.SendMessageToUser(userIDToNotify, serverMessage); err != nil {
			logger.Warnf("SERVICE_NOTIFICATION", "Error enviando notificación (ID: %d) a UserID %d online: %v", event.Id, userIDToNotify, err)
			// No devolver error aquí, la notificación está guardada, el envío falló pero puede recuperarse luego.
		} else {
			logger.Infof("SERVICE_NOTIFICATION", "Notificación (ID: %d) enviada a UserID %d online.", event.Id, userIDToNotify)
		}
	} else {
		logger.Infof("SERVICE_NOTIFICATION", "Usuario %d no está online. Notificación (ID: %d) guardada.", userIDToNotify, event.Id)
		// Aquí podría ir la lógica para una notificación push si estuviera implementada.
	}

	return nil
}

// GetNotifications recupera notificaciones para un usuario.
func GetNotifications(userID int64, onlyUnread bool, limit int, offset int) ([]wsmodels.NotificationInfo, error) {
	if notificationDB == nil {
		return nil, fmt.Errorf("NotificationService no inicializado")
	}
	// La función queries.GetNotificationsForUser ya hace la conversión a wsmodels.NotificationInfo
	// y construye un payload básico.
	// Si se necesita un payload más enriquecido (ej. con nombres de usuario, etc.), se podría hacer aquí
	// o modificar queries.GetNotificationsForUser.
	notifications, err := queries.GetNotificationsForUser(notificationDB, userID, onlyUnread, limit, offset)
	if err != nil {
		logger.Errorf("SERVICE_NOTIFICATION", "Error obteniendo notificaciones para UserID %d: %v", userID, err)
		return nil, err
	}
	return notifications, nil
}

// MarkRead marca una notificación como leída.
func MarkRead(userID int64, notificationID string) error {
	if notificationDB == nil {
		return fmt.Errorf("NotificationService no inicializado")
	}
	err := queries.MarkNotificationAsRead(notificationDB, notificationID, userID)
	if err != nil {
		logger.Errorf("SERVICE_NOTIFICATION", "Error marcando notificación %s como leída para UserID %d: %v", notificationID, userID, err)
		return err
	}
	logger.Successf("SERVICE_NOTIFICATION", "Notificación %s marcada como leída para UserID %d", notificationID, userID)
	return nil
}

// MarkAllRead marca todas las notificaciones de un usuario como leídas.
func MarkAllRead(userID int64) (int64, error) {
	if notificationDB == nil {
		return 0, fmt.Errorf("NotificationService no inicializado")
	}
	rowsAffected, err := queries.MarkAllNotificationsAsRead(notificationDB, userID)
	if err != nil {
		logger.Errorf("SERVICE_NOTIFICATION", "Error marcando todas las notificaciones como leídas para UserID %d: %v", userID, err)
		return 0, err
	}
	logger.Successf("SERVICE_NOTIFICATION", "%d notificaciones marcadas como leídas para UserID %d", rowsAffected, userID)
	return rowsAffected, nil
}
