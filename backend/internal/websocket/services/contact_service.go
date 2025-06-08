package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/google/uuid"
)

// AcceptFriendRequest procesa la aceptación de una solicitud de amistad.
// Actualiza el estado del contacto a 'accepted' y crea un chat entre los usuarios.
func AcceptFriendRequest(userID int64, notificationId string, timestamp string, manager *customws.ConnectionManager[wsmodels.WsUserData]) error {
	logger.Infof("SERVICE_CONTACT", "Procesando aceptación de solicitud de amistad para user %d", userID)

	// Obtener la notificación para validar y obtener el ID del otro usuario
	notification, err := queries.GetNotificationById(notificationId)
	if err != nil {
		return fmt.Errorf("error obteniendo notificación: %w", err)
	}

	if notification == nil {
		return fmt.Errorf("notificación no encontrada: %s", notificationId)
	}

	// Validar que la notificación sea para este usuario
	if notification.UserId != userID {
		return fmt.Errorf("notificación no pertenece al usuario %d", userID)
	}

	// Obtener el ID del otro usuario del payload
	otherUserId := notification.OtherUserId
	if otherUserId == 0 {
		return fmt.Errorf("ID del otro usuario no encontrado en la notificación")
	}

	// Actualizar el estado del contacto
	err = queries.UpdateContactStatus(userID, otherUserId, "accepted", timestamp)
	if err != nil {
		return fmt.Errorf("error actualizando estado del contacto: %w", err)
	}

	// Crear chat entre los usuarios con UUID
	chatId := uuid.NewString()

	// Actualizar el chatId en el contacto
	err = queries.UpdateContactChatId(userID, otherUserId, chatId)
	if err != nil {
		return fmt.Errorf("error actualizando chatId del contacto: %w", err)
	}

	// Notificar al otro usuario
	notificationMsg := types.ServerToClientMessage{
		Type:       types.MessageTypeNewNotification,
		FromUserID: userID,
		Payload: map[string]interface{}{
			"type":      "friend_request_accepted",
			"title":     "Solicitud aceptada",
			"message":   "Tu solicitud de amistad ha sido aceptada",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	if err := manager.SendMessageToUser(otherUserId, notificationMsg); err != nil {
		logger.Warnf("SERVICE_CONTACT", "Error enviando notificación de aceptación a user %d: %v", otherUserId, err)
	}

	logger.Successf("SERVICE_CONTACT", "Solicitud de amistad aceptada exitosamente para user %d", userID)
	return nil
}

// RejectFriendRequest procesa el rechazo de una solicitud de amistad.
// Actualiza el estado del contacto a 'rejected'.
func RejectFriendRequest(userID int64, notificationId string, timestamp string, manager *customws.ConnectionManager[wsmodels.WsUserData]) error {
	logger.Infof("SERVICE_CONTACT", "Procesando rechazo de solicitud de amistad para user %d", userID)

	// Obtener la notificación para validar y obtener el ID del otro usuario
	notification, err := queries.GetNotificationById(notificationId)
	if err != nil {
		return fmt.Errorf("error obteniendo notificación: %w", err)
	}

	if notification == nil {
		return fmt.Errorf("notificación no encontrada: %s", notificationId)
	}

	// Validar que la notificación sea para este usuario
	if notification.UserId != userID {
		return fmt.Errorf("notificación no pertenece al usuario %d", userID)
	}

	// Obtener el ID del otro usuario del payload
	otherUserId := notification.OtherUserId
	if otherUserId == 0 {
		return fmt.Errorf("ID del otro usuario no encontrado en la notificación")
	}

	// Actualizar el estado del contacto
	err = queries.UpdateContactStatus(userID, otherUserId, "rejected", timestamp)
	if err != nil {
		return fmt.Errorf("error actualizando estado del contacto: %w", err)
	}

	// Notificar al otro usuario
	notificationMsg := types.ServerToClientMessage{
		Type:       types.MessageTypeNewNotification,
		FromUserID: userID,
		Payload: map[string]interface{}{
			"type":      "friend_request_rejected",
			"title":     "Solicitud rechazada",
			"message":   "Tu solicitud de amistad ha sido rechazada",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	if err := manager.SendMessageToUser(otherUserId, notificationMsg); err != nil {
		logger.Warnf("SERVICE_CONTACT", "Error enviando notificación de rechazo a user %d: %v", otherUserId, err)
	}

	logger.Successf("SERVICE_CONTACT", "Solicitud de amistad rechazada exitosamente para user %d", userID)
	return nil
}

// CreateContactRequest crea una nueva solicitud de contacto.
// Inserta un nuevo contacto con estado 'pending' y crea un chat asociado.
func CreateContactRequest(senderID, recipientID int64, manager *customws.ConnectionManager[wsmodels.WsUserData]) error {
	logger.Infof("SERVICE_CONTACT", "User %d iniciando contacto con user %d", senderID, recipientID)

	// Crear chatID con UUID
	chatID := uuid.NewString()

	// Crear el contacto con estado 'pending'
	err := queries.CreateContact(senderID, recipientID, chatID, "pending")
	if err != nil {
		return fmt.Errorf("error creando contacto: %w", err)
	}

	// Crear y guardar el evento
	event := &models.Event{
		EventType:      models.EventTypeFriendRequest,
		EventTitle:     "Nueva solicitud de contacto",
		Description:    "Has recibido una nueva solicitud de contacto.",
		UserId:         recipientID,
		OtherUserId:    sql.NullInt64{Int64: senderID, Valid: true},
		CreateAt:       time.Now(),
		IsRead:         false,
		Status:         models.EventStatusPending,
		ActionRequired: true,
	}

	if err := queries.CreateEvent(event); err != nil {
		// Aunque falle la creación del evento, no consideramos que sea un error fatal
		// para el flujo principal de creación de contacto. Solo lo logueamos.
		logger.Errorf("SERVICE_CONTACT", "Error creando evento de notificación para user %d: %v", recipientID, err)
	}

	// Notificar al usuario receptor
	notificationMsg := types.ServerToClientMessage{
		Type:       types.MessageTypeNewNotification,
		FromUserID: senderID,
		Payload: map[string]interface{}{
			"type":      "friend_request_received",
			"title":     "Nueva solicitud de amistad",
			"message":   "Has recibido una nueva solicitud de amistad",
			"senderId":  senderID,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	if err := manager.SendMessageToUser(recipientID, notificationMsg); err != nil {
		logger.Warnf("SERVICE_CONTACT", "Error enviando notificación de solicitud de amistad a user %d: %v", recipientID, err)
	}

	logger.Successf("SERVICE_CONTACT", "Solicitud de contacto de user %d a user %d enviada exitosamente", senderID, recipientID)
	return nil
}
