package handlers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// HandleGetNotifications maneja la solicitud para obtener la lista de notificaciones.
func HandleGetNotifications(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_NOTIFICATION", "Usuario %d solicitó lista de notificaciones. PID: %s", conn.ID, msg.PID)

	// Decodificar payload si es necesario para parámetros (onlyUnread, limit, offset)
	type GetNotificationsPayload struct {
		OnlyUnread bool `json:"onlyUnread,omitempty"`
		Limit      int  `json:"limit,omitempty"`
		Offset     int  `json:"offset,omitempty"`
	}
	var payload GetNotificationsPayload
	if msg.Payload != nil {
		payloadBytes, err := json.Marshal(msg.Payload)
		if err != nil {
			conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload (marshal): "+err.Error())
			return errors.New("error marshalling GetNotifications payload: " + err.Error())
		}
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload (unmarshal): "+err.Error())
			return errors.New("error unmarshalling GetNotifications payload: " + err.Error())
		}
	}

	// Establecer valores por defecto si no se proporcionan o son inválidos
	if payload.Limit <= 0 {
		payload.Limit = 20 // Default limit
	}
	if payload.Offset < 0 {
		payload.Offset = 0 // Default offset
	}

	notifications, err := services.GetNotifications(conn.ID, payload.OnlyUnread, payload.Limit, payload.Offset)
	if err != nil {
		logger.Errorf("HANDLER_NOTIFICATION", "Error obteniendo notificaciones para user %d: %v", conn.ID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error al obtener tus notificaciones: "+err.Error())
		return err
	}

	responseMsg := types.ServerToClientMessage{
		PID:     conn.Manager().Callbacks().GeneratePID(), // Usar el PID del cliente si se espera respuesta al mismo PID, o generar nuevo
		Type:    types.MessageTypeNotificationList,
		Payload: notifications,
	}
	// Si el cliente envió un PID en la solicitud, es buena práctica usarlo en la respuesta si es una respuesta directa.
	// Sin embargo, para listas o eventos, a veces se generan PIDs nuevos desde el servidor.
	// Por consistencia con otros handlers como GetChatList, generamos uno nuevo.

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("HANDLER_NOTIFICATION", "Error enviando lista de notificaciones a user %d: %v", conn.ID, err)
		return err
	}

	logger.Successf("HANDLER_NOTIFICATION", "Lista de notificaciones enviada a user %d. PID respuesta: %s", conn.ID, responseMsg.PID)
	return nil
}

// HandleMarkNotificationRead maneja la solicitud para marcar una notificación como leída.
func HandleMarkNotificationRead(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_NOTIFICATION", "Usuario %d solicitó marcar notificación como leída. PID: %s", conn.ID, msg.PID)

	type MarkReadPayload struct {
		NotificationID string `json:"notificationId"`
	}
	var payload MarkReadPayload
	if msg.Payload == nil {
		conn.SendErrorNotification(msg.PID, 400, "Payload es requerido para marcar notificación como leída.")
		return errors.New("payload vacío para MarkNotificationRead")
	}

	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload (marshal): "+err.Error())
		return fmt.Errorf("error marshalling MarkNotificationRead payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload (unmarshal): "+err.Error())
		return fmt.Errorf("error unmarshalling MarkNotificationRead payload: %w", err)
	}

	if payload.NotificationID == "" {
		conn.SendErrorNotification(msg.PID, 400, "notificationId es requerido.")
		return errors.New("notificationId vacío en MarkNotificationRead")
	}

	if err := services.MarkRead(conn.ID, payload.NotificationID); err != nil {
		logger.Errorf("HANDLER_NOTIFICATION", "Error marcando notificación %s como leída para user %d: %v", payload.NotificationID, conn.ID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error al marcar notificación como leída: "+err.Error())
		return err
	}

	// Enviar un ServerAck o una confirmación específica MessageTypeNotificationRead
	ackPayload := types.AckPayload{
		AcknowledgedPID: msg.PID,
		Status:          "notification_marked_as_read",
	}
	serverAckMsg := types.ServerToClientMessage{
		PID:     conn.Manager().Callbacks().GeneratePID(),
		Type:    types.MessageTypeServerAck, // O MessageTypeNotificationRead si se quiere un tipo más específico
		Payload: ackPayload,
	}

	if msg.PID != "" { // Solo enviar ack si el cliente envió un PID
		if err := conn.SendMessage(serverAckMsg); err != nil {
			logger.Warnf("HANDLER_NOTIFICATION", "Error enviando ServerAck/NotificationReadAck para UserID %d, NotifID %s: %v", conn.ID, payload.NotificationID, err)
		}
	}

	logger.Successf("HANDLER_NOTIFICATION", "Notificación %s marcada como leída para user %d. PID original: %s", payload.NotificationID, conn.ID, msg.PID)
	return nil
}

// HandleMarkAllNotificationsRead podría ser una funcionalidad adicional.
// Por ahora, nos enfocamos en marcar una individualmente.
// Si se implementa, sería similar a HandleMarkNotificationRead pero llamando a services.MarkAllRead.
