package queries

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
)

// GetUserBySessionToken busca un usuario basado en un token de sesión.
// Esta es una implementación de marcador de posición.
// Deberás reemplazar esto con tu lógica real de consulta a la base de datos.
func GetUserBySessionToken(db *sql.DB, token string) (*models.User, error) {
	// TODO: Implementar la lógica real para validar el token de sesión
	// y recuperar la información del usuario.
	// Esto típicamente implicaría:
	// 1. Consultar la tabla 'Session' para encontrar una sesión activa con ese token.
	// 2. Verificar que la sesión no haya expirado.
	// 3. Obtener el UserID de la sesión.
	// 4. Consultar la tabla 'User' para obtener los detalles del usuario.

	// Implementación de marcador de posición:
	if token == "valid-token-for-user-1" {
		return &models.User{
			Id:       1,
			UserName: "UserOne",
			// Rellena otros campos necesarios si es preciso
		}, nil
	} else if token == "valid-token-for-user-2" {
		return &models.User{
			Id:       2,
			UserName: "UserTwo",
		}, nil
	}

	return nil, sql.ErrNoRows // Simula token no encontrado o inválido
}

// CreateChatMessage inserta un nuevo mensaje de chat en la base de datos.
// Actualiza el ID y CreatedAt del struct msg pasado por referencia con los valores de la BD.
func CreateChatMessage(db *sql.DB, msg *models.ChatMessage) error {
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now().UTC()
	}

	// query := `INSERT INTO "ChatMessage" ("FromUserID", "ToUserID", "Content", "CreatedAt", "StatusID", "ChatID")
	//           VALUES (?, ?, ?, ?, ?, ?) RETURNING "ID", "CreatedAt"` // MySQL no soporta RETURNING directamente en INSERT.
	// Se necesita un enfoque diferente o quitar RETURNING si no es esencial para esta función.
	// Asumiremos por ahora que es para PostgreSQL y lo ajustaremos si es necesario.
	// Confirmado: Es MySQL. RETURNING no es soportado. Se necesita obtener el ID de otra manera.

	// Para MySQL, primero se inserta y luego se puede obtener el LastInsertId si la PK es AUTO_INCREMENT.
	// Si el ID es un UUID generado por la aplicación, no se necesita RETURNING.
	// El esquema de ChatMessage no especifica AUTO_INCREMENT para ID, por lo que asumimos que se genera antes o es un UUID.
	// "CreatedAt" tampoco puede ser devuelto así en MySQL.
	// Simplificaremos: no devolveremos nada y asumiremos que msg.ID ya está asignado (ej. UUID) y CreatedAt es el que se fijó.

	queryMySQL := `INSERT INTO "ChatMessage" ("FromUserID", "ToUserID", "Content", "CreatedAt", "StatusID", "ChatID")
	               VALUES (?, ?, ?, ?, ?, ?)`

	var chatID sql.NullString
	if msg.ChatID != "" {
		chatID = sql.NullString{String: msg.ChatID, Valid: true}
	} else {
		chatID = sql.NullString{Valid: false}
	}

	result, err := db.Exec(queryMySQL,
		msg.FromUserID,
		msg.ToUserID,
		msg.Content,
		msg.CreatedAt, // Este es el valor que se insertará
		msg.StatusID,
		chatID,
	)

	if err != nil {
		return fmt.Errorf("error insertando chat message: %w", err)
	}

	// Si ChatMessage.ID es AUTO_INCREMENT, podemos obtenerlo así:
	id, err := result.LastInsertId()
	if err != nil {
		// Esto podría fallar si la tabla no tiene AUTO_INCREMENT o el driver no lo soporta bien.
		// O si el ID no es int64.
		// Dado que el esquema de Message (que es similar a ChatMessage) usa VARCHAR para ID (UUID),
		// es probable que ChatMessage.ID tampoco sea AUTO_INCREMENT.
		// En ese caso, msg.ID debe ser establecido por la aplicación ANTES de llamar a esta función.
		// Por ahora, asumiremos que si hay un error aquí, es porque no es AUTO_INCREMENT y msg.ID ya está bien.
		// Considerar registrar una advertencia si LastInsertId falla y se esperaba.
	} else {
		msg.ID = id // Actualizar el ID del mensaje si es AUTO_INCREMENT
	}

	// msg.CreatedAt ya está establecido al valor que se intentó insertar.
	// Si la BD tiene un DEFAULT CURRENT_TIMESTAMP y se quiere el valor de la BD,
	// se necesitaría otra consulta SELECT después del INSERT.

	return nil
}

