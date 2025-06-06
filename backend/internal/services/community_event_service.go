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

// CommunityEventResponse es una versión limpia de CommunityEvent para la respuesta JSON
type CommunityEventResponse struct {
	Id                   int64     `json:"id"`
	Title                string    `json:"title"`
	Description          *string   `json:"description,omitempty"`
	EventDate            time.Time `json:"event_date"`
	Location             *string   `json:"location,omitempty"`
	Capacity             *int64    `json:"capacity,omitempty"`
	Price                *float64  `json:"price,omitempty"`
	Tags                 []string  `json:"tags,omitempty"`
	OrganizerCompanyName *string   `json:"organizer_company_name,omitempty"`
	OrganizerUserId      *int64    `json:"organizer_user_id,omitempty"`
	OrganizerLogoUrl     *string   `json:"organizer_logo_url,omitempty"`
	ImageUrl             *string   `json:"image_url,omitempty"`
	CreatedByUserId      int64     `json:"created_by_user_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// PaginatedCommunityEventsResponse es la estructura para la respuesta paginada de eventos.
type PaginatedCommunityEventsResponse struct {
	Events     []CommunityEventResponse `json:"events"`
	Total      int                      `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"pageSize"`
	TotalPages int                      `json:"totalPages"`
}

// convertToResponse convierte un CommunityEvent a CommunityEventResponse
func convertToResponse(event models.CommunityEvent) CommunityEventResponse {
	response := CommunityEventResponse{
		Id:              event.Id,
		Title:           event.Title,
		EventDate:       event.EventDate,
		CreatedByUserId: event.CreatedByUserId,
		CreatedAt:       event.CreatedAt,
		UpdatedAt:       event.UpdatedAt,
	}

	// Convertir campos opcionales
	if event.Description.Valid {
		response.Description = &event.Description.String
	}
	if event.Location.Valid {
		response.Location = &event.Location.String
	}
	if event.Capacity.Valid {
		response.Capacity = &event.Capacity.Int64
	}
	if event.Price.Valid {
		response.Price = &event.Price.Float64
	}
	if event.OrganizerCompanyName.Valid {
		response.OrganizerCompanyName = &event.OrganizerCompanyName.String
	}
	if event.OrganizerUserId.Valid {
		response.OrganizerUserId = &event.OrganizerUserId.Int64
	}
	if event.OrganizerLogoUrl.Valid {
		response.OrganizerLogoUrl = &event.OrganizerLogoUrl.String
	}
	if event.ImageUrl.Valid {
		response.ImageUrl = &event.ImageUrl.String
	}
	response.Tags = event.Tags

	return response
}

// GetMyCommunityEvents recupera los eventos de un usuario de forma paginada.
func (s *CommunityEventService) GetMyCommunityEvents(userID int64, page, pageSize int) (*PaginatedCommunityEventsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // Un valor por defecto
	}

	offset := (page - 1) * pageSize

	events, total, err := queries.GetCommunityEventsByUserIDPaginated(s.DB, userID, pageSize, offset)
	if err != nil {
		// El error ya está logueado en la capa de queries
		return nil, errors.New("no se pudieron recuperar los eventos")
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	// Convertir los eventos a la estructura de respuesta limpia
	cleanEvents := make([]CommunityEventResponse, len(events))
	for i, event := range events {
		cleanEvents[i] = convertToResponse(event)
	}

	response := &PaginatedCommunityEventsResponse{
		Events:     cleanEvents,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	logger.Infof("COMMUNITY_EVENT_SERVICE", "Retrieved %d events for user %d (page %d, pageSize %d)", len(events), userID, page, pageSize)
	return response, nil
}
