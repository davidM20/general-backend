package queries

import (
	"database/sql"
	"strconv"
	"time"

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

func GetUnifiedFeed(db *sql.DB, userID int64, limit int, offset int) ([]wsmodels.FeedItem, int, error) {
	// Primero, obtenemos el recuento total para la paginación, incluyendo todos los tipos de items.
	countQuery := `
    SELECT COUNT(*) FROM (
        (
            SELECT ce.Id FROM CommunityEvent ce
            WHERE ce.PostType IN ('EVENTO', 'DESAFIO', 'ARTICULO')
        )
        UNION ALL
        (
            SELECT u.Id FROM User u
            WHERE u.StatusAuthorizedId = 1 AND u.RoleId IN (?, ?, ?) -- 1:estudiante, 2:egresado, 3:empresa
        )
    ) as feed_items;
    `
	var totalItems int
	// Los argumentos aquí (1, 2, 3) corresponden a los RoleId para estudiantes, egresados y empresas.
	err := db.QueryRow(countQuery, 1, 2, 3).Scan(&totalItems)
	if err != nil {
		logger.Errorf("GetUnifiedFeed", "Error al contar los items del feed: %v", err)
		return nil, 0, err
	}

	// Consulta principal para obtener los datos de la página actual.
	query := `
    (
        -- Source 1: Community Events (Events, Challenges, Articles, etc.)
        SELECT
            'event' AS item_type,
            ce.Id AS item_id,
            ce.Title AS title,
            ce.Description AS description,
            ce.ImageUrl AS image_url,
            ce.CreatedAt AS created_at,
            ce.PostType AS sub_type,
            -- User related fields (organizer/creator)
            COALESCE(u.Id, 0) as user_id,
            COALESCE(u.FirstName, '') as user_first_name,
            COALESCE(u.LastName, '') as user_last_name,
            COALESCE(u.CompanyName, ce.OrganizerCompanyName) as company_name,
            COALESCE(u.Picture, ce.OrganizerLogoUrl) as user_avatar,
            -- Columnas para hacer match con la query de usuarios
            NULL as user_sector,
            NULL as user_username,
            -- Scoring: Prioritize newer content. Penalize heavily if already viewed.
            (DATEDIFF(NOW(), ce.CreatedAt) * -0.6) + (IF(vi.UserId IS NULL, 0, -100)) AS relevance_score
        FROM
            CommunityEvent ce
        LEFT JOIN User u ON ce.CreatedByUserId = u.Id
        LEFT JOIN FeedItemView vi ON vi.UserId = ? AND vi.ItemType = 'COMMUNITY_EVENT' AND vi.ItemId = ce.Id
        WHERE ce.PostType IN ('EVENTO', 'DESAFIO', 'ARTICULO')
    )
    UNION ALL
    (
        -- Source 2: Users (Students, Graduates, and Companies)
        SELECT
            CASE
                WHEN u.RoleId IN (1, 2) THEN 'student'
                WHEN u.RoleId = 3 THEN 'company'
            END AS item_type,
            u.Id AS item_id,
            CASE
                WHEN u.RoleId = 3 THEN u.CompanyName
                ELSE CONCAT(u.FirstName, ' ', u.LastName)
            END AS title,
            u.Summary AS description,
            u.Picture AS image_url,
            u.CreatedAt AS created_at,
            'profile' AS sub_type,
            u.Id as user_id,
            u.FirstName as user_first_name,
            u.LastName as user_last_name,
            u.CompanyName as company_name,
            u.Picture as user_avatar,
            u.Sector as user_sector,
            u.UserName as user_username,
            -- Scoring: Similar to events, but with slightly less weight on recency.
            (DATEDIFF(NOW(), u.CreatedAt) * -0.5) + (IF(vi.UserId IS NULL, 0, -100)) AS relevance_score
        FROM
            User u
        LEFT JOIN FeedItemView vi ON vi.UserId = ? AND vi.ItemType = 'USER' AND vi.ItemId = u.Id
        WHERE u.StatusAuthorizedId = 1 AND u.RoleId IN (?, ?, ?) -- 1, 2, 3
    )
    -- Final Ordering and Pagination, applied to the whole UNION result.
    ORDER BY relevance_score DESC, created_at DESC, item_id DESC
    LIMIT ? OFFSET ?;
    `

	logger.Debugf("GetUnifiedFeed", "Ejecutando consulta unificada de feed para UserID %d con Limit: %d, Offset: %d", userID, limit, offset)

	// Ejecuta la consulta.
	rows, err := db.Query(query, userID, userID, 1, 2, 3, limit, offset)
	if err != nil {
		logger.Errorf("GetUnifiedFeed", "Error al ejecutar la consulta de feed unificado para UserID %d: %v", userID, err)
		return nil, 0, err
	}
	defer rows.Close()

	var feedItems []wsmodels.FeedItem
	for rows.Next() {
		var itemType, title, description, imageUrl, subType, userFirstName, userLastName, companyName, userAvatar, userSector, userUsername sql.NullString
		var itemID, userID sql.NullInt64
		var createdAt sql.NullTime
		var relevanceScore sql.NullFloat64

		if err := rows.Scan(
			&itemType, &itemID, &title, &description, &imageUrl, &createdAt, &subType,
			&userID, &userFirstName, &userLastName, &companyName, &userAvatar, &userSector, &userUsername,
			&relevanceScore,
		); err != nil {
			logger.Errorf("GetUnifiedFeed", "Error al escanear fila de feed unificado: %v", err)
			continue
		}

		var data interface{}
		idStr := ""

		switch itemType.String {
		case "event":
			idStr = "event-" + strconv.FormatInt(itemID.Int64, 10)
			data = wsmodels.EventFeedData{
				Title:       title.String,
				Company:     companyName.String,
				CompanyLogo: userAvatar.String,
				Date:        formatEventDate(createdAt),
				Location:    companyName.String, // Asumiendo que el evento ocurre en la ubicación de la empresa
				Image:       imageUrl.String,
				Description: description.String,
				PostType:    subType.String,
				EventID:     itemID.Int64,
			}
		case "student":
			idStr = "user-" + strconv.FormatInt(itemID.Int64, 10)
			data = wsmodels.StudentFeedData{
				Name:        title.String,
				Avatar:      userAvatar.String,
				Career:      "Carrera por definir",     // Placeholder
				University:  "Universidad por definir", // Placeholder
				Skills:      []string{},
				Description: description.String,
				UserID:      itemID.Int64,
				UserName:    userUsername.String,
			}
		case "company":
			idStr = "user-" + strconv.FormatInt(itemID.Int64, 10)
			data = wsmodels.CompanyFeedData{
				Name:        title.String,
				Logo:        userAvatar.String,
				Industry:    userSector.String,
				Location:    companyName.String, // Asumiendo que company_name es la ubicación
				Description: description.String,
				UserID:      itemID.Int64,
				UserName:    userUsername.String,
			}
		default:
			logger.Warnf("GetUnifiedFeed", "Tipo de item desconocido encontrado: %s", itemType.String)
			continue
		}

		feedItem := wsmodels.FeedItem{
			ID:        idStr,
			Type:      itemType.String,
			Timestamp: createdAt.Time.Format(time.RFC3339),
			Data:      data,
		}
		feedItems = append(feedItems, feedItem)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf("GetUnifiedFeed", "Error durante el recorrido de las filas del feed: %v", err)
		return nil, 0, err
	}

	logger.Successf("GetUnifiedFeed", "Procesados %d items del feed unificado para el usuario %d", len(feedItems), userID)
	return feedItems, totalItems, nil
}

func formatEventDate(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format("Jan 02, 2006")
	}
	return "" // Return empty string if date is not available
}