// GetAcceptedContacts recupera todos los contactos aceptados para un userID.
func GetAcceptedContacts(db *sql.DB, userID int64) ([]models.Contact, error) {
	query := `SELECT "ContactId", "User1Id", "User2Id", "Status", "ChatId"
	          FROM "Contact"
	          WHERE ("User1Id" = ? OR "User2Id" = ?) AND "Status" = 'accepted'`

	rows, err := db.Query(query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("error consultando contactos aceptados para userID %d: %w", userID, err)
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var contact models.Contact
		if err := rows.Scan(&contact.ContactId, &contact.User1Id, &contact.User2Id, &contact.Status, &contact.ChatId); err != nil {
			return nil, fmt.Errorf("error escaneando contacto: %w", err)
		}
		contacts = append(contacts, contact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error después de iterar sobre filas de contactos: %w", err)
	}
	return contacts, nil
}

// GetUserBaseInfo recupera información básica del usuario.
func GetUserBaseInfo(db *sql.DB, userID int64) (*models.UserBaseInfo, error) {
	user := &models.UserBaseInfo{}
	query := `SELECT "Id", "FirstName", "LastName", "UserName", "Picture", "RoleId" FROM "User" WHERE "Id" = ?`

	var firstName, lastName, picture sql.NullString

	err := db.QueryRow(query, userID).Scan(
		&user.ID,
		&firstName,
		&lastName,
		&user.UserName,
		&picture,
		&user.RoleId,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario con ID %d no encontrado", userID)
		}
		return nil, fmt.Errorf("error consultando información base de usuario para ID %d: %w", userID, err)
	}

	user.FirstName = firstName.String
	user.LastName = lastName.String
	user.Picture = picture.String

	return user, nil
}

// GetLastChatMessageBetweenUsers recupera el último mensaje entre dos usuarios.
func GetLastChatMessageBetweenUsers(db *sql.DB, userID1 int64, userID2 int64) (*models.ChatMessage, error) {
	query := `SELECT "ID", "ChatID", "FromUserID", "ToUserID", "Content", "CreatedAt", "StatusID"
	          FROM "ChatMessage"
	          WHERE (("FromUserID" = ? AND "ToUserID" = ?) OR ("FromUserID" = ? AND "ToUserID" = ?))
	          ORDER BY "CreatedAt" DESC
	          LIMIT 1`

	msg := &models.ChatMessage{}
	var chatID sql.NullString

	err := db.QueryRow(query, userID1, userID2, userID2, userID1).Scan(
		&msg.ID,
		&chatID,
		&msg.FromUserID,
		&msg.ToUserID,
		&msg.Content,
		&msg.CreatedAt,
		&msg.StatusID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error obteniendo el último mensaje entre %d y %d: %w", userID1, userID2, err)
	}
	if chatID.Valid {
		msg.ChatID = chatID.String
	}

	return msg, nil
}

// GetUnreadMessageCount cuenta los mensajes no leídos para toUserID de fromUserID.
// Asume que StatusID = 3 (o cualquier valor que signifique 'leído')
const ReadStatusID = 3 // Suponiendo que 3 significa "leído"

func GetUnreadMessageCount(db *sql.DB, toUserID int64, fromUserID int64) (int, error) {
	query := `SELECT COUNT(*) FROM "ChatMessage"
	          WHERE "ToUserID" = ? AND "FromUserID" = ? AND "StatusID" < ?`
	// Contamos mensajes con StatusID < ReadStatusID asumiendo que los estados son 1 (sent), 2 (delivered), 3 (read)

	var count int
	err := db.QueryRow(query, toUserID, fromUserID, ReadStatusID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error contando mensajes no leídos para %d de %d: %w", toUserID, fromUserID, err)
	}
	return count, nil
}

