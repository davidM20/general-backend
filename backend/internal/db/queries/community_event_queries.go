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
