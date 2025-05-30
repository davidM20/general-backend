package handlers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// HandleDataRequest maneja solicitudes de datos genéricas y las redirecciona a los handlers específicos
func HandleDataRequest(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_DATA", "Data request recibida de UserID %d. PID: %s", conn.ID, msg.PID)

	// Define the expected structure of the "data_request" payload
	type DataRequestPayload struct {
		Action   string                 `json:"action"`
		Resource string                 `json:"resource,omitempty"` // Resource might be optional for actions like "ping"
		Data     map[string]interface{} `json:"data,omitempty"`     // Specific data for the action/resource
	}

	var requestData DataRequestPayload

	// The original msg.Payload is expected to be the DataRequestPayload
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error interno procesando payload (marshal): "+err.Error())
		return fmt.Errorf("error marshalling data_request payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload de data_request (unmarshal): "+err.Error())
		return fmt.Errorf("error unmarshalling data_request payload: %w", err)
	}

	logger.Debugf("HANDLER_DATA", "Procesando data_request: action='%s', resource='%s'", requestData.Action, requestData.Resource)

	// --- Action Dispatching ---

	// 1. Handle "ping" action (simple, no resource needed)
	if requestData.Action == "ping" {
		logger.Debugf("HANDLER_DATA", "Ping recibido de UserID %d", conn.ID)
		if msg.PID != "" {
			ackPayload := types.AckPayload{AcknowledgedPID: msg.PID, Status: "pong"}
			ackMsg := types.ServerToClientMessage{
				PID:        conn.Manager().Callbacks().GeneratePID(),
				Type:       types.MessageTypeServerAck,
				FromUserID: conn.ID,
				Payload:    ackPayload,
			}
			if err := conn.SendMessage(ackMsg); err != nil {
				logger.Warnf("HANDLER_DATA", "Error enviando pong (ServerAck) a UserID %d: %v", conn.ID, err)
			} else {
				logger.Debugf("HANDLER_DATA", "Pong enviado a UserID %d para PID %s", conn.ID, msg.PID)
			}
		}
		return nil
	}

	// 2. For other actions, a resource is typically required.
	if requestData.Resource == "" {
		errMsg := fmt.Sprintf("Recurso no especificado para la acción '%s'", requestData.Action)
		logger.Warn("HANDLER_DATA", errMsg)
		conn.SendErrorNotification(msg.PID, 400, errMsg)
		return errors.New(errMsg)
	}

	// Prepare a message for the sub-handler.
	// It carries the original PID (for ACKs) and the specific `requestData.Data` as its payload.
	// The Type can be preserved or updated if the sub-handler expects a specific one.
	subHandlerMessage := types.ClientToServerMessage{
		PID:     msg.PID,          // Preserve original PID for ACK traceability
		Type:    msg.Type,         // Preserve original type, or modify below if needed
		Payload: requestData.Data, // The specific data for the sub-handler
	}

	// 3. Dispatch based on Action and Resource combination
	switch requestData.Action {
	case "get_list":
		switch requestData.Resource {
		case "chat":
			return HandleGetChatList(conn, msg)
		case "notification":
			return HandleGetNotifications(conn, subHandlerMessage)
		default:
			return handleUnsupportedResource(conn, msg.PID, requestData.Action, requestData.Resource)
		}

	case "get_pending":
		switch requestData.Resource {
		case "notification":
			// For "get_pending" notifications, we inject "onlyUnread: true" into the data.
			pendingData := requestData.Data
			if pendingData == nil {
				pendingData = make(map[string]interface{})
			}
			pendingData["onlyUnread"] = true
			subHandlerMessage.Payload = pendingData // Update payload with the modified data
			return HandleGetNotifications(conn, subHandlerMessage)
		default:
			return handleUnsupportedResource(conn, msg.PID, requestData.Action, requestData.Resource)
		}

	case "get_history":
		switch requestData.Resource {
		case "chat":
			// HandleGetChatHistory expects subHandlerMessage.Payload (requestData.Data)
			// to contain {chatId, limit, beforeMessageId}
			return HandleGetChatHistory(conn, subHandlerMessage)
		default:
			return handleUnsupportedResource(conn, msg.PID, requestData.Action, requestData.Resource)
		}

	case "send_message":
		switch requestData.Resource {
		case "chat":
			// HandleSendChatMessage expects subHandlerMessage.Payload (requestData.Data)
			// to contain the message content (e.g., {chatId, text})
			// The original code set a specific message type.
			subHandlerMessage.Type = types.MessageTypeSendChatMessage
			return HandleSendChatMessage(conn, subHandlerMessage)
		default:
			return handleUnsupportedResource(conn, msg.PID, requestData.Action, requestData.Resource)
		}

	default:
		errMsg := fmt.Sprintf("Acción '%s' no soportada.", requestData.Action)
		logger.Warn("HANDLER_DATA", errMsg)
		conn.SendErrorNotification(msg.PID, 400, errMsg)
		return errors.New(errMsg)
	}
}
