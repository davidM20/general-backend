package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

/*
TODO: Objetivo del NotificationService

El NotificationService es responsable de manejar todas las notificaciones del sistema, incluyendo:

1. Notificaciones de Contacto:
   - Solicitudes de amistad recibidas
   - Confirmaciones de solicitudes aceptadas
   - Rechazos de solicitudes

2. Notificaciones del Sistema:
   - Actualizaciones de estado
   - Mensajes importantes
   - Alertas de seguridad

3. Notificaciones de Eventos:
   - Eventos de proyecto
   - Recordatorios
   - Invitaciones

4. Características Principales:
   - Persistencia en base de datos
   - Envío en tiempo real si el usuario está online
   - Manejo de estados (leído/no leído)
   - Soporte para acciones requeridas
   - Metadata flexible para diferentes tipos de notificaciones

5. Integración:
   - Se integra con el WebSocketManager para envío en tiempo real
   - Utiliza el sistema de queries para persistencia
   - Proporciona una API clara para otros servicios
*/

var (
	notificationDB *sql.DB
)

// InitializeNotificationService inyecta las dependencias necesarias
func InitializeNotificationService(db *sql.DB) {
	notificationDB = db
	logger.Info("SERVICE_NOTIFICATION", "NotificationService inicializado con conexión a BD.")
}

// CreateFriendRequestNotification crea una notificación de solicitud de amistad
func CreateFriendRequestNotification(fromUserID, toUserID int64, requestMessage string) error {
	metadata := models.EventMetadata{
		RequestMessage: requestMessage,
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("error marshalling metadata: %w", err)
	}

	event := models.Event{
		EventType:      models.EventTypeFriendRequest,
		EventTitle:     "Nueva solicitud de amistad",
		Description:    fmt.Sprintf("Has recibido una solicitud de amistad"),
		UserId:         toUserID,
		OtherUserId:    sql.NullInt64{Int64: fromUserID, Valid: true},
		Status:         models.EventStatusPending,
		ActionRequired: true,
		Metadata:       metadataJSON,
	}

	return createAndSendNotification(event)
}

// CreateFriendRequestResponseNotification crea una notificación de respuesta a solicitud de amistad
func CreateFriendRequestResponseNotification(fromUserID, toUserID int64, status string, contactID string) error {
	metadata := models.EventMetadata{
		ContactId: contactID,
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("error marshalling metadata: %w", err)
	}

	event := models.Event{
		EventType:      models.EventTypeRequestResponse,
		EventTitle:     "Solicitud de amistad respondida",
		Description:    fmt.Sprintf("Tu solicitud de amistad ha sido %s", status),
		UserId:         toUserID,
		OtherUserId:    sql.NullInt64{Int64: fromUserID, Valid: true},
		Status:         status,
		ActionRequired: false,
		ActionTakenAt:  sql.NullTime{Time: time.Now(), Valid: true},
		Metadata:       metadataJSON,
	}

	return createAndSendNotification(event)
}

// CreateSystemNotification crea una notificación del sistema
func CreateSystemNotification(userID int64, title, message string, systemEventType string, additionalData any) error {
	metadata := models.EventMetadata{
		SystemEventType: systemEventType,
		AdditionalData:  additionalData,
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("error marshalling metadata: %w", err)
	}

	event := models.Event{
		EventType:      models.EventTypeSystem,
		EventTitle:     title,
		Description:    message,
		UserId:         userID,
		Status:         models.EventStatusPending,
		ActionRequired: false,
		Metadata:       metadataJSON,
	}

	return createAndSendNotification(event)
}

