package services

import (
	"database/sql"
	"errors"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	// "github.com/go-playground/validator/v10" // Descomentar si se usa validación
)

// CommunityEventService maneja la lógica de negocio para los eventos comunitarios.
type CommunityEventService struct {
	DB *sql.DB
	// validate *validator.Validate // Descomentar si se usa validación
}

// NewCommunityEventService crea una nueva instancia de CommunityEventService.
func NewCommunityEventService(db *sql.DB) *CommunityEventService {
	return &CommunityEventService{
		DB: db,
		// validate: validator.New(), // Descomentar si se usa validación
	}
}

// CreateCommunityEvent crea un nuevo evento comunitario.
func (s *CommunityEventService) CreateCommunityEvent(eventData models.CommunityEventCreateRequest, createdByUserID int64) (*models.CommunityEvent, error) {
	// // Validar los datos de entrada usando el DTO con etiquetas de validación
	// if err := s.validate.Struct(eventData); err != nil {
	// 	logger.Warnf("COMMUNITY_EVENT_SERVICE", "Validation error for new community event: %v", err)
	// 	return nil, err // Devolver errores de validación específicos
	// }

	// Validaciones básicas (se pueden expandir)
	if eventData.Title == "" {
		return nil, errors.New("el título del evento es obligatorio")
	}
	if eventData.EventDate.IsZero() || eventData.EventDate.Before(time.Now()) {
		return nil, errors.New("la fecha del evento no es válida o es una fecha pasada")
	}

	// Llamar a la consulta para insertar el evento
	newEventId, err := queries.InsertCommunityEvent(s.DB, eventData, createdByUserID)
	if err != nil {
		// El error ya está logueado en la capa de queries
		return nil, errors.New("no se pudo crear el evento comunitario")
	}

	// Construir y devolver el objeto CommunityEvent creado (sin consultar de nuevo por simplicidad)
	// En una aplicación real, podrías querer consultar el evento recién insertado para obtener todos los campos.
	createdEvent := &models.CommunityEvent{
		Id:                   newEventId,
		Title:                eventData.Title,
		Description:          models.ToNullString(eventData.Description),
		EventDate:            eventData.EventDate,
		Location:             models.ToNullString(eventData.Location),
		Capacity:             models.ToNullInt64(eventData.Capacity),
		Price:                models.ToNullFloat64(eventData.Price),
		Tags:                 eventData.Tags, // Directamente desde el request, ya es []string
		OrganizerCompanyName: models.ToNullString(eventData.OrganizerCompanyName),
		OrganizerUserId:      models.ToNullInt64(eventData.OrganizerUserId),
		ImageUrl:             models.ToNullString(eventData.ImageUrl),
		CreatedByUserId:      createdByUserID,
		CreatedAt:            time.Now(), // Aproximación, la DB tiene el valor exacto
		UpdatedAt:            time.Now(), // Aproximación
	}

	logger.Successf("COMMUNITY_EVENT_SERVICE", "Community event '%s' (ID: %d) created successfully by UserID: %d", createdEvent.Title, createdEvent.Id, createdByUserID)
	return createdEvent, nil
}
