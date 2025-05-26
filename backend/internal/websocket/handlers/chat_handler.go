package handlers

import (
	"encoding/json"
	"errors"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// HandleGetChatList maneja la solicitud del cliente para obtener su lista de chats.
func HandleGetChatList(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_CHAT", "User %d solicitó lista de chats. PID: %s", conn.ID, msg.PID)

	chatList, err := services.GetChatListForUser(conn.ID, conn.Manager())
	if err != nil {
		logger.Errorf("HANDLER_CHAT", "Error obteniendo chat list para user %d: %v", conn.ID, err)
		// Enviar notificación de error al cliente
		errMsg := types.ServerToClientMessage{
			PID:  msg.PID, // Usar el PID del cliente si está disponible, o generar uno nuevo
			Type: types.MessageTypeErrorNotification,
			Error: &types.ErrorPayload{
				OriginalPID: msg.PID,
				Code:        500,
				Message:     "Error al obtener la lista de chats: " + err.Error(),
			},
		}
		if sendErr := conn.SendMessage(errMsg); sendErr != nil {
			logger.Errorf("HANDLER_CHAT", "Error enviando notificación de error de GetChatList a user %d: %v", conn.ID, sendErr)
		}
		return err // Devolver el error para que el callback principal lo maneje si es necesario
	}

	responseMsg := types.ServerToClientMessage{
		PID:     conn.Manager().Callbacks().GeneratePID(), // Generar nuevo PID para la respuesta del servidor
		Type:    types.MessageTypeChatList,
		Payload: chatList,
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("HANDLER_CHAT", "Error enviando chat list a user %d: %v", conn.ID, err)
		return err
	}

	logger.Successf("HANDLER_CHAT", "Chat list enviada a user %d. PID respuesta: %s", conn.ID, responseMsg.PID)
	return nil
}

// HandleSendChatMessage maneja un mensaje de chat entrante de un cliente.
func HandleSendChatMessage(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_CHAT", "Mensaje de chat recibido de UserID %d. PID: %s, TargetUserID: %d", conn.ID, msg.PID, msg.TargetUserID)

	// 1. Decodificar el payload del mensaje
	// Asumimos una estructura para el payload del mensaje de chat, por ejemplo:
	type ChatMessagePayload struct {
		Text         string `json:"text"`
		TargetUserID int64  `json:"targetUserId"` // Redundante con msg.TargetUserID, pero puede estar en payload
		// ClientTempID string `json:"clientTempId,omitempty"` // ID temporal del cliente para correlación
	}

	var chatPayload ChatMessagePayload
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload (marshal): "+err.Error())
		return errors.New("error marshalling chat payload: " + err.Error())
	}
	if err := json.Unmarshal(payloadBytes, &chatPayload); err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload (unmarshal): "+err.Error())
		return errors.New("error unmarshalling chat payload: " + err.Error())
	}

	if chatPayload.Text == "" {
		conn.SendErrorNotification(msg.PID, 400, "El texto del mensaje no puede estar vacío.")
		return errors.New("texto del mensaje vacío")
	}

	// Asegurar que TargetUserID está presente (ya sea en msg o en payload)
	targetUserID := msg.TargetUserID
	if targetUserID == 0 {
		targetUserID = chatPayload.TargetUserID
	}
	if targetUserID == 0 {
		conn.SendErrorNotification(msg.PID, 400, "TargetUserID es requerido para enviar un mensaje de chat.")
		return errors.New("targetUserID no especificado")
	}

	if targetUserID == conn.ID {
		conn.SendErrorNotification(msg.PID, 400, "No puedes enviarte mensajes a ti mismo de esta manera.")
		return errors.New("intento de auto-mensaje")
	}

	// 2. Llamar al servicio para procesar y guardar el mensaje
	savedMessage, err := services.ProcessAndSaveChatMessage(
		conn.ID,      // fromUserID
		targetUserID, // toUserID
		chatPayload.Text,
		conn.Manager(),
	)

	if err != nil {
		logger.Errorf("HANDLER_CHAT", "Error procesando/guardando mensaje de chat de user %d para %d: %v", conn.ID, targetUserID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al procesar tu mensaje: "+err.Error())
		return err
	}

	// 3. Enviar ACK al remitente (cliente que envió el mensaje)
	// El ACK confirma que el servidor recibió y procesó (guardó) el mensaje.
	// El PID original del cliente (msg.PID) se usa para el AcknowledgedPID.
	if msg.PID != "" { // Solo enviar ServerAck si el cliente envió un PID
		ackPayload := types.AckPayload{
			AcknowledgedPID: msg.PID,
			Status:          "processed_and_saved",
			// Podríamos añadir el ID del mensaje guardado aquí si el cliente lo necesita
			// "messageServerId": savedMessage.ID,
		}
		ackMsg := types.ServerToClientMessage{
			PID:     conn.Manager().Callbacks().GeneratePID(),
			Type:    types.MessageTypeServerAck,
			Payload: ackPayload,
		}
		if err := conn.SendMessage(ackMsg); err != nil {
			logger.Warnf("HANDLER_CHAT", "Error enviando ServerAck por chat msg a UserID %d para PID %s: %v", conn.ID, msg.PID, err)
		}
	}

	// 4. El servicio ProcessAndSaveChatMessage ya se encargó de enviar el mensaje al destinatario.

	logger.Successf("HANDLER_CHAT", "Mensaje de chat de UserID %d a %d procesado. MsgDB ID: %s", conn.ID, targetUserID, savedMessage.ID)
	return nil
}

// TODO: Implementar HandleMessagesRead, HandleTypingIndicatorOn, HandleTypingIndicatorOff
