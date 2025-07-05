package services

import (
	"database/sql"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/davidM20/micro-service-backend-go.git/pkg/phonetic"
)

// CommunityEventService maneja la lógica de negocio para los eventos comunitarios.
type CommunityEventService struct {
	db *sql.DB
}

// NewCommunityEventService crea una nueva instancia de CommunityEventService.
func NewCommunityEventService(db *sql.DB) *CommunityEventService {
	return &CommunityEventService{db: db}
}

// CreateCommunityEvent valida los datos, genera claves fonéticas y crea un nuevo evento.
func (s *CommunityEventService) CreateCommunityEvent(req models.CommunityEventCreateRequest, createdByUserID int64) (*models.CommunityEvent, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("el título del evento no puede estar vacío")
	}

	pKey, sKey, err := phonetic.GenerateKeysForPhrase(req.Title)
	if err != nil {
		logger.Errorf("SERVICE", "Error al generar claves fonéticas para '%s': %v", req.Title, err)
		// No detenemos la creación, simplemente no tendremos claves fonéticas
		pKey, sKey = "", ""
	}

	// Usamos la función de queries en lugar de la lógica de DB directa
	newEventId, err := queries.CreateCommunityEvent(s.db, req, createdByUserID, pKey, sKey)
	if err != nil {
		// El error ya es logueado en la capa de queries
		return nil, err
	}

	// Usamos la función de queries para obtener el evento recién creado
	return queries.GetCommunityEventByID(s.db, newEventId)
}

// GetMyCommunityEvents recupera los eventos de un usuario con paginación.
func (s *CommunityEventService) GetMyCommunityEvents(userID int64, page, pageSize int) (*models.PaginatedCommunityEvents, error) {
	// Usamos la función de queries paginada
	return queries.GetMyCommunityEvents(s.db, userID, page, pageSize)
}