// MarkFeedItemsViewed inserta registros de items vistos por un usuario en la BD.
// Utiliza INSERT IGNORE para evitar errores en caso de duplicados.
func MarkFeedItemsViewed(db *sql.DB, userID int64, items []wsmodels.FeedItemViewRef) error {
	if len(items) == 0 {
		return nil
	}

	// Preparamos una transacción para asegurar que todas las inserciones se completen o ninguna lo haga.
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback si algo sale mal

	stmt, err := tx.Prepare("INSERT IGNORE INTO FeedItemView (UserId, ItemType, ItemId, ViewedAt) VALUES (?, ?, ?, NOW())")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		// Normalizar el ItemType para que coincida con el ENUM de la BD
		var dbItemType string
		switch item.ItemType {
		case "student", "company", "user":
			dbItemType = "USER"
		case "event":
			dbItemType = "COMMUNITY_EVENT"
		default:
			logger.Warnf("MarkFeedItemsViewed", "ItemType desconocido '%s' para ItemID %d, omitiendo.", item.ItemType, item.ItemID)
			continue
		}

		if _, err := stmt.Exec(userID, dbItemType, item.ItemID); err != nil {
			logger.Errorf("MarkFeedItemsViewed", "Error ejecutando INSERT para UserID %d, ItemID %d: %v", userID, item.ItemID, err)
			// Continuamos para intentar insertar los demás
		}
	}

	return tx.Commit()
}
