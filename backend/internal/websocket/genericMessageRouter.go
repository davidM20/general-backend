package websocket

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/handlers"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

/*
REGLAS Y GUÍA PARA MODIFICAR EL ROUTER DE MENSAJES WEBSOCKET

1. ESTRUCTURA DE MENSAJES:
   - Cada mensaje debe tener un 'resource' y una 'action'
   - El payload debe seguir la estructura DataRequestPayload
   - Los recursos y acciones deben ser descriptivos y en minúsculas

2. AGREGAR NUEVOS RECURSOS:
   - Agregar el nuevo recurso en actionHandlers
   - Crear un handler específico para el recurso
   - Seguir el patrón de manejo de errores existente
   - Documentar el nuevo recurso en los comentarios

3. AGREGAR NUEVAS ACCIONES:
   - Agregar la acción bajo el recurso correspondiente en actionHandlers
   - Crear un handler específico si es necesario
   - Mantener la consistencia en el manejo de errores

4. MANEJO DE ERRORES:
   - Usar los tipos de error definidos
   - Incluir mensajes descriptivos
   - Registrar errores usando el logger
   - Notificar al cliente cuando sea necesario

5. CONVENCIÓN DE NOMBRES:
   - Handlers: handle[Resource][Action]
   - Funciones de utilidad: verbos descriptivos
   - Variables: camelCase
   - Constantes: UPPER_CASE

6. DOCUMENTACIÓN:
   - Documentar cada nuevo recurso y sus acciones
   - Explicar el propósito de cada handler
   - Mantener actualizada esta guía

7. RECURSOS DISPONIBLES:
   - chat:
     * get_list: Lista de chats
     * get_history: Historial de chat
     * send_message: Envío de mensajes
   - notification:
     * get_list: Lista de notificaciones
     * get_pending: Notificaciones pendientes
     * mark_read: Marcar notificaciones como leídas
   - dashboard:
     * get_info: Información del panel de control
   - friend:
     * accept_request: Aceptar solicitud de amistad
     * reject_request: Rechazar solicitud de amistad
   - feed:
     * get_list: Obtener lista de items del feed

8. ESTRUCTURA DE PAYLOAD:
   - Para chat/get_history:
     {
       "chatID": string,
       "limit": number,
       "beforeMessageId": string (opcional)
     }
   - Para chat/send_message:
     {
       "text": string,
       "chatID": string,
       "timestamp": string
     }
   - Para notification/mark_read:
     {
       "notificationId": string,
       "timestamp": string
     }
   - Para friend/accept_request y friend/reject_request:
     {
       "notificationId": string,
       "timestamp": string
     }
   - Para feed/get_list:
     No se requiere payload en "data". El servidor devolverá la lista de items del feed.
*/

// DataRequestPayload define la estructura esperada para los mensajes de data_request
type DataRequestPayload struct {
	Action   string                 `json:"action"`             // Acción a realizar (ej: "get_list", "send_message")
	Resource string                 `json:"resource,omitempty"` // Recurso específico (ej: "chat", "notification")
	Data     map[string]interface{} `json:"data,omitempty"`     // Datos específicos para la acción/recurso
}

// ResourceHandler define la interfaz para los manejadores de recursos
type ResourceHandler func(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage, requestData DataRequestPayload) error

// actionHandlers mapea las acciones y recursos a sus respectivos handlers
var actionHandlers = map[string]map[string]ResourceHandler{
	// Chat: Manejo de mensajes y listas de chat
	"chat": {
		"get_list": func(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage, _ DataRequestPayload) error {
			return handlers.HandleGetChatList(conn, msg)
		},
		"get_history": func(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage, requestData DataRequestPayload) error {
			subHandlerMessage := types.ClientToServerMessage{
				PID:     msg.PID,
				Type:    msg.Type,
				Payload: requestData.Data,
			}
			return handlers.HandleGetChatHistory(conn, subHandlerMessage)
		},
		"send_message": handleSendChatMessage,
	},
	// Notification: Manejo de notificaciones
	"notification": {
		"get_list": func(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage, _ DataRequestPayload) error {
			return handlers.HandleGetNotifications(conn, msg)
		},
		"get_pending": handlePendingNotifications,
		"mark_read": func(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage, _ DataRequestPayload) error {
			return handlers.HandleMarkNotificationRead(conn, msg)
		},
	},
	// Dashboard: Información del panel de control
	"dashboard": {
		"get_info": handleDashboardInfo,
	},
	// Friend: Manejo de solicitudes de amistad
	"friend": {
		"accept_request": func(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage, _ DataRequestPayload) error {
			return handlers.HandleAcceptFriendRequest(conn, msg)
		},
		"reject_request": func(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage, _ DataRequestPayload) error {
			return handlers.HandleRejectFriendRequest(conn, msg)
		},
	},
	// Feed: Manejo de items del feed
	"feed": {
		"get_list": func(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage, _ DataRequestPayload) error {
			return handlers.HandleGetFeedList(conn, msg)
		},
	},
}

// HandleDataRequest es el punto de entrada principal para procesar mensajes de data_request.
// Valida y procesa los mensajes entrantes, redirigiendo a los handlers específicos según la acción y recurso.
func HandleDataRequest(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_DATA", "Data request recibida de UserID %d. PID: %s", conn.ID, msg.PID)

	requestData, err := parseRequestPayload(msg)
	if err != nil {
		return err
	}
	if requestData.Action == "ping" {
		return handlePing(conn, msg)
	}

	if requestData.Resource == "" {
		return handleMissingResource(conn, msg.PID, requestData.Action)
	}

	handler, exists := getHandler(requestData.Resource, requestData.Action)
	if !exists {
		return handleUnsupportedResource(conn, msg.PID, requestData.Resource, requestData.Action)
	}

	return handler(conn, msg, requestData)
}

