package handlers

import (
	"encoding/json"

	"github.com/davidM20/micro-service-backend-go.git/internal/db"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// SearchRequestPayload define la estructura para el payload de una solicitud de búsqueda.
type SearchRequestPayload struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// HandleSearchUsers maneja la búsqueda de usuarios.
// Por ahora, solo registra el evento y no tiene implementación.
func HandleSearchUsers(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("SEARCH_HANDLER", "Búsqueda de usuarios iniciada por UserID %d. PID: %s", conn.ID, msg.PID)
	// TODO: Implementar la lógica de búsqueda de usuarios
	// 1. Parsear y validar el payload de búsqueda (query, limit, offset)
	// 2. Realizar la consulta a la base de datos para buscar usuarios
	// 3. Formatear los resultados para la tarjeta de búsqueda
	// 4. Enviar los resultados de vuelta al cliente
	conn.SendErrorNotification(msg.PID, 501, "Funcionalidad de búsqueda de usuarios no implementada.")
	return nil
}

// HandleSearchCompanies maneja la búsqueda de empresas.
// Por ahora, solo registra el evento y no tiene implementación.
func HandleSearchCompanies(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("SEARCH_HANDLER", "Búsqueda de empresas iniciada por UserID %d. PID: %s", conn.ID, msg.PID)
	// TODO: Implementar la lógica de búsqueda de empresas
	// 1. Parsear y validar el payload de búsqueda (query, limit, offset)
	// 2. Realizar la consulta a la base de datos para buscar empresas
	// 3. Formatear los resultados para la tarjeta de búsqueda
	// 4. Enviar los resultados de vuelta al cliente
	conn.SendErrorNotification(msg.PID, 501, "Funcionalidad de búsqueda de empresas no implementada.")
	return nil
}

// HandleSearchAll maneja la búsqueda de usuarios y empresas.
func HandleSearchAll(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("SEARCH_HANDLER", "Búsqueda 'all' iniciada por UserID %d. PID: %s", conn.ID, msg.PID)

	// 1. Parsear el payload
	var payload SearchRequestPayload
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		logger.Warnf("SEARCH_HANDLER", "Error al decodificar payload de búsqueda 'all': %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de búsqueda inválido.")
		return nil
	}

	// Enviar ACK para confirmar la recepción y evitar timeouts en el cliente
	if msg.PID != "" {
		ackPayload := types.AckPayload{AcknowledgedPID: msg.PID, Status: "processing_search"}
		ackMsg := types.ServerToClientMessage{
			PID:        conn.Manager().Callbacks().GeneratePID(),
			Type:       types.MessageTypeServerAck,
			FromUserID: 0,
			Payload:    ackPayload,
		}
		if err := conn.SendMessage(ackMsg); err != nil {
			logger.Warnf("SEARCH_HANDLER", "Error enviando ACK para búsqueda 'all', PID %s: %v", msg.PID, err)
		}
	}

	// Validar query
	if payload.Query == "" {
		conn.SendErrorNotification(msg.PID, 400, "El término de búsqueda no puede estar vacío.")
		return nil
	}

	// 2. Usar el servicio de búsqueda
	dbConn := db.GetDB()

	searchService := services.NewSearchService(dbConn)

	// Establecer valores por defecto para limit y offset
	if payload.Limit <= 0 {
		payload.Limit = 20 // Límite por defecto
	}
	if payload.Offset < 0 {
		payload.Offset = 0
	}

	results, err := searchService.SearchAll(conn.ID, payload.Query, payload.Limit, payload.Offset)
	if err != nil {
		logger.Errorf("SEARCH_HANDLER", "Error en el servicio de búsqueda 'all': %v", err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al realizar la búsqueda.")
		return nil
	}

	// 3. Enviar los resultados de vuelta al cliente
	responsePayload := map[string]interface{}{
		"results": results,
	}

	responseMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       "search_results",
		FromUserID: 0, // Sistema
		Payload:    responsePayload,
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("SEARCH_HANDLER", "Error al enviar resultados de búsqueda 'all': %v", err)
	}

	return nil
}

// HandleSearchGraduates maneja la búsqueda de egresados.
// Por ahora, solo registra el evento y no tiene implementación.
func HandleSearchGraduates(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("SEARCH_HANDLER", "Búsqueda de egresados iniciada por UserID %d. PID: %s", conn.ID, msg.PID)
	// TODO: Implementar la lógica de búsqueda de egresados
	conn.SendErrorNotification(msg.PID, 501, "Funcionalidad de búsqueda de egresados no implementada.")
	return nil
}
