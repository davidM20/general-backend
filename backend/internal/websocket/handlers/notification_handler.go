// Package handlers contiene los manejadores de WebSocket para las notificaciones del sistema.
// Este paquete implementa la lógica de negocio para gestionar las notificaciones en tiempo real
// a través de conexiones WebSocket.
//
// Casos de uso principales:
//
// 1. HandleGetNotifications:
//    - Obtiene la lista de notificaciones para un usuario específico
//    - Soporta paginación mediante limit y offset
//    - Permite filtrar por notificaciones no leídas (onlyUnread)
//    - Si se solicitan notificaciones no leídas y hay menos de 15, complementa con leídas
//    - Caso de uso ideal: Cuando el cliente necesita mostrar el feed de notificaciones
//      o actualizar la bandeja de entrada de notificaciones
//
// 2. HandleMarkNotificationRead:
//    - Marca una notificación específica como leída
//    - Requiere el ID de la notificación
//    - Caso de uso ideal: Cuando el usuario abre/interactúa con una notificación
//      y se necesita actualizar su estado de lectura
//
// 3. HandleMarkAllNotificationsRead:
//    - Marca todas las notificaciones del usuario como leídas
//    - Retorna el número de notificaciones actualizadas
//    - Caso de uso ideal: Cuando el usuario quiere marcar todas sus notificaciones
//      como leídas de una vez, por ejemplo, al hacer clic en "Marcar todas como leídas"
//
// Notas importantes:
// - Todas las funciones manejan errores y envían respuestas apropiadas al cliente
// - Se mantiene un registro detallado de las operaciones mediante logs
// - Las respuestas incluyen ACKs para confirmar la recepción de mensajes
// - Se implementa paginación para manejar grandes volúmenes de notificaciones
// - Se complementan las notificaciones no leídas con leídas para mantener un mínimo
//   de 15 resultados cuando sea necesario

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
	logger.Infof("HANDLER_NOTIFICATION", "Usuario %d solicitó lista de notificaciones. PID: %s, Payload: %+v", conn.ID, msg.PID, msg.Payload)

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
	// payload.Limit es el que se usará para las llamadas a servicios.
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

	const minResultsDesired = 15
	currentLimitFromPayload := payload.Limit // Límite efectivo usado para la primera llamada y como máximo final.

	// Lógica para complementar si se pidieron 'onlyUnread' y se obtuvieron pocas.
	if payload.OnlyUnread && len(notifications) < minResultsDesired && len(notifications) < currentLimitFromPayload {
		logger.Infof("HANDLER_NOTIFICATION", "Usuario %d: (onlyUnread=true) Obtuvo %d notificaciones (de %d solicitadas). Menos de %d. Intentando complementar.", conn.ID, len(notifications), currentLimitFromPayload, minResultsDesired)

		allPotentiallyRelevantNotifications, errAll := services.GetNotifications(conn.ID, false, currentLimitFromPayload, payload.Offset)
		if errAll != nil {
			logger.Warnf("HANDLER_NOTIFICATION", "Usuario %d: Error obteniendo notificaciones (incluyendo leídas) para complementar: %v. Se continuará con las %d no leídas obtenidas.", conn.ID, errAll, len(notifications))
		} else {
			logger.Debugf("HANDLER_NOTIFICATION", "Usuario %d: Obtenidas %d notificaciones (incl. leídas) para complementar. No leídas originales: %d.", conn.ID, len(allPotentiallyRelevantNotifications), len(notifications))

			finalNotifications := []wsmodels.NotificationInfo{}
			existingIDs := make(map[string]bool)

			for _, n := range notifications {
				finalNotifications = append(finalNotifications, n)
				existingIDs[n.ID] = true
			}

			targetTotalCount := minResultsDesired
			if currentLimitFromPayload < minResultsDesired {
				targetTotalCount = currentLimitFromPayload
			}
			// Si ya teníamos más no leídas que minResultsDesired (pero menos que currentLimitFromPayload),
			// el targetTotalCount debería ser len(finalNotifications) para no truncar,
			// pero esto ya está cubierto porque el loop de abajo no añadirá más si ya se alcanzó targetTotalCount.
			// La condición del if principal (len(notifications) < minResultsDesired) previene este caso de todos modos.

			// Asegurar no exceder el límite original del payload y no ser menor a las ya obtenidas si estas superan el minResultsDesired
			if targetTotalCount < len(finalNotifications) {
				targetTotalCount = len(finalNotifications)
			}
			if targetTotalCount > currentLimitFromPayload {
				targetTotalCount = currentLimitFromPayload
			}

			for _, an := range allPotentiallyRelevantNotifications {
				if len(finalNotifications) >= targetTotalCount {
					break
				}
				if _, exists := existingIDs[an.ID]; !exists {
					finalNotifications = append(finalNotifications, an)
					existingIDs[an.ID] = true
				}
			}
			notifications = finalNotifications
			logger.Infof("HANDLER_NOTIFICATION", "Usuario %d: Lista complementada a %d notificaciones (objetivo era %d, límite original %d).", conn.ID, len(notifications), targetTotalCount, currentLimitFromPayload)
		}
	}

	// DEBUG: Loguear las notificaciones recuperadas ANTES de enviarlas
	for i, notif := range notifications {
		// Suponiendo que wsmodels.Notification tiene un campo 'Read bool'. Ajustar si es necesario.
		logger.Debugf("HANDLER_NOTIFICATION", "Notificación %d para UserID %d (ANTES DE ENVIAR): ID=%s, Read=%t, Type=%s, Title=%s, ProfileID=%d, ProfileName=%s, ProfilePic=%s, Payload=%+v",
			i, conn.ID, notif.ID, notif.IsRead, notif.Type, notif.Title, notif.Profile.ID, notif.Profile.FirstName+" "+notif.Profile.LastName, notif.Profile.Picture, notif.Payload)
	}

	responseMsg := types.ServerToClientMessage{
		PID:     conn.Manager().Callbacks().GeneratePID(),
		Type:    types.MessageTypeNotificationList,
		Payload: notifications,
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("HANDLER_NOTIFICATION", "Error enviando lista de notificaciones a user %d: %v", conn.ID, err)
		return err
	}

	// Enviar ACK al cliente confirmando que el request fue procesado exitosamente
	if msg.PID != "" {
		ackPayload := types.AckPayload{
			AcknowledgedPID: msg.PID,
			Status:          "notification_list_sent",
		}
		ackMsg := types.ServerToClientMessage{
			PID:     conn.Manager().Callbacks().GeneratePID(),
			Type:    types.MessageTypeServerAck,
			Payload: ackPayload,
		}
		if err := conn.SendMessage(ackMsg); err != nil {
			logger.Warnf("HANDLER_NOTIFICATION", "Error enviando ServerAck para GetNotifications a UserID %d para PID %s: %v", conn.ID, msg.PID, err)
		}
	}

	logger.Successf("HANDLER_NOTIFICATION", "Lista de notificaciones enviada a user %d. Total: %d. PID respuesta: %s", conn.ID, len(notifications), responseMsg.PID)
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
		Type:    types.MessageTypeServerAck,
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

