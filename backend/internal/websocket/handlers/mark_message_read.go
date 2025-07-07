package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// HandleMarkMessageRead procesa la petición del cliente para marcar un mensaje como leído.
// Se espera un payload: { "messageId": string }
func HandleMarkMessageRead(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	const logComponent = "HANDLER_MARK_MESSAGE_READ"

	var payload struct {
		MessageId string `json:"messageId"`
	}

	raw, err := json.Marshal(msg.Payload)
	if err != nil {
		logger.Warnf(logComponent, "Error marshalling payload: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "payload inválido")
		return fmt.Errorf("payload inválido: %w", err)
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		logger.Warnf(logComponent, "Error unmarshalling payload: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "payload incorrecto")
		return fmt.Errorf("payload incorrecto: %w", err)
	}

	if payload.MessageId == "" {
		conn.SendErrorNotification(msg.PID, 400, "messageId requerido")
		return fmt.Errorf("messageId requerido")
	}

	if err := services.MarkMessageAsRead(conn.ID, payload.MessageId, conn.Manager()); err != nil {
		logger.Errorf(logComponent, "Error marcando mensaje %s como leído: %v", payload.MessageId, err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno")
		return err
	}

	// ACK
	conn.SendServerAck(msg.PID, "marked_read", nil)
	logger.Infof(logComponent, "Mensaje %s marcado como leído por UserID %d", payload.MessageId, conn.ID)
	return nil
}