// parseRequestPayload convierte el payload del mensaje en una estructura DataRequestPayload.
// Maneja los errores de marshalling y unmarshalling.
func parseRequestPayload(msg types.ClientToServerMessage) (DataRequestPayload, error) {
	var requestData DataRequestPayload
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		logger.Warnf("HANDLER_DATA", "Error marshalling data_request payload para PID %s, UserID %d: %v", msg.PID, msg.TargetUserID, err)
		return requestData, fmt.Errorf("error marshalling data_request payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("HANDLER_DATA", "Error unmarshalling data_request payload para PID %s, UserID %d: %v. Payload: %s", msg.PID, msg.TargetUserID, err, string(payloadBytes))
		return requestData, fmt.Errorf("error unmarshalling data_request payload: %w", err)
	}
	return requestData, nil
}

// handlePing maneja las solicitudes de ping, enviando una respuesta pong.
// Si no hay PID, retorna silenciosamente.
func handlePing(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	if msg.PID == "" {
		return nil
	}

	ackPayload := types.AckPayload{AcknowledgedPID: msg.PID, Status: "pong"}
	ackMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       types.MessageTypeServerAck,
		FromUserID: conn.ID,
		Payload:    ackPayload,
	}

	if err := conn.SendMessage(ackMsg); err != nil {
		logger.Warnf("HANDLER_DATA", "Error enviando pong (ServerAck) a UserID %d para PID %s: %v", conn.ID, msg.PID, err)
		return err
	}

	logger.Debugf("HANDLER_DATA", "Pong enviado a UserID %d para PID %s", conn.ID, msg.PID)
	return nil
}

// handleDashboardInfo procesa las solicitudes de información del dashboard.
// Envía un ACK inmediato y procesa la solicitud en una goroutine separada.
func handleDashboardInfo(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage, requestData DataRequestPayload) error {
	if msg.PID != "" {
		if err := sendProcessingAck(conn, msg.PID, "processing_dashboard_info"); err != nil {
			logger.Warnf("HANDLER_DATA", "Error enviando ACK para get_info/dashboard a UserID %d, PID %s: %v", conn.ID, msg.PID, err)
		}
	}

	go func(currentConn *customws.Connection[wsmodels.WsUserData], originalMsg types.ClientToServerMessage) {
		if err := handlers.HandleGetDashboardInfo(currentConn, originalMsg); err != nil {
			logger.Errorf("HANDLER_DATA", "Error en goroutine HandleGetDashboardInfo para UserID %d, PID %s: %v", currentConn.ID, originalMsg.PID, err)
		}
	}(conn, msg)

	return nil
}

// handlePendingNotifications procesa las solicitudes de notificaciones pendientes.
// Agrega el flag onlyUnread al payload antes de procesar.
func handlePendingNotifications(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage, requestData DataRequestPayload) error {
	pendingData := requestData.Data
	if pendingData == nil {
		pendingData = make(map[string]interface{})
	}
	pendingData["onlyUnread"] = true

	subHandlerMessage := types.ClientToServerMessage{
		PID:     msg.PID,
		Type:    msg.Type,
		Payload: pendingData,
	}

	return handlers.HandleGetNotifications(conn, subHandlerMessage)
}

// handleSendChatMessage procesa el envío de mensajes de chat.
// Configura el tipo de mensaje específico para chat antes de procesar.
func handleSendChatMessage(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage, requestData DataRequestPayload) error {
	subHandlerMessage := types.ClientToServerMessage{
		PID:     msg.PID,
		Type:    types.MessageTypeSendChatMessage,
		Payload: requestData.Data,
	}
	return handlers.HandleSendChatMessage(conn, subHandlerMessage)
}

// sendProcessingAck envía un ACK de procesamiento al cliente.
// Utilizado para notificar que una solicitud está siendo procesada.
func sendProcessingAck(conn *customws.Connection[wsmodels.WsUserData], pid string, status string) error {
	ackPayload := types.AckPayload{AcknowledgedPID: pid, Status: status}
	ackMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       types.MessageTypeServerAck,
		FromUserID: 0,
		Payload:    ackPayload,
	}
	return conn.SendMessage(ackMsg)
}

// getHandler busca el handler correspondiente para una acción y recurso específicos.
// Retorna el handler y un booleano indicando si se encontró.
func getHandler(resource, action string) (ResourceHandler, bool) {
	if resources, exists := actionHandlers[resource]; exists {
		if handler, exists := resources[action]; exists {
			return handler, true
		}
	}
	return nil, false
}

// handleMissingResource maneja el caso cuando no se especifica un recurso.
// Envía una notificación de error al cliente.
func handleMissingResource(conn *customws.Connection[wsmodels.WsUserData], pid string, action string) error {
	errMsg := fmt.Sprintf("Recurso no especificado para la acción '%s'", action)
	logger.Warnf("HANDLER_DATA", "Missing resource for action '%s'. UserID: %d, PID: %s", action, conn.ID, pid)
	conn.SendErrorNotification(pid, 400, errMsg)
	return errors.New(errMsg)
}
