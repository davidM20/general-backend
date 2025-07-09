package handlers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// HandleAcceptFriendRequest maneja la solicitud del cliente para aceptar una solicitud de amistad.
func HandleAcceptFriendRequest(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_CONTACT", "User %d aceptando solicitud de amistad. PID: %s", conn.ID, msg.PID)

	type AcceptFriendRequestPayload struct {
		NotificationId string `json:"notificationId"`
		Timestamp      string `json:"timestamp"`
	}

	var payload AcceptFriendRequestPayload
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error procesando payload de accept_request: "+err.Error())
		return fmt.Errorf("error marshalling accept_request payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload de accept_request: "+err.Error())
		return fmt.Errorf("error unmarshalling accept_request payload: %w", err)
	}

	if payload.NotificationId == "" {
		conn.SendErrorNotification(msg.PID, 400, "NotificationId es requerido para aceptar la solicitud.")
		return errors.New("notificationId no especificado en accept_request")
	}

	err = services.AcceptFriendRequest(conn.ID, payload.NotificationId, payload.Timestamp, conn.Manager())
	if err != nil {
		logger.Errorf("HANDLER_CONTACT", "Error aceptando solicitud de amistad para user %d: %v", conn.ID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error al aceptar la solicitud de amistad: "+err.Error())
		return err
	}

	ackPayload := types.AckPayload{
		AcknowledgedPID: msg.PID,
		Status:          "friend_request_accepted",
	}
	ackMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       types.MessageTypeServerAck,
		FromUserID: conn.ID,
		Payload:    ackPayload,
	}
	if err := conn.SendMessage(ackMsg); err != nil {
		logger.Warnf("HANDLER_CONTACT", "Error enviando ServerAck para AcceptFriendRequest a UserID %d para PID %s: %v", conn.ID, msg.PID, err)
	}

	logger.Successf("HANDLER_CONTACT", "Solicitud de amistad aceptada para user %d. PID respuesta: %s", conn.ID, ackMsg.PID)
	return nil
}

// HandleRejectFriendRequest maneja la solicitud del cliente para rechazar una solicitud de amistad.
func HandleRejectFriendRequest(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_CONTACT", "User %d rechazando solicitud de amistad. PID: %s", conn.ID, msg.PID)

	type RejectFriendRequestPayload struct {
		NotificationId string `json:"notificationId"`
		Timestamp      string `json:"timestamp"`
	}

	var payload RejectFriendRequestPayload
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error procesando payload de reject_request: "+err.Error())
		return fmt.Errorf("error marshalling reject_request payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload de reject_request: "+err.Error())
		return fmt.Errorf("error unmarshalling reject_request payload: %w", err)
	}

	if payload.NotificationId == "" {
		conn.SendErrorNotification(msg.PID, 400, "NotificationId es requerido para rechazar la solicitud.")
		return errors.New("notificationId no especificado en reject_request")
	}

	err = services.RejectFriendRequest(conn.ID, payload.NotificationId, payload.Timestamp, conn.Manager())
	if err != nil {
		logger.Errorf("HANDLER_CONTACT", "Error rechazando solicitud de amistad para user %d: %v", conn.ID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error al rechazar la solicitud de amistad: "+err.Error())
		return err
	}

	ackPayload := types.AckPayload{
		AcknowledgedPID: msg.PID,
		Status:          "friend_request_rejected",
	}
	ackMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       types.MessageTypeServerAck,
		FromUserID: conn.ID,
		Payload:    ackPayload,
	}
	if err := conn.SendMessage(ackMsg); err != nil {
		logger.Warnf("HANDLER_CONTACT", "Error enviando ServerAck para RejectFriendRequest a UserID %d para PID %s: %v", conn.ID, msg.PID, err)
	}

	logger.Successf("HANDLER_CONTACT", "Solicitud de amistad rechazada para user %d. PID respuesta: %s", conn.ID, ackMsg.PID)
	return nil
}

