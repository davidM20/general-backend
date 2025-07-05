package queries

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

/*
Package queries proporciona un lugar centralizado para toda la lógica de acceso a la base de datos.
Este archivo contiene funciones para interactuar con las tablas de la base de datos.

NORMAS Y DIRECTRICES PARA ESTE ARCHIVO:

1. CONEXIÓN A LA BASE DE DATOS:
  - La variable global `DB *sql.DB` se inicializa en el arranque.
  - NO pasar el puntero de conexión a la base de datos como argumento a las funciones.
  - Todas las funciones de consulta dentro de este paquete deben usar la variable global `DB` directamente.

2. REUTILIZACIÓN Y RESPONSABILIDAD DEL CÓDIGO:
  - Antes de añadir una nueva función, comprueba si una existente puede ser reutilizada o generalizada.
  - Cada función debe tener una única responsabilidad, claramente definida (p. ej., obtener datos de usuario, insertar un mensaje).
  - Mantén las funciones concisas y enfocadas.

3. DOCUMENTACIÓN:
  - Documenta cada nueva función y tipo.
  - Los comentarios deben explicar el propósito de la función, sus parámetros y lo que devuelve.
  - Explica cualquier lógica compleja o comportamiento no obvio.

4. MANEJO DE ERRORES:
  - Comprueba siempre los errores devueltos por `DB.Query`, `DB.QueryRow`, `DB.Exec` y `rows.Scan`.
  - Utiliza `fmt.Errorf("contexto: %w", err)` para envolver los errores, proporcionando contexto sin perder el error original.
  - Maneja `sql.ErrNoRows` específicamente cuando se espera que una consulta a veces no devuelva resultados (p. ej., `GetUserBy...`).

5. CONVENCIONES DE NOMENCLATURA:
  - Sigue las convenciones de nomenclatura idiomáticas de Go (p. ej., `CamelCase` para identificadores exportados).
  - Usa nombres descriptivos para las funciones (p. ej., `GetUserBySessionToken`, `CreateMessage`).

6. CONSTANTES:
  - Para campos de estado o IDs de tipo (p. ej., estado del mensaje), define constantes en la parte superior del archivo.
  - Usa estas constantes en lugar de números mágicos para mejorar la legibilidad y el mantenimiento.

7. MANEJO DE COLUMNAS ANULABLES:
  - Usa `sql.NullString`, `sql.NullInt64`, `sql.NullTime`, etc., para columnas de la base de datos que pueden ser NULL.
  - Comprueba siempre el campo `Valid` antes de acceder al valor de un tipo anulable.

8. SEGURIDAD:
  - Para prevenir la inyección de SQL, SIEMPRE usa consultas parametrizadas con `?` como marcadores de posición.
  - NUNCA construyas consultas concatenando cadenas con entradas proporcionadas por el usuario.

9. AÑADIR NUEVAS CONSULTAS:
  - Agrupa las funciones relacionadas (p. ej., todas las consultas relacionadas con el usuario, todas las relacionadas con los mensajes).
  - Considera las implicaciones de rendimiento. Usa `JOIN`s con criterio y añade cláusulas `LIMIT` donde sea aplicable.
  - Asegúrate de que tu consulta devuelva solo las columnas necesarias.
*/
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
	description := models.ToNullString(&eventData.Description)
	location := models.ToNullString(&eventData.Location)

	var capacity sql.NullInt32
	if eventData.Capacity != nil {
		capacity.Valid = true
		capacity.Int32 = int32(*eventData.Capacity)
	}

	price := models.ToNullFloat64(eventData.Price)

	var tags []string
	if len(eventData.Tags) > 0 && string(eventData.Tags) != "null" {
		if err := json.Unmarshal(eventData.Tags, &tags); err != nil {
			logger.Errorf("COMMUNITY_EVENT_QUERIES", "Error unmarshalling tags to JSON for event '%s': %v", eventData.Title, err)
			return 0, err
		}
	}
	tagsJSON, err := models.TagsToJSON(tags)
	if err != nil {
		logger.Errorf("COMMUNITY_EVENT_QUERIES", "Error marshalling tags to JSON for event '%s': %v", eventData.Title, err)
		return 0, err
	}
	organizerCompanyName := models.ToNullString(&eventData.OrganizerCompanyName)
	organizerUserID := models.ToNullInt64(eventData.OrganizerUserId)
	imageURL := models.ToNullString(&eventData.ImageUrl)

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
			event.Tags = json.RawMessage(tagsJSON.String)
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf("COMMUNITY_EVENT_QUERIES", "Error after iterating through community event rows for user ID %d: %v", userID, err)
		return nil, total, err
	}

	return events, total, nil
}