// createAndSendNotification es una función interna que crea y envía una notificación
func createAndSendNotification(event models.Event) error {
	if notificationDB == nil {
		return fmt.Errorf("NotificationService no inicializado")
	}

	// Asegurarse de que CreateAt esté establecido si la BD no lo hace por defecto
	if event.CreateAt.IsZero() {
		event.CreateAt = time.Now()
	}

	if err := queries.CreateEvent(notificationDB, &event); err != nil {
		logger.Errorf("SERVICE_NOTIFICATION", "Error creando evento: %v", err)
		return fmt.Errorf("error creando evento: %w", err)
	}

	// notificationForClient, err := mapEventToNotificationInfo(event) // Comentado temporalmente por linter
	// if err != nil { // Comentado temporalmente por linter
	// 	logger.Errorf("SERVICE_NOTIFICATION", "Error mapeando evento a NotificationInfo para evento ID %d: %v", event.Id, err) // Comentado temporalmente por linter
	// 	// No retornar error aquí, el evento ya fue creado. Solo loguear. // Comentado temporalmente por linter
	// } // Comentado temporalmente por linter

	// Verificar si el usuario está online usando el servicio de presencia
	// Esta parte necesitaría acceso al ConnectionManager, que no está disponible globalmente aquí.
	// Por ahora, asumiremos que el manager se pasa o se accede de otra forma si es necesario.
	// Para el flujo actual de GetNotifications, no se envía en tiempo real, solo se recupera.

	// Aquí se podría añadir la lógica para enviar la notificación en tiempo real
	// si el usuario está conectado, similar a como se hacía antes pero
	// obteniendo la conexión del manager. Por ejemplo:
	// if manager != nil {
	//    if conn, exists := manager.GetConnection(event.UserId); exists {
	//        // ... enviar mensaje ...
	//    }
	// }

	logger.Infof("SERVICE_NOTIFICATION", "Evento (ID: %d) creado para UserID %d.", event.Id, event.UserId)
	return nil
}

