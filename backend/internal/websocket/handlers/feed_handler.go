package handlers

import (
	"errors"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

/*
 * ===================================================
 * MANEJADOR DE SOLICITUDES WEBSOCKET PARA EL FEED
 * ===================================================
 *
 * RESPONSABILIDAD:
 * ----------------
 * Este manejador es responsable de procesar las solicitudes WebSocket entrantes
 * relacionadas con el recurso "feed". Específicamente, maneja la acción "get_list"
 * para obtener y enviar la lista de items del feed al cliente.
 *
 * FUNCIONAMIENTO:
 * ---------------
 * 1. Recibe la solicitud del cliente a través del router de mensajes WebSocket.
 * 2. Utiliza FeedService para obtener los datos del feed.
 * 3. Construye un mensaje de respuesta con los items del feed.
 * 4. Envía la respuesta al cliente a través de la conexión WebSocket.
 * 5. Maneja errores y envía notificaciones de error si es necesario.
 *
 * USO:
 * ----
 * Es invocado por el router genérico de mensajes WebSocket (genericMessageRouter.go)
 * cuando se recibe una solicitud para "feed" con la acción "get_list".
 *
 * INYECCIÓN DE DEPENDENCIAS:
 * -------------------------
 * Se inicializa con una instancia de FeedService para desacoplar la lógica de negocio.
 */

// FeedHandler maneja las solicitudes WebSocket para el recurso feed.
type FeedHandler struct {
	feedService *services.FeedService
}

// feedHandlerGlobal es una instancia global del FeedHandler.
// Se inicializa en main o en un paquete de inicialización.
// TODO: Asegurar que esta instancia se inicialice correctamente con sus dependencias.
var feedHandlerGlobal *FeedHandler

// InitializeFeedHandler establece la instancia global del FeedHandler.
// Esta función debe ser llamada durante la inicialización de la aplicación,
// después de crear las instancias de los servicios necesarios.
func InitializeFeedHandler(fs *services.FeedService) {
	feedHandlerGlobal = NewFeedHandler(fs)
	logger.Info("FEED_HANDLER", "FeedHandler inicializado correctamente.")
}

// NewFeedHandler crea una nueva instancia de FeedHandler.
func NewFeedHandler(feedService *services.FeedService) *FeedHandler {
	return &FeedHandler{feedService: feedService}
}

// HandleGetFeedList procesa la solicitud para obtener la lista de items del feed.
// Esta función puede ser llamada directamente desde el router si se opta por no usar la instancia global.
func HandleGetFeedList(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	if feedHandlerGlobal == nil || feedHandlerGlobal.feedService == nil {
		logger.Error("FEED_HANDLER", "HandleGetFeedList llamado pero FeedHandler no está inicializado.")
		conn.SendErrorNotification(msg.PID, 500, "Error interno del servidor: FeedHandler no inicializado.")
		return errors.New("FeedHandler no inicializado")
	}
	return feedHandlerGlobal.getFeedList(conn, msg)
}

// getFeedList es el método interno que realmente maneja la lógica.
func (h *FeedHandler) getFeedList(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	userID := conn.ID
	logger.Infof("FEED_HANDLER", "Procesando get_list para el feed, UserID: %d, PID: %s", userID, msg.PID)

	// Extraer parámetros de paginación del payload
	var page, limit int
	if data, ok := msg.Payload.(map[string]interface{}); ok {
		if p, ok := data["page"].(float64); ok {
			page = int(p)
		}
		if l, ok := data["limit"].(float64); ok {
			limit = int(l)
		}
	}

	// Establecer valores por defecto si no se proporcionaron
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10 // Límite por defecto
	}

	// Enviar ACK para el PID original si existe, para satisfacer al WSClient del frontend
	if msg.PID != "" {
		ackPayload := types.AckPayload{AcknowledgedPID: msg.PID, Status: "processing_feed_list"} // O "received_ok"
		ackMsg := types.ServerToClientMessage{
			PID:        conn.Manager().Callbacks().GeneratePID(),
			Type:       types.MessageTypeServerAck,
			FromUserID: 0, // Sistema
			Payload:    ackPayload,
		}
		if errAck := conn.SendMessage(ackMsg); errAck != nil {
			logger.Warnf("FEED_HANDLER", "Error enviando ServerAck para PID %s a UserID %d: %v", msg.PID, userID, errAck)
			// No retornar error aquí necesariamente, continuar para enviar los datos del feed es prioritario.
		} else {
			logger.Debugf("FEED_HANDLER", "ServerAck enviado para get_list/feed (PID: %s) a UserID %d", msg.PID, userID)
		}
	}

	// El servicio ahora devuelve la estructura de payload completa, lista para ser enviada.
	payload, err := h.feedService.GetFeedItems(userID, page, limit)
	if err != nil {
		// El servicio ya registra el error, así que aquí solo notificamos al cliente.
		errorMsg := fmt.Sprintf("no se pudo obtener el feed para el usuario %d", userID)
		conn.SendErrorNotification(msg.PID, 500, errorMsg)
		return fmt.Errorf("error desde feedService.GetFeedItems: %w", err)
	}

	// Creamos el mensaje de respuesta con el payload que ya tiene el formato correcto.
	responseMessage := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       types.MessageTypeDataEvent,
		Payload:    payload, // El payload ya contiene {items: [...], pagination: {...}}
		FromUserID: 0,       // Indica que es un mensaje del sistema/servidor
	}

	if err := conn.SendMessage(responseMessage); err != nil {
		logger.Errorf("FEED_HANDLER", "Error enviando la lista del feed a UserID %d: %v", userID, err)
		return err
	}

	logger.Successf("FEED_HANDLER", "Lista del feed (data_event) enviada exitosamente a UserID %d. Items: %d", userID, len(payload.Items))
	return nil
}
