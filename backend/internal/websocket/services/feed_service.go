package services

import (
	"database/sql"
	"sort"
	"time"

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

// GetFeedItems obtiene la lista de items para el feed.
func (s *FeedService) GetFeedItems(userID int64) ([]wsmodels.FeedItem, error) {
	logger.Infof("FEED_SERVICE", "Usuario %d solicitó items del feed.", userID)

	// Definir límites para las consultas (pueden ser configurables)
	userLimit := 30
	eventLimit := 30

	var allFeedItems []wsmodels.FeedItem

	// Obtener usuarios recientes (estudiantes y empresas)
	userFeedItems, err := queries.GetRecentUsersForFeed(s.DB, userLimit)
	if err != nil {
		logger.Errorf("FEED_SERVICE", "Error obteniendo usuarios para el feed: %v", err)
		// Podríamos decidir devolver un feed parcial o un error completo.
		// Por ahora, continuamos y solo logueamos el error.
	} else {
		allFeedItems = append(allFeedItems, userFeedItems...)
	}

	// Obtener eventos comunitarios recientes
	communityEventFeedItems, err := queries.GetRecentCommunityEventsForFeed(s.DB, eventLimit)
	if err != nil {
		logger.Errorf("FEED_SERVICE", "Error obteniendo eventos comunitarios para el feed: %v", err)
		// Similar al manejo de errores de usuarios.
	} else {
		allFeedItems = append(allFeedItems, communityEventFeedItems...)
	}

	// Ordenar todos los items por Timestamp (del más reciente al más antiguo)
	// Asume que Timestamp es parseable a time.Time. RFC3339 es parseable.
	sort.Slice(allFeedItems, func(i, j int) bool {
		ti, erri := time.Parse(time.RFC3339, allFeedItems[i].Timestamp)
		tj, errj := time.Parse(time.RFC3339, allFeedItems[j].Timestamp)
		// Si hay error al parsear, esos items van al final o se manejan como se prefiera.
		if erri != nil || errj != nil {
			return erri == nil // Pone los no parseables (errj != nil) al final
		}
		return ti.After(tj) // ti > tj para orden descendente
	})

	// TODO: Aplicar un límite general al feed combinado si es necesario, ej. los 20 más recientes.
	// maxTotalFeedItems := 20
	// if len(allFeedItems) > maxTotalFeedItems {
	// 	allFeedItems = allFeedItems[:maxTotalFeedItems]
	// }

	logger.Successf("FEED_SERVICE", "Devueltos %d items del feed para el usuario %d.", len(allFeedItems), userID)
	return allFeedItems, nil
}
