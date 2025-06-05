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

// HandleGetChatList maneja la solicitud del cliente para obtener su lista de chats.
func HandleGetChatList(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_CHAT", "User %d solicitó lista de chats. PID: %s", conn.ID, msg.PID)

	chatList, err := services.GetChatListForUser(conn.ID, conn.Manager())
	if err != nil {
		logger.Errorf("HANDLER_CHAT", "Error obteniendo chat list para user %d: %v", conn.ID, err)
		errMsg := types.ServerToClientMessage{
			PID:  msg.PID,
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
		return err
	}

	responseMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       types.MessageTypeChatList,
		FromUserID: conn.ID,
		Payload:    chatList,
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("HANDLER_CHAT", "Error enviando chat list a user %d: %v", conn.ID, err)
		return err
	}

	if msg.PID != "" {
		ackPayload := types.AckPayload{
			AcknowledgedPID: msg.PID,
			Status:          "chat_list_sent",
		}
		ackMsg := types.ServerToClientMessage{
			PID:        conn.Manager().Callbacks().GeneratePID(),
			Type:       types.MessageTypeServerAck,
			FromUserID: conn.ID,
			Payload:    ackPayload,
		}
		if err := conn.SendMessage(ackMsg); err != nil {
			logger.Warnf("HANDLER_CHAT", "Error enviando ServerAck para GetChatList a UserID %d para PID %s: %v", conn.ID, msg.PID, err)
		}
	}

	logger.Successf("HANDLER_CHAT", "Chat list enviada a user %d. PID respuesta: %s", conn.ID, responseMsg.PID)
	return nil
}

// HandleGetChatHistory maneja la solicitud del cliente para obtener el historial de mensajes de un chat.
func HandleGetChatHistory(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_CHAT", "User %d solicitó historial de chat. PID: %s", conn.ID, msg.PID)

	type GetChatHistoryPayload struct {
		ChatID          string `json:"chatId"`
		Limit           int    `json:"limit,omitempty"`
		BeforeMessageID string `json:"beforeMessageId,omitempty"`
	}

	var historyPayload GetChatHistoryPayload
	// msg.Payload should now directly contain the data for GetChatHistoryPayload
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error procesando payload de get_history (marshal): "+err.Error())
		return fmt.Errorf("error marshalling get_history payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &historyPayload); err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload de get_history (unmarshal): "+err.Error())
		return fmt.Errorf("error unmarshalling get_history payload: %w", err)
	}

	if historyPayload.ChatID == "" {
		conn.SendErrorNotification(msg.PID, 400, "ChatID es requerido para obtener el historial.")
		return errors.New("chatID no especificado en get_history")
	}

	if historyPayload.Limit <= 0 {
		historyPayload.Limit = 50 // Default limit
	}

	messages, err := services.GetChatHistory(historyPayload.ChatID, conn.ID, historyPayload.Limit, historyPayload.BeforeMessageID, conn.Manager())
	if err != nil {
		logger.Errorf("HANDLER_CHAT", "Error obteniendo historial para chat %s, user %d: %v", historyPayload.ChatID, conn.ID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error al obtener el historial del chat: "+err.Error())
		return err
	}

	responseMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       types.MessageTypeChatHistory,
		FromUserID: conn.ID,
		Payload:    messages,
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("HANDLER_CHAT", "Error enviando historial de chat %s a user %d: %v", historyPayload.ChatID, conn.ID, err)
		return err
	}

	if msg.PID != "" {
		ackPayload := types.AckPayload{
			AcknowledgedPID: msg.PID,
			Status:          "chat_history_sent",
		}
		ackMsg := types.ServerToClientMessage{
			PID:        conn.Manager().Callbacks().GeneratePID(),
			Type:       types.MessageTypeServerAck,
			FromUserID: conn.ID,
			Payload:    ackPayload,
		}
		if err := conn.SendMessage(ackMsg); err != nil {
			logger.Warnf("HANDLER_CHAT", "Error enviando ServerAck para GetChatHistory a UserID %d para PID %s: %v", conn.ID, msg.PID, err)
		}
	}

	logger.Successf("HANDLER_CHAT", "Historial de chat %s enviado a user %d. PID respuesta: %s", historyPayload.ChatID, conn.ID, responseMsg.PID)
	return nil
}