// SetUserOnlineStatus actualiza el estado online de un usuario en la tabla 'Online'.
// Si isOnline es true, inserta o actualiza el registro.
// Si isOnline es false, actualiza el estado a offline y la marca de tiempo (como last_seen).
// La tabla 'Online' usa 'CreateAt' como el timestamp.
func SetUserOnlineStatus(db *sql.DB, userID int64, isOnline bool) error {
	now := time.Now().UTC()
	if isOnline {
		// Inserta o actualiza el registro del usuario en la tabla "Online".
		// Se utiliza ON DUPLICATE KEY UPDATE para la operación de "upsert" en MySQL.
		// El estado se establece en 1 (online) y CreateAt se actualiza.
		queryMysql := `INSERT INTO "Online" ("UserOnlineId", "CreateAt", "Status")
		               VALUES (?, ?, 1)
		               ON DUPLICATE KEY UPDATE "CreateAt" = VALUES("CreateAt"), "Status" = 1`
		// Nota: VALUES(ColumnName) en la cláusula UPDATE se refiere al valor que se habría insertado.
		_, err := db.Exec(queryMysql, userID, now)
		if err != nil {
			return fmt.Errorf("error estableciendo estado online para userID %d: %w", userID, err)
		}
	} else {
		// Marcar como offline (Status = 0) y actualizar CreateAt (interpretado como LastSeenAt).
		queryMysql := `UPDATE "Online" SET "Status" = 0, "CreateAt" = ? WHERE "UserOnlineId" = ?`
		res, err := db.Exec(queryMysql, now, userID)
		if err != nil {
			return fmt.Errorf("error estableciendo estado offline para userID %d: %w", userID, err)
		}
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			// Si el usuario no estaba en la tabla Online (por ejemplo, nunca se conectó o fue purgado),
			// no es necesariamente un error. Podríamos insertar un registro offline si la lógica lo requiriera.
			// Por ahora, se considera que no hacer nada está bien si no hay filas afectadas.
			// logger.Warnf("DB_QUERIES", "SetUserOnlineStatus: UserID %d no encontrado en tabla Online al intentar marcar como offline.", userID)
		}
	}
	return nil
}

