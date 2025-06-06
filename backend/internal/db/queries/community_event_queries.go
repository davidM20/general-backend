package queries

import (
	"database/sql"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// InsertCommunityEvent inserta un nuevo evento comunitario en la base de datos.
func InsertCommunityEvent(db *sql.DB, eventData models.CommunityEventCreateRequest, createdByUserID int64) (int64, error) {
	query := `
        INSERT INTO CommunityEvent (
            Title, Description, EventDate, Location, Capacity, Price, Tags, 
            OrganizerCompanyName, OrganizerUserId, ImageUrl, 
            CreatedByUserId, CreatedAt, UpdatedAt
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	now := time.Now()

	// Convertir punteros y slices a tipos SQL adecuados
	description := models.ToNullString(eventData.Description)
	location := models.ToNullString(eventData.Location)
	capacity := models.ToNullInt64(eventData.Capacity)
	price := models.ToNullFloat64(eventData.Price)
	tagsJSON, err := models.TagsToJSON(eventData.Tags)
	if err != nil {
		logger.Errorf("COMMUNITY_EVENT_QUERIES", "Error marshalling tags to JSON for event '%s': %v", eventData.Title, err)
		return 0, err
	}
	organizerCompanyName := models.ToNullString(eventData.OrganizerCompanyName)
	organizerUserID := models.ToNullInt64(eventData.OrganizerUserId)
	imageURL := models.ToNullString(eventData.ImageUrl)

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		return db.Exec(
			query,
			eventData.Title,
			description,
			eventData.EventDate,
			location,
			capacity,
			price,
			tagsJSON,
			organizerCompanyName,
			organizerUserID,
			imageURL,
			createdByUserID,
			now, // CreatedAt
			now, // UpdatedAt
		)
	})

	if err != nil {
		logger.Errorf("COMMUNITY_EVENT_QUERIES", "Error inserting community event '%s': %v", eventData.Title, err)
		return 0, err
	}

	sqlResult := result.(sql.Result)
	newEventId, err := sqlResult.LastInsertId()
	if err != nil {
		logger.Errorf("COMMUNITY_EVENT_QUERIES", "Error getting last insert ID for community event '%s': %v", eventData.Title, err)
		return 0, err
	}

	return newEventId, nil
}

// GetCommunityEventsByUserIDPaginated recupera una lista paginada de eventos creados por un usuario específico.
// También devuelve el recuento total de eventos para ese usuario para la paginación.
func GetCommunityEventsByUserIDPaginated(db *sql.DB, userID int64, limit, offset int) ([]models.CommunityEvent, int, error) {
	// Primero, la consulta para obtener el recuento total
	var total int
	countQuery := "SELECT COUNT(*) FROM CommunityEvent WHERE CreatedByUserId = ?"
	err := db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		logger.Errorf("COMMUNITY_EVENT_QUERIES", "Error counting community events for user ID %d: %v", userID, err)
		return nil, 0, err
	}

	// Si no hay eventos, no necesitamos hacer la segunda consulta
	if total == 0 {
		return []models.CommunityEvent{}, 0, nil
	}

	// Ahora, la consulta para obtener los eventos paginados
	query := `
        SELECT 
            Id, Title, Description, EventDate, Location, Capacity, Price, Tags,
            OrganizerCompanyName, OrganizerUserId, OrganizerLogoUrl, ImageUrl,
            CreatedByUserId, CreatedAt, UpdatedAt
        FROM CommunityEvent 
        WHERE CreatedByUserId = ?
        ORDER BY CreatedAt DESC
        LIMIT ? OFFSET ?
    `
	rows, err := db.Query(query, userID, limit, offset)
	if err != nil {
		logger.Errorf("COMMUNITY_EVENT_QUERIES", "Error fetching paginated community events for user ID %d: %v", userID, err)
		return nil, total, err
	}
	defer rows.Close()

	var events []models.CommunityEvent
	for rows.Next() {
		var event models.CommunityEvent
		var tagsJSON sql.NullString
		err := rows.Scan(
			&event.Id,
			&event.Title,
			&event.Description,
			&event.EventDate,
			&event.Location,
			&event.Capacity,
			&event.Price,
			&tagsJSON,
			&event.OrganizerCompanyName,
			&event.OrganizerUserId,
			&event.OrganizerLogoUrl,
			&event.ImageUrl,
			&event.CreatedByUserId,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			logger.Errorf("COMMUNITY_EVENT_QUERIES", "Error scanning community event row for user ID %d: %v", userID, err)
			continue // O podríamos devolver el error y detener el proceso
		}

		// Deserializar las etiquetas si no son nulas
		if tagsJSON.Valid {
			event.Tags, err = models.TagsFromJSON(tagsJSON)
			if err != nil {
				logger.Warnf("COMMUNITY_EVENT_QUERIES", "Could not unmarshal tags for event ID %d: %v", event.Id, err)
				// No devolver un error, simplemente dejar las etiquetas como nil y continuar
				event.Tags = nil
			}
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf("COMMUNITY_EVENT_QUERIES", "Error after iterating through community event rows for user ID %d: %v", userID, err)
		return nil, total, err
	}

	return events, total, nil
}