// mapEventToNotificationInfo convierte un models.Event a wsmodels.NotificationInfo
func mapEventToNotificationInfo(event models.Event) (wsmodels.NotificationInfo, error) {
	wsPayload := make(map[string]interface{})
	if event.OtherUserId.Valid {
		wsPayload["otherUserId"] = event.OtherUserId.Int64
	}
	if event.ProyectId.Valid {
		wsPayload["projectId"] = event.ProyectId.Int64
	}
	if event.GroupId.Valid {
		wsPayload["groupId"] = event.GroupId.Int64
	}

	// Agregar metadata al payload si no es nula
	if len(event.Metadata) > 0 && string(event.Metadata) != "null" {
		var metadataMap map[string]interface{} // Usar un mapa genérico para metadata
		if err := json.Unmarshal(event.Metadata, &metadataMap); err == nil {
			for k, v := range metadataMap {
				wsPayload[k] = v
			}
		} else {
			logger.Warnf("SERVICE_NOTIFICATION", "Error unmarshalling metadata para evento ID %d: %v. Metadata: %s", event.Id, err, string(event.Metadata))
			// No devolver error, solo loguear. El resto de la notificación es válida.
		}
	}

	notificationInfo := wsmodels.NotificationInfo{
		ID:             fmt.Sprintf("%d", event.Id),
		Type:           event.EventType,
		Title:          event.EventTitle,
		Message:        event.Description,
		Timestamp:      event.CreateAt,
		IsRead:         event.IsRead,
		Payload:        wsPayload, // Contiene metadata y otros IDs
		ActionRequired: event.ActionRequired,
		Status:         event.Status,
		OtherUserId:    event.OtherUserId.Int64,
		ProyectId:      event.ProyectId.Int64,
		GroupId:        event.GroupId.Int64,
	}
	if event.ActionTakenAt.Valid {
		notificationInfo.ActionTakenAt = &event.ActionTakenAt.Time
	}

	// Obtener información del perfil si hay OtherUserId
	if event.OtherUserId.Valid {
		otherUserInfo, err := queries.GetUserBaseInfo(notificationDB, event.OtherUserId.Int64)
		if err == nil && otherUserInfo != nil {
			notificationInfo.Profile = wsmodels.ProfileData{
				ID:        otherUserInfo.ID,
				FirstName: otherUserInfo.FirstName,
				LastName:  otherUserInfo.LastName,
				UserName:  otherUserInfo.UserName,
				Picture:   otherUserInfo.Picture,
			}
		} else if err != nil {
			logger.Warnf("SERVICE_NOTIFICATION", "Error obteniendo UserBaseInfo para OtherUserId %d (Evento ID %d): %v", event.OtherUserId.Int64, event.Id, err)
		}
	}
	return notificationInfo, nil
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

// GetNotifications recupera las notificaciones para un usuario.
func GetNotifications(userID int64, onlyUnread bool, limit int, offset int) ([]wsmodels.NotificationInfo, error) {
	if notificationDB == nil {
		return nil, fmt.Errorf("NotificationService no inicializado")
	}
	logger.Infof("SERVICE_NOTIFICATION", "Obteniendo notificaciones para UserID %d (onlyUnread: %t, limit: %d, offset: %d)", userID, onlyUnread, limit, offset)

	// Llamar a la query que recupera los eventos
	events, err := queries.GetEventsByUserID(notificationDB, userID, onlyUnread, limit, offset)
	if err != nil {
		logger.Errorf("SERVICE_NOTIFICATION", "Error obteniendo eventos para UserID %d desde la BD: %v", userID, err)
		return nil, fmt.Errorf("error obteniendo eventos: %w", err)
	}

	notificationsInfo := make([]wsmodels.NotificationInfo, 0, len(events))
	for _, event := range events {
		notificationForClient, errMap := mapEventToNotificationInfo(event)
		if errMap != nil {
			// Loguear el error pero continuar, para no fallar toda la lista por una notificación
			logger.Warnf("SERVICE_NOTIFICATION", "Error mapeando evento ID %d para UserID %d: %v", event.Id, userID, errMap)
			continue
		}
		notificationsInfo = append(notificationsInfo, notificationForClient)
	}

	logger.Successf("SERVICE_NOTIFICATION", "%d notificaciones recuperadas para UserID %d", len(notificationsInfo), userID)
	return notificationsInfo, nil
}

// MarkRead marca una notificación específica como leída.
func MarkRead(userID int64, notificationIDStr string) error {
	if notificationDB == nil {
		return fmt.Errorf("NotificationService no inicializado")
	}

	eventID, err := strconv.ParseInt(notificationIDStr, 10, 64)
	if err != nil {
		logger.Errorf("SERVICE_NOTIFICATION", "Error convirtiendo notificationID '%s' a int64: %v", notificationIDStr, err)
		return fmt.Errorf("notificationID inválido: %w", err)
	}

	logger.Infof("SERVICE_NOTIFICATION", "Marcando notificación ID %d como leída para UserID %d", eventID, userID)

	// Llamar a la query para actualizar el estado IsRead del evento
	err = queries.MarkEventAsRead(notificationDB, eventID)
	if err != nil {
		logger.Errorf("SERVICE_NOTIFICATION", "Error marcando evento ID %d como leído para UserID %d en BD: %v", eventID, userID, err)
		return fmt.Errorf("error marcando como leído: %w", err)
	}

	logger.Successf("SERVICE_NOTIFICATION", "Notificación ID %d marcada como leída para UserID %d", eventID, userID)
	return nil
}

// MarkAllRead marca todas las notificaciones de un usuario como leídas.
func MarkAllRead(userID int64) (int64, error) {
	if notificationDB == nil {
		return 0, fmt.Errorf("NotificationService no inicializado")
	}
	logger.Infof("SERVICE_NOTIFICATION", "Marcando todas las notificaciones como leídas para UserID %d", userID)

	// Llamar a la query para actualizar el estado IsRead de todos los eventos del usuario
	rowsAffected, err := queries.MarkAllEventsAsReadForUser(notificationDB, userID)
	if err != nil {
		logger.Errorf("SERVICE_NOTIFICATION", "Error marcando todos los eventos como leídos para UserID %d en BD: %v", userID, err)
		return 0, fmt.Errorf("error marcando todos como leídos: %w", err)
	}

	logger.Successf("SERVICE_NOTIFICATION", "%d notificaciones marcadas como leídas para UserID %d", rowsAffected, userID)
	return rowsAffected, nil
}
