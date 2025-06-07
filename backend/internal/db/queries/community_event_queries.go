package queries

import (
	"database/sql"
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
