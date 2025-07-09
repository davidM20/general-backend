package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/google/uuid"
)

const handlerSendChatMessageLogComponent = "HANDLER_SEND_CHAT_MESSAGE"

// SendChatMessagePayload define la estructura esperada en msg.Payload para un mensaje de chat.
// Ajusta según lo que realmente envía el cliente.
type SendChatMessagePayload struct {
	ChatId string `json:"chatId"`
	Text   string `json:"text"`
	// Timestamp     int64  `json:"timestamp"` // El timestamp se genera en el backend al guardar
	MediaId       string `json:"mediaId,omitempty"`
	ResponseTo    string `json:"responseTo,omitempty"`    // Para responder a un mensaje específico
	TypeMessageId int64  `json:"typeMessageId,omitempty"` // El backend puede determinar esto o el cliente puede enviarlo
	// TargetUserId int64  `json:"targetUserId"` // No es necesario desde el payload, se infiere en el servicio
}

// DataRequestPayload ya no es necesaria aquí si este handler es específico para send_message
// type DataRequestPayload struct {
// 	Action   string                 `json:"action"`
// 	Resource string                 `json:"resource"`
// 	Data     SendChatMessagePayload `json:"data"`
// }

// HandleSendChatMessage procesa la solicitud para enviar un mensaje de chat.
func HandleSendChatMessage(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof(handlerSendChatMessageLogComponent, "Procesando send_chat_message de UserID %d, PID: %s", conn.ID, msg.PID)

	var payload SendChatMessagePayload
	b, err := json.Marshal(msg.Payload) // Convertir interface{} a bytes
	if err != nil {
		logger.Errorf(handlerSendChatMessageLogComponent, "Error al convertir payload (Marshal) para UserID %d, PID %s: %v", conn.ID, msg.PID, err)
		conn.SendServerAck(msg.PID, "error", fmt.Errorf("payload inválido: %w", err))
		return err // Devuelve el error para que customws pueda manejarlo si es necesario
	}
	// Unmarshal para validar y extraer campos
	if err := json.Unmarshal(b, &payload); err != nil {
		logger.Errorf(handlerSendChatMessageLogComponent, "Error al parsear payload (Unmarshal) para UserID %d, PID %s: %v", conn.ID, msg.PID, err)
		conn.SendServerAck(msg.PID, "error", fmt.Errorf("formato de payload incorrecto: %w", err))
		return err
	}

	// Validaciones básicas del payload
	if payload.ChatId == "" {
		logger.Warnf(handlerSendChatMessageLogComponent, "ChatId vacío en send_message de UserID %d, PID %s", conn.ID, msg.PID)
		conn.SendServerAck(msg.PID, "error", fmt.Errorf("chatId es requerido"))
		return fmt.Errorf("chatId es requerido")
	}
	if payload.Text == "" && payload.MediaId == "" {
		logger.Warnf(handlerSendChatMessageLogComponent, "Mensaje vacío (sin texto ni media) de UserID %d, PID %s", conn.ID, msg.PID)
		conn.SendServerAck(msg.PID, "error", fmt.Errorf("el mensaje no puede estar vacío"))
		return fmt.Errorf("el mensaje no puede estar vacío")
	}

	// Generar un nuevo ID de mensaje único (ULID/UUID recomendado)
	// Usamos el PID del cliente como base o generamos uno nuevo si el PID no es adecuado como ID de mensaje.
	// Por ahora, generaremos un UUID nuevo para el mensaje.
	messageServerID := uuid.NewString()

	// Reconstruir el payload para ProcessAndSaveChatMessage, que espera map[string]interface{}
	// O modificar ProcessAndSaveChatMessage para que acepte un struct.
	// Por ahora, mantenemos la firma de ProcessAndSaveChatMessage.
	servicePayload := map[string]interface{}{
		"chatId":           payload.ChatId,
		"text":             payload.Text, // Para compatibilidad futura
		"content":          payload.Text, // Clave que espera el servicio
		"mediaId":          payload.MediaId,
		"replyToMessageId": payload.ResponseTo, // unificar nomenclatura
		// typeMessageId podría pasarse o dejarse que el servicio lo determine
	}
	if payload.TypeMessageId != 0 {
		servicePayload["typeMessageId"] = payload.TypeMessageId
	}

	// El servicio ahora debería devolver el mensaje guardado
	savedMessage, err := services.ProcessAndSaveChatMessage(conn.ID, servicePayload, messageServerID, conn.Manager())
	if err != nil {
		logger.Errorf(handlerSendChatMessageLogComponent, "Error en ProcessAndSaveChatMessage para UserID %d, PID %s: %v", conn.ID, msg.PID, err)
		conn.SendServerAck(msg.PID, "error", err) // Enviar el error del servicio al cliente
		return fmt.Errorf("error procesando mensaje en servicio: %w", err)
	}

	// Enviar una confirmación de estado 'sent' al remitente original.
	// Esto reemplaza el simple "processed" ACK con una notificación de estado más informativa.
	statusUpdatePayload := map[string]interface{}{
		"originalPID": msg.PID, // El PID original que el cliente envió
		"message":     savedMessage,
	}

	statusUpdateMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       "message_status_update", // Nuevo tipo de mensaje para que el frontend lo maneje
		FromUserID: conn.ID,
		Payload:    statusUpdatePayload,
	}

	if err := conn.SendMessage(statusUpdateMsg); err != nil {
		logger.Errorf(handlerSendChatMessageLogComponent, "Error enviando message_status_update a UserID %d para PID %s: %v", conn.ID, msg.PID, err)
		// No devolvemos error aquí para no cerrar la conexión, pero sí lo registramos.
	}

	logger.Successf(handlerSendChatMessageLogComponent, "Mensaje de UserID %d (ChatID: %s, PID: %s) procesado. Notificación de estado 'sent' enviada.", conn.ID, payload.ChatId, msg.PID)

	// El ACK genérico ya no es necesario si enviamos una actualización de estado.
	// conn.SendServerAck(msg.PID, "processed", nil)

	return nil
}
