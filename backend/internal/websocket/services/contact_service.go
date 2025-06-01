package services

import (
	"fmt"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
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

	// Crear chat entre los usuarios
	chatId, err := queries.CreateChat(userID, otherUserId)
	if err != nil {
		return fmt.Errorf("error creando chat: %w", err)
	}

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