// CreateCommunityEvent inserta un nuevo evento comunitario en la base de datos,
// incluyendo sus claves fonéticas, y devuelve el ID del nuevo registro.
func CreateCommunityEvent(db *sql.DB, req models.CommunityEventCreateRequest, createdByUserID int64, pKey, sKey string) (int64, error) {
	query := `
        INSERT INTO CommunityEvent (
            Title, Description, EventDate, Location, Capacity, Price, Tags,
            OrganizerCompanyName, OrganizerUserId, OrganizerLogoUrl, ImageUrl, CreatedByUserID,
            dmeta_title_primary, dmeta_title_secondary, CreatedAt, UpdatedAt
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	eventDate, err := time.Parse(time.RFC3339, req.EventDate)
	if err != nil {
		return 0, fmt.Errorf("formato de fecha de evento inválido: %w", err)
	}

	description := models.ToNullString(&req.Description)
	location := models.ToNullString(&req.Location)

	var capacity sql.NullInt32
	if req.Capacity != nil {
		capacity.Valid = true
		capacity.Int32 = int32(*req.Capacity)
	}

	price := models.ToNullFloat64(req.Price)

	var tags []string
	if len(req.Tags) > 0 && string(req.Tags) != "null" {
		if err := json.Unmarshal(req.Tags, &tags); err != nil {
			logger.Errorf("QUERIES", "Error al desglosar etiquetas JSON para el evento '%s': %v", req.Title, err)
			return 0, err
		}
	}
	tagsJSON, err := models.TagsToJSON(tags)
	if err != nil {
		logger.Errorf("QUERIES", "Error al convertir etiquetas a JSON para el evento '%s': %v", req.Title, err)
		return 0, err
	}

	organizerCompanyName := models.ToNullString(&req.OrganizerCompanyName)
	organizerUserID := models.ToNullInt64(req.OrganizerUserId)
	organizerLogoUrl := models.ToNullString(&req.OrganizerLogoUrl)
	imageUrl := models.ToNullString(&req.ImageUrl)
	now := time.Now()

	result, err := db.Exec(query,
		req.Title, description, eventDate, location, capacity, price, tagsJSON,
		organizerCompanyName, organizerUserID, organizerLogoUrl, imageUrl, createdByUserID,
		pKey, sKey, now, now,
	)
	if err != nil {
		logger.Errorf("QUERIES", "Error al insertar un nuevo evento comunitario: %v", err)
		return 0, fmt.Errorf("no se pudo crear el evento en la base de datos: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Errorf("QUERIES", "Error al obtener el LastInsertId para el nuevo evento: %v", err)
		return 0, fmt.Errorf("no se pudo obtener el ID del nuevo evento: %w", err)
	}

	logger.Successf("QUERIES", "Evento comunitario creado con éxito con ID: %d", id)
	return id, nil
}

// GetCommunityEventByID recupera un único evento comunitario por su ID.
func GetCommunityEventByID(db *sql.DB, eventID int64) (*models.CommunityEvent, error) {
	query := `
        SELECT
            Id, Title, Description, EventDate, Location, Capacity, Price, Tags,
            OrganizerCompanyName, OrganizerUserId, OrganizerLogoUrl, ImageUrl, CreatedByUserId,
            CreatedAt, UpdatedAt
        FROM CommunityEvent WHERE Id = ?`

	var event models.CommunityEvent
	var tagsJSON sql.NullString
	err := db.QueryRow(query, eventID).Scan(
		&event.Id, &event.Title, &event.Description, &event.EventDate, &event.Location,
		&event.Capacity, &event.Price, &tagsJSON, &event.OrganizerCompanyName,
		&event.OrganizerUserId, &event.OrganizerLogoUrl, &event.ImageUrl, &event.CreatedByUserId,
		&event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("evento con ID %d no encontrado", eventID)
		}
		logger.Errorf("QUERIES", "Error al escanear el evento con ID %d: %v", eventID, err)
		return nil, err
	}

	if tagsJSON.Valid {
		event.Tags = json.RawMessage(tagsJSON.String)
	}

	return &event, nil
}

// GetMyCommunityEvents recupera una lista paginada de eventos creados por un usuario específico.
func GetMyCommunityEvents(db *sql.DB, userID int64, page, pageSize int) (*models.PaginatedCommunityEvents, error) {
	var totalEvents int
	countQuery := "SELECT COUNT(*) FROM CommunityEvent WHERE CreatedByUserId = ?"
	err := db.QueryRow(countQuery, userID).Scan(&totalEvents)
	if err != nil {
		logger.Errorf("QUERIES", "Error al contar los eventos para el usuario %d: %v", userID, err)
		return nil, err
	}

	if totalEvents == 0 {
		return &models.PaginatedCommunityEvents{
			Data:       []models.CommunityEvent{},
			Pagination: models.PaginationDetails{TotalItems: 0, TotalPages: 0, CurrentPage: page, PageSize: pageSize},
		}, nil
	}

	offset := (page - 1) * pageSize
	query := `
        SELECT
            Id, Title, Description, EventDate, Location, Capacity, Price, Tags,
            OrganizerCompanyName, OrganizerUserId, OrganizerLogoUrl, ImageUrl, CreatedByUserId,
            CreatedAt, UpdatedAt
        FROM CommunityEvent
        WHERE CreatedByUserId = ?
        ORDER BY EventDate DESC
        LIMIT ? OFFSET ?`

	rows, err := db.Query(query, userID, pageSize, offset)
	if err != nil {
		logger.Errorf("QUERIES", "Error al obtener la lista de eventos para el usuario %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	var events []models.CommunityEvent
	for rows.Next() {
		var event models.CommunityEvent
		var tagsJSON sql.NullString
		if err := rows.Scan(
			&event.Id, &event.Title, &event.Description, &event.EventDate, &event.Location,
			&event.Capacity, &event.Price, &tagsJSON, &event.OrganizerCompanyName,
			&event.OrganizerUserId, &event.OrganizerLogoUrl, &event.ImageUrl, &event.CreatedByUserId,
			&event.CreatedAt, &event.UpdatedAt,
		); err != nil {
			logger.Errorf("QUERIES", "Error al escanear la fila del evento: %v", err)
			continue
		}
		if tagsJSON.Valid {
			event.Tags = json.RawMessage(tagsJSON.String)
		}
		events = append(events, event)
	}

	totalPages := int(math.Ceil(float64(totalEvents) / float64(pageSize)))
	return &models.PaginatedCommunityEvents{
		Data: events,
		Pagination: models.PaginationDetails{
			TotalItems:  totalEvents,
			TotalPages:  totalPages,
			CurrentPage: page,
			PageSize:    pageSize,
		},
	}, nil
}