// HandleContactRequest maneja una nueva solicitud de contacto de un usuario a otro.
func HandleContactRequest(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	var payload struct {
		ToUserID       int64  `json:"toUserId"`
		RequestMessage string `json:"message"`
	}

	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error procesando payload: "+err.Error())
		return fmt.Errorf("error marshalling payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload: "+err.Error())
		return fmt.Errorf("error unmarshalling payload: %w", err)
	}

	fromUserID := conn.ID
	if fromUserID == payload.ToUserID {
		conn.SendErrorNotification(msg.PID, 400, "No puedes enviarte una solicitud de contacto a ti mismo.")
		return nil // Error ya notificado al cliente
	}

	// --- NUEVA VALIDACIÓN: Verificar que el usuario destino existe antes de crear la solicitud ---
	if _, err := queries.GetUserBaseInfo(payload.ToUserID); err != nil {
		// El usuario no existe o está inactivo
		notFoundMsg := fmt.Sprintf("El usuario destinatario (ID: %d) no existe.", payload.ToUserID)
		logger.Warnf("HANDLER_CONTACT", "Intento de crear contacto con usuario inexistente. From: %d, To: %d", fromUserID, payload.ToUserID)
		conn.SendErrorNotification(msg.PID, 404, notFoundMsg)
		return nil // Se notificó adecuadamente, no propagar error
	}

	// --- NUEVA VALIDACIÓN: Verificar si ya existe una solicitud de contacto entre los usuarios ---
	exists, err := queries.CheckContactExists(fromUserID, payload.ToUserID)
	if err != nil {
		logger.Errorf("HANDLER_CONTACT", "Error al verificar si el contacto ya existe entre %d y %d: %v", fromUserID, payload.ToUserID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al procesar la solicitud.")
		return err // Devolver err para que el middleware lo maneje si es necesario
	}
	if exists {
		logger.Warnf("HANDLER_CONTACT", "Intento de crear contacto duplicado entre %d y %d", fromUserID, payload.ToUserID)
		conn.SendErrorNotification(msg.PID, 409, "Ya existe una solicitud de contacto o amistad con este usuario.")
		return nil // Error controlado, no es necesario propagar.
	}

	chatID := uuid.New().String()

	err = queries.CreateContact(fromUserID, payload.ToUserID, chatID, "pending")
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			// Error por clave foránea (usuario no existe) o por duplicado
			switch mysqlErr.Number {
			case 1452:
				// FK constraint: usuario destino no existe (fallback de seguridad)
				errorMsg := fmt.Sprintf("El usuario destinatario (ID: %d) no existe.", payload.ToUserID)
				conn.SendErrorNotification(msg.PID, 404, errorMsg)
				return nil // Error controlado
			case 1062:
				// Duplicated entry (ya existe una solicitud de contacto)
				conn.SendErrorNotification(msg.PID, 409, "Ya existe una solicitud de contacto pendiente o aceptada con este usuario.")
				return nil // Error controlado
			}
		}
		logger.Errorf("HANDLER_CONTACT", "Error creando contacto en la BD para UserID %d hacia %d: %v", fromUserID, payload.ToUserID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al crear la solicitud.")
		return nil // Error ya notificado, no propagar para evitar doble notificación
	}

	err = services.CreateFriendRequestNotification(fromUserID, payload.ToUserID, payload.RequestMessage)
	if err != nil {
		logger.Errorf("HANDLER_CONTACT", "Error creando notificación para UserID %d: %v", payload.ToUserID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al crear la notificación.")
		return err
	}

	// Enviar notificación en tiempo real si el destinatario está conectado.
	// El servicio se encargará de esto internamente
	go services.ProcessAndSendNotification(
		payload.ToUserID,
		"FRIEND_REQUEST",
		"Nueva solicitud de contacto",
		"Has recibido una nueva solicitud de contacto.",
		map[string]interface{}{"otherUserId": fromUserID, "requestMessage": payload.RequestMessage},
		conn.Manager(),
	)

	ackPayload := types.AckPayload{
		AcknowledgedPID: msg.PID,
		Status:          "contact_request_sent",
	}
	ackMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       types.MessageTypeServerAck,
		FromUserID: conn.ID,
		Payload:    ackPayload,
	}
	if err := conn.SendMessage(ackMsg); err != nil {
		logger.Warnf("HANDLER_CONTACT", "Error enviando ServerAck para ContactRequest a UserID %d para PID %s: %v", conn.ID, msg.PID, err)
	}

	logger.Successf("HANDLER_CONTACT", "Solicitud de contacto enviada de %d a %d", fromUserID, payload.ToUserID)

	return nil
}