// HandleMarkAllNotificationsRead maneja la solicitud para marcar todas las notificaciones como leídas.
func HandleMarkAllNotificationsRead(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_NOTIFICATION", "Usuario %d solicitó marcar todas las notificaciones como leídas. PID: %s", conn.ID, msg.PID)

	rowsAffected, err := services.MarkAllRead(conn.ID)
	if err != nil {
		logger.Errorf("HANDLER_NOTIFICATION", "Error marcando todas las notificaciones como leídas para user %d: %v", conn.ID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error al marcar todas las notificaciones como leídas: "+err.Error())
		return err
	}

	// Enviar confirmación con el número de notificaciones marcadas como leídas
	ackPayload := types.AckPayload{
		AcknowledgedPID: msg.PID,
		Status:          "all_notifications_marked_as_read",
	}
	serverAckMsg := types.ServerToClientMessage{
		PID:     conn.Manager().Callbacks().GeneratePID(),
		Type:    types.MessageTypeServerAck,
		Payload: ackPayload,
	}

	// Enviar ACK primero
	if msg.PID != "" {
		if err := conn.SendMessage(serverAckMsg); err != nil {
			logger.Warnf("HANDLER_NOTIFICATION", "Error enviando ServerAck para MarkAllNotificationsRead a UserID %d: %v", conn.ID, err)
		}
	}

	// Enviar mensaje con el número de notificaciones marcadas
	responseMsg := types.ServerToClientMessage{
		PID:  conn.Manager().Callbacks().GeneratePID(),
		Type: types.MessageTypeDataEvent,
		Payload: map[string]interface{}{
			"notificationsMarked": rowsAffected,
		},
	}
	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Warnf("HANDLER_NOTIFICATION", "Error enviando número de notificaciones marcadas a UserID %d: %v", conn.ID, err)
	}

	logger.Successf("HANDLER_NOTIFICATION", "%d notificaciones marcadas como leídas para user %d. PID original: %s", rowsAffected, conn.ID, msg.PID)
	return nil
}
