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

func HandleContactRequest(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_CONTACT", "User %d iniciando contacto. PID: %s", conn.ID, msg.PID)

	type ContactRequestPayload struct {
		RecipientID int64 `json:"recipientId"`
	}

	var payload ContactRequestPayload
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error procesando payload de contact_request: "+err.Error())
		return fmt.Errorf("error marshalling contact_request payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload de contact_request: "+err.Error())
		return fmt.Errorf("error unmarshalling contact_request payload: %w", err)
	}

	if payload.RecipientID == 0 {
		conn.SendErrorNotification(msg.PID, 400, "RecipientID es requerido para iniciar contacto.")
		return errors.New("recipientID no especificado en contact_request")
	}

	err = services.CreateContactRequest(conn.ID, payload.RecipientID, conn.Manager())
	if err != nil {
		logger.Errorf("HANDLER_CONTACT", "Error iniciando contacto para user %d: %v", conn.ID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error al iniciar contacto: "+err.Error())
		return err
	}
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

	logger.Successf("HANDLER_CONTACT", "Solicitud de contacto enviada para user %d. PID respuesta: %s", conn.ID, ackMsg.PID)
	return nil
}