// GetUserContactIDs recupera los IDs de los contactos aceptados de un usuario.
func GetUserContactIDs(db *sql.DB, userID int64) ([]int64, error) {
	// Selecciona el ID del otro usuario en el contacto
	query := `SELECT
	            CASE
	                WHEN "User1Id" = ? THEN "User2Id"
	                ELSE "User1Id"
	            END AS "OtherUserId"
	          FROM "Contact"
	          WHERE ("User1Id" = ? OR "User2Id" = ?) AND "Status" = 'accepted'`

	rows, err := db.Query(query, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("error consultando IDs de contactos para userID %d: %w", userID, err)
	}
	defer rows.Close()

	var contactIDs []int64
	for rows.Next() {
		var contactID int64
		if err := rows.Scan(&contactID); err != nil {
			return nil, fmt.Errorf("error escaneando ID de contacto: %w", err)
		}
		contactIDs = append(contactIDs, contactID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error después de iterar sobre filas de IDs de contactos: %w", err)
	}
	return contactIDs, nil
}

// --- Notificaciones / Eventos ---

// CreateEvent guarda un nuevo evento/notificación en la base de datos.
// Actualiza el ID del evento pasado por referencia.
func CreateEvent(db *sql.DB, event *models.Event) error {
	if event.CreateAt.IsZero() {
		event.CreateAt = time.Now().UTC()
	}
	query := `INSERT INTO "Event" ("EventType", "EventTitle", "Description", "UserId", "OtherUserId", "ProyectId", "CreateAt", "IsRead")
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query,
		event.EventType,
		event.EventTitle,
		event.Description,
		event.UserId,
		event.OtherUserId, // sql.NullInt64 se maneja directamente por el driver
		event.ProyectId,   // sql.NullInt64 se maneja directamente por el driver
		event.CreateAt,
		event.IsRead,
	)
	if err != nil {
		return fmt.Errorf("error insertando evento: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		// Podría ser que la tabla no use AUTO_INCREMENT o el driver no lo soporte bien
		// o el ID no sea int64. Si el ID es un UUID, debe ser establecido por la app.
		// Por ahora, si hay error, no actualizamos el ID asumiendo que ya está (ej. si fuera UUID)
		// o no es necesario recuperarlo para esta operación.
		// Sin embargo, la tabla Event tiene ID BIGINT AUTO_INCREMENT, por lo que esto debería funcionar.
		return fmt.Errorf("error obteniendo LastInsertId para evento: %w", err)
	}
	event.Id = id
	return nil
}

// GetNotificationsForUser recupera las notificaciones para un usuario, con paginación.
// Incluye lógica para convertir models.Event a wsmodels.NotificationInfo.
func GetNotificationsForUser(db *sql.DB, userID int64, onlyUnread bool, limit int, offset int) ([]wsmodels.NotificationInfo, error) {
	var args []interface{}
	queryStr := `SELECT "Id", "EventType", "EventTitle", "Description", "UserId", "OtherUserId", "ProyectId", "CreateAt", "IsRead"
	            FROM "Event"
	            WHERE "UserId" = ?`
	args = append(args, userID)

	if onlyUnread {
		queryStr += ` AND "IsRead" = FALSE` // O `IsRead = 0` si es TINYINT(1) en MySQL. El DDL dice IsRead bool, pero no especifica. En MySQL bool es TINYINT(1)
	}

	queryStr += ` ORDER BY "CreateAt" DESC`

	if limit > 0 {
		queryStr += ` LIMIT ?`
		args = append(args, limit)
	}
	if offset > 0 {
		queryStr += ` OFFSET ?`
		args = append(args, offset)
	}

	rows, err := db.Query(queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("error consultando notificaciones para userID %d: %w", userID, err)
	}
	defer rows.Close()

	var notifications []wsmodels.NotificationInfo
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(
			&event.Id,
			&event.EventType,
			&event.EventTitle,
			&event.Description,
			&event.UserId,
			&event.OtherUserId,
			&event.ProyectId,
			&event.CreateAt,
			&event.IsRead,
		); err != nil {
			return nil, fmt.Errorf("error escaneando evento: %w", err)
		}

		// Construir el payload para NotificationInfo a partir de los campos de Event
		payload := make(map[string]interface{})
		if event.OtherUserId.Valid {
			payload["otherUserId"] = event.OtherUserId.Int64
			// Podríamos querer cargar el nombre de usuario de OtherUserId aquí
			// otherUserInfo, _ := GetUserBaseInfo(db, event.OtherUserId.Int64)
			// if otherUserInfo != nil { payload["otherUsername"] = otherUserInfo.UserName }
		}
		if event.ProyectId.Valid {
			payload["projectId"] = event.ProyectId.Int64
		}

		notifications = append(notifications, wsmodels.NotificationInfo{
			ID:        fmt.Sprintf("%d", event.Id),
			Type:      event.EventType,
			Title:     event.EventTitle,
			Message:   event.Description,
			Timestamp: event.CreateAt,
			IsRead:    event.IsRead,
			Payload:   payload,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error después de iterar sobre filas de eventos: %w", err)
	}
	return notifications, nil
}

// MarkNotificationAsRead marca una notificación específica como leída para un usuario.
func MarkNotificationAsRead(db *sql.DB, notificationID string, userID int64) error {
	// Asumimos que notificationID es el event.Id convertido a string.
	query := `UPDATE "Event" SET "IsRead" = TRUE WHERE "Id" = ? AND "UserId" = ?`
	// En MySQL, TRUE es 1.
	result, err := db.Exec(query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("error marcando notificación %s como leída para userID %d: %w", notificationID, userID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error obteniendo filas afectadas al marcar notificación %s como leída: %w", notificationID, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("notificación %s no encontrada o no pertenece al usuario %d", notificationID, userID)
	}
	return nil
}

// MarkAllNotificationsAsRead marca todas las notificaciones como leídas para un usuario.
func MarkAllNotificationsAsRead(db *sql.DB, userID int64) (int64, error) {
	query := `UPDATE "Event" SET "IsRead" = TRUE WHERE "UserId" = ? AND "IsRead" = FALSE`
	result, err := db.Exec(query, userID)
	if err != nil {
		return 0, fmt.Errorf("error marcando todas las notificaciones como leídas para userID %d: %w", userID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error obteniendo filas afectadas al marcar todas las notificaciones como leídas: %w", err)
	}
	return rowsAffected, nil
}

// --- Perfil de Usuario ---

// GetUserFullProfileData recupera los datos principales del perfil de un usuario desde la tabla User.
func GetUserFullProfileData(db *sql.DB, userID int64) (*models.User, error) {
	user := &models.User{}
	// Unir con tablas relacionadas para obtener nombres en lugar de solo IDs
	query := `SELECT
	            u.Id, u.FirstName, u.LastName, u.UserName, u.Email, u.Phone, u.Sex, u.DocId,
	            u.NationalityId, n.CountryName AS NationalityName, u.Birthdate, u.Picture,
	            u.DegreeId, d.DegreeName AS DegreeName, u.UniversityId, un.Name AS UniversityName,
	            u.RoleId, r.Name AS RoleName, u.StatusAuthorizedId, u.Summary, u.Address, u.Github, u.Linkedin,
	            u.CreateAt, u.UpdateAt
	        FROM "User" u
	        LEFT JOIN "Nationality" n ON u.NationalityId = n.Id
	        LEFT JOIN "Degree" d ON u.DegreeId = d.Id
	        LEFT JOIN "University" un ON u.UniversityId = un.Id
	        LEFT JOIN "Role" r ON u.RoleId = r.Id
	        WHERE u.Id = ?`

	err := db.QueryRow(query, userID).Scan(
		&user.Id, &user.FirstName, &user.LastName, &user.UserName, &user.Email, &user.Phone, &user.Sex, &user.DocId,
		&user.NationalityId, &user.NationalityName, &user.Birthdate, &user.Picture,
		&user.DegreeId, &user.DegreeName, &user.UniversityId, &user.UniversityName,
		&user.RoleId, &user.RoleName, &user.StatusAuthorizedId, &user.Summary, &user.Address, &user.Github, &user.Linkedin,
		&user.CreateAt, &user.UpdateAt, // Asegúrate que los campos CreateAt y UpdateAt existan en tu struct models.User y tabla User
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario con ID %d no encontrado para perfil completo", userID)
		}
		return nil, fmt.Errorf("error consultando datos completos de perfil para ID %d: %w", userID, err)
	}
	return user, nil
}

// GetEducationItemsForUser recupera los items de educación para un usuario.
func GetEducationItemsForUser(db *sql.DB, personID int64) ([]models.Education, error) {
	query := `SELECT e.Id, e.PersonId, e.Institution, e.Degree, e.Campus, e.GraduationDate, e.CountryId, n.CountryName AS CountryName
	          FROM "Education" e
	          LEFT JOIN "Nationality" n ON e.CountryId = n.Id
	          WHERE e.PersonId = ? ORDER BY e.GraduationDate DESC, e.Id DESC`
	rows, err := db.Query(query, personID)
	if err != nil {
		return nil, fmt.Errorf("error consultando items de educación para PersonID %d: %w", personID, err)
	}
	defer rows.Close()

	var items []models.Education
	for rows.Next() {
		var item models.Education
		var countryName sql.NullString
		if err := rows.Scan(&item.Id, &item.PersonId, &item.Institution, &item.Degree, &item.Campus, &item.GraduationDate, &item.CountryId, &countryName); err != nil {
			return nil, fmt.Errorf("error escaneando item de educación: %w", err)
		}
		item.CountryName = countryName.String // Asignar el nombre del país
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetWorkExperienceItemsForUser recupera los items de experiencia laboral para un usuario.
func GetWorkExperienceItemsForUser(db *sql.DB, personID int64) ([]models.WorkExperience, error) {
	query := `SELECT w.Id, w.PersonId, w.Company, w.Position, w.StartDate, w.EndDate, w.Description, w.CountryId, n.CountryName AS CountryName
	          FROM "WorkExperience" w
	          LEFT JOIN "Nationality" n ON w.CountryId = n.Id
	          WHERE w.PersonId = ? ORDER BY w.EndDate DESC, w.StartDate DESC, w.Id DESC`
	rows, err := db.Query(query, personID)
	if err != nil {
		return nil, fmt.Errorf("error consultando items de experiencia laboral para PersonID %d: %w", personID, err)
	}
	defer rows.Close()

	var items []models.WorkExperience
	for rows.Next() {
		var item models.WorkExperience
		var countryName sql.NullString
		if err := rows.Scan(&item.Id, &item.PersonId, &item.Company, &item.Position, &item.StartDate, &item.EndDate, &item.Description, &item.CountryId, &countryName); err != nil {
			return nil, fmt.Errorf("error escaneando item de experiencia laboral: %w", err)
		}
		item.CountryName = countryName.String // Asignar el nombre del país
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetCertificationItemsForUser recupera las certificaciones para un usuario.
func GetCertificationItemsForUser(db *sql.DB, personID int64) ([]models.Certifications, error) {
	query := `SELECT Id, PersonId, Certification, Institution, DateObtained
	          FROM "Certifications" WHERE PersonId = ? ORDER BY DateObtained DESC, Id DESC`
	rows, err := db.Query(query, personID)
	if err != nil {
		return nil, fmt.Errorf("error consultando certificaciones para PersonID %d: %w", personID, err)
	}
	defer rows.Close()

	var items []models.Certifications
	for rows.Next() {
		var item models.Certifications
		if err := rows.Scan(&item.Id, &item.PersonId, &item.Certification, &item.Institution, &item.DateObtained); err != nil {
			return nil, fmt.Errorf("error escaneando certificación: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetSkillItemsForUser recupera las habilidades para un usuario.
func GetSkillItemsForUser(db *sql.DB, personID int64) ([]models.Skills, error) {
	query := `SELECT Id, PersonId, Skill, Level FROM "Skills" WHERE PersonId = ? ORDER BY Skill ASC, Id DESC`
	rows, err := db.Query(query, personID)
	if err != nil {
		return nil, fmt.Errorf("error consultando skills para PersonID %d: %w", personID, err)
	}
	defer rows.Close()

	var items []models.Skills
	for rows.Next() {
		var item models.Skills
		if err := rows.Scan(&item.Id, &item.PersonId, &item.Skill, &item.Level); err != nil {
			return nil, fmt.Errorf("error escaneando skill: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetLanguageItemsForUser recupera los idiomas para un usuario.
func GetLanguageItemsForUser(db *sql.DB, personID int64) ([]models.Languages, error) {
	query := `SELECT Id, PersonId, Language, Level FROM "Languages" WHERE PersonId = ? ORDER BY Language ASC, Id DESC`
	rows, err := db.Query(query, personID)
	if err != nil {
		return nil, fmt.Errorf("error consultando idiomas para PersonID %d: %w", personID, err)
	}
	defer rows.Close()

	var items []models.Languages
	for rows.Next() {
		var item models.Languages
		if err := rows.Scan(&item.Id, &item.PersonId, &item.Language, &item.Level); err != nil {
			return nil, fmt.Errorf("error escaneando idioma: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetProjectItemsForUser recupera los proyectos para un usuario.
func GetProjectItemsForUser(db *sql.DB, personID int64) ([]models.Project, error) {
	query := `SELECT Id, PersonID, Title, Role, Description, Company, Document, ProjectStatus, StartDate, ExpectedEndDate
	          FROM "Project" WHERE PersonID = ? ORDER BY StartDate DESC, Id DESC`
	rows, err := db.Query(query, personID)
	if err != nil {
		return nil, fmt.Errorf("error consultando proyectos para PersonID %d: %w", personID, err)
	}
	defer rows.Close()

	var items []models.Project
	for rows.Next() {
		var item models.Project
		if err := rows.Scan(
			&item.Id, &item.PersonID, &item.Title, &item.Role, &item.Description, &item.Company,
			&item.Document, &item.ProjectStatus, &item.StartDate, &item.ExpectedEndDate,
		); err != nil {
			return nil, fmt.Errorf("error escaneando proyecto: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// TODO: Funciones para Crear, Actualizar, Eliminar items de perfil (Educación, Experiencia, etc.)
// TODO: Funciones para Actualizar campos del perfil principal (User.Summary, User.Picture, etc.)
