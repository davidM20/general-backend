package queries

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
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

/*
 * ===================================================
 * CONSULTAS SQL PARA EL SERVICIO DE FEED
 * ===================================================
 *
 * Este archivo contiene las consultas SQL necesarias para obtener los datos
 * que se mostrarán en el feed de la aplicación.
 *
 * Tablas involucradas:
 * - User: Para obtener perfiles de estudiantes y empresas.
 * - CommunityEvent: Para obtener eventos comunitarios.
 *
 * Consideraciones:
 * - Optimizar las consultas para rendimiento (índices, límites).
 * - Manejar correctamente los campos NULL.
 * - Adaptar los datos al formato wsmodels.FeedItem.
 */

// GetRecentUsersForFeed recupera usuarios (estudiantes y empresas) para el feed.
// TODO: Añadir lógica de paginación/límite y ordenamiento más sofisticado.
func GetRecentUsersForFeed(db *sql.DB, limit int) ([]wsmodels.FeedItem, error) {
	query := `
		SELECT
			u.Id,
			u.FirstName,
			u.LastName,
			u.UserName,
			u.Picture,
			u.Summary, -- Usado como descripción para estudiantes y empresas
			u.Sector,  -- Usado como industria para empresas
			u.CompanyName,
			u.Location,
			u.RoleId,
			u.CreatedAt, -- O u.UpdatedAt para ordenamiento más dinámico
			COALESCE(r.Name, '') as RoleNameDb,
			COALESCE(deg.DegreeName, '') as ExtractedDegreeName, -- Usando el nombre de columna correcto de la tabla Degree
			COALESCE(uni.Name, '') as ExtractedUniversityName -- Usando el nombre de columna correcto de la tabla University
		FROM User u
		LEFT JOIN Role r ON u.RoleId = r.Id
		LEFT JOIN Degree deg ON u.DegreeId = deg.Id
		LEFT JOIN University uni ON u.UniversityId = uni.Id
		WHERE u.StatusAuthorizedId = 1 AND (u.RoleId = ? OR u.RoleId = ?)
		ORDER BY u.CreatedAt DESC
		LIMIT ?;
	`
	logger.Debugf("GetRecentUsersForFeed", "Ejecutando consulta de usuarios para feed con RoleStudent: %d, RoleCompany: %d, Limit: %d", models.RoleStudent, models.RoleBusiness, limit)
	rows, err := db.Query(query, models.RoleStudent, models.RoleBusiness, limit)
	if err != nil {
		logger.Errorf("GetRecentUsersForFeed", "Error al consultar usuarios para feed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var feedItems []wsmodels.FeedItem
	rowCount := 0
	for rows.Next() {
		rowCount++
		var userID int64
		var firstName, lastName, userName, picture, summary, sector, companyName, location, roleNameDbResult sql.NullString
		var degreeNameStr, universityNameStr sql.NullString // Variables para escanear DegreeName y University.Name
		var roleID sql.NullInt64
		var createdAt time.Time

		if err := rows.Scan(
			&userID, &firstName, &lastName, &userName, &picture,
			&summary, &sector, &companyName, &location,
			&roleID, &createdAt, &roleNameDbResult,
			&degreeNameStr, &universityNameStr, // Escaneando los nuevos campos
		); err != nil {
			logger.Errorf("GetRecentUsersForFeed", "Error al escanear fila de usuario (fila procesada #%d): %v", rowCount, err)
			continue
		}

		logger.Debugf("GetRecentUsersForFeed", "Fila de usuario escaneada (fila #%d): UserID: %d, UserName: %s, RoleID: %v, DegreeName: %s, UniversityName: %s", rowCount, userID, userName.String, roleID, degreeNameStr.String, universityNameStr.String)

		itemType := ""
		var data interface{}

		if roleID.Valid && roleID.Int64 == int64(models.RoleStudent) {
			itemType = "student"
			data = wsmodels.StudentFeedData{
				Name:        firstName.String + " " + lastName.String,
				Avatar:      picture.String,
				Career:      degreeNameStr.String,     // Usando el valor escaneado
				University:  universityNameStr.String, // Usando el valor escaneado
				Skills:      []string{},               // TODO: Obtener skills si es necesario
				Description: summary.String,
			}
			logger.Debugf("GetRecentUsersForFeed", "Usuario ID %d (UserName: %s) procesado como ESTUDIANTE.", userID, userName.String)
		} else if roleID.Valid && roleID.Int64 == int64(models.RoleBusiness) {
			itemType = "company"
			data = wsmodels.CompanyFeedData{
				Name:        companyName.String,
				Logo:        picture.String,
				Industry:    sector.String,
				Location:    location.String,
				Description: summary.String,
			}
			logger.Debugf("GetRecentUsersForFeed", "Usuario ID %d (UserName: %s) procesado como EMPRESA.", userID, userName.String)
		} else {
			logger.Warnf("GetRecentUsersForFeed", "Usuario ID %d (UserName: %s) con RoleID %v no coincide con estudiante (%d) o empresa (%d), omitiendo.", userID, userName.String, roleID, models.RoleStudent, models.RoleBusiness)
			continue
		}

		feedItems = append(feedItems, wsmodels.FeedItem{
			ID:        "user-" + userName.String,
			Type:      itemType,
			Timestamp: createdAt.Format(time.RFC3339),
			Data:      data,
		})
	}

	logger.Debugf("GetRecentUsersForFeed", "Procesadas %d filas de la consulta de usuarios. %d items de usuario añadidos al feed.", rowCount, len(feedItems))

	if err = rows.Err(); err != nil {
		logger.Errorf("GetRecentUsersForFeed", "Error después de iterar filas de usuario: %v", err)
		return nil, err
	}
	return feedItems, nil
}

// GetRecentCommunityEventsForFeed recupera eventos comunitarios recientes para el feed.
func GetRecentCommunityEventsForFeed(db *sql.DB, limit int) ([]wsmodels.FeedItem, error) {
	query := `
		SELECT
			ce.Id,
			ce.Title,
			ce.Description,
			ce.EventDate,
			ce.Location,
			ce.ImageUrl,
			ce.OrganizerCompanyName,
			COALESCE(u.Picture, '') as OrganizerLogo, -- Usar la imagen de perfil del usuario
			ce.CreatedAt,
			COALESCE(u.FirstName, '') as CreatorFirstName,
			COALESCE(u.LastName, '') as CreatorLastName
		FROM CommunityEvent ce
		LEFT JOIN User u ON ce.CreatedByUserId = u.Id -- Para obtener nombre del creador si se desea
		ORDER BY ce.CreatedAt DESC
		LIMIT ?;
	`
	rows, err := db.Query(query, limit)
	if err != nil {
		logger.Errorf("GetRecentCommunityEventsForFeed", "Error al consultar eventos para feed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var feedItems []wsmodels.FeedItem
	for rows.Next() {
		var eventID int64
		var title, description, location, imageUrl, organizerCompany, organizerLogo, creatorFirstName, creatorLastName sql.NullString
		var eventDate, createdAt time.Time

		if err := rows.Scan(&eventID, &title, &description, &eventDate, &location, &imageUrl, &organizerCompany, &organizerLogo, &createdAt, &creatorFirstName, &creatorLastName); err != nil {
			logger.Errorf("GetRecentCommunityEventsForFeed", "Error al escanear fila de evento: %v", err)
			continue
		}

		feedItems = append(feedItems, wsmodels.FeedItem{
			ID:        "event-" + strconv.FormatInt(eventID, 10),
			Type:      "event",
			Timestamp: createdAt.Format(time.RFC3339),
			Data: wsmodels.EventFeedData{
				Title:       title.String,
				Company:     organizerCompany.String,
				CompanyLogo: organizerLogo.String, // Esto ahora será u.Picture
				Date:        eventDate.Format("Jan 02, 2006"),
				Location:    location.String,
				Image:       imageUrl.String,
				Description: description.String,
			},
		})
	}
	if err = rows.Err(); err != nil {
		logger.Errorf("GetRecentCommunityEventsForFeed", "Error después de iterar filas de evento: %v", err)
		return nil, err
	}
	return feedItems, nil
}
