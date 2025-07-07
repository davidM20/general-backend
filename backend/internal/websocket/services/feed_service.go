package services

import (
	"database/sql"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

/*
 * ===================================================
 * SERVICIO PARA LA GESTIÓN DEL FEED
 * ===================================================
 *
 * RESPONSABILIDAD:
 * ----------------
 * Este servicio se encarga de la lógica de negocio para obtener y procesar
 * los items del feed de la aplicación.
 *
 * IMPLEMENTACIÓN ACTUAL:
 * ---------------------
 * Interactúa con la capa de base de datos (queries) para obtener
 * datos reales de usuarios (estudiantes, empresas) y eventos comunitarios.
 * Combina y ordena estos datos para construir el feed.
 *
 * USO:
 * ----
 * Es utilizado por FeedHandler para obtener los datos que se enviarán al cliente
 * a través de WebSocket.
 *
 * INYECCIÓN DE DEPENDENCIAS:
 * -------------------------
 * El servicio se inicializa con una conexión a la base de datos (*sql.DB).
 */

// FeedService maneja la lógica de negocio para el feed.
type FeedService struct {
	DB *sql.DB
}

// NewFeedService crea una nueva instancia de FeedService.
func NewFeedService(db *sql.DB) *FeedService {
	return &FeedService{DB: db}
}

// GetFeedItems obtiene una lista paginada de items para el feed de un usuario.
// Ahora devuelve un payload completo que incluye la información de paginación.
func (s *FeedService) GetFeedItems(userID int64, page, limit int) (*wsmodels.FeedListResponsePayload, error) {
	logger.Infof("FEED_SERVICE", "Usuario %d solicitó items del feed. Página: %d, Límite: %d", userID, page, limit)

	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10 // Límite por defecto
	}
	offset := (page - 1) * limit

	// La nueva función GetUnifiedFeed ya combina y ordena los items en la BD
	// y además devuelve el conteo total de items.
	feedItems, totalItems, err := queries.GetUnifiedFeed(s.DB, userID, limit, offset)
	if err != nil {
		logger.Errorf("FEED_SERVICE", "Error obteniendo el feed unificado para el UserID %d: %v", userID, err)
		return nil, err
	}

	// Calculamos si hay más páginas de forma fiable.
	hasMore := (offset + len(feedItems)) < totalItems

	pagination := &wsmodels.PaginationInfo{
		TotalItems:  totalItems,
		CurrentPage: page,
		HasMore:     hasMore,
	}

	response := &wsmodels.FeedListResponsePayload{
		Items:      feedItems,
		Pagination: pagination,
	}

	logger.Successf("FEED_SERVICE", "Devueltos %d de %d items del feed para el usuario %d. Hay más: %t", len(feedItems), totalItems, userID, hasMore)
	return response, nil
}
