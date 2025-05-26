package queries

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/google/uuid"
)

// GetUserBySessionToken busca un usuario basado en un token de sesión.
func GetUserBySessionToken(db *sql.DB, token string) (*models.User, error) {
	// Paso 1: Buscar sesión activa por token
	var userId int64
	var roleId int
	err := db.QueryRow(`
		SELECT UserId, RoleId 
		FROM Session 
		WHERE Tk = ? 
		LIMIT 1
	`, token).Scan(&userId, &roleId)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows // Token no encontrado
		}
		return nil, fmt.Errorf("error querying session: %w", err)
	}

	// Paso 2: Obtener datos del usuario
	var user models.User
	err = db.QueryRow(`
		SELECT 
			Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId,
			NationalityId, Birthdate, Picture, DegreeId, UniversityId,
			RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin
		FROM User 
		WHERE Id = ? AND StatusAuthorizedId = 1
		LIMIT 1
	`, userId).Scan(
		&user.Id, &user.FirstName, &user.LastName, &user.UserName, &user.Email, &user.Phone, &user.Sex, &user.DocId,
		&user.NationalityId, &user.Birthdate, &user.Picture, &user.DegreeId, &user.UniversityId,
		&user.RoleId, &user.StatusAuthorizedId, &user.Summary, &user.Address, &user.Github, &user.Linkedin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows // Usuario no encontrado o inactivo
		}
		return nil, fmt.Errorf("error querying user: %w", err)
	}

	return &user, nil
}

// CreateMessage inserta un nuevo mensaje en la tabla Message.
// Actualiza el ID del struct msg pasado por referencia con el UUID generado.
func CreateMessage(db *sql.DB, msg *models.Message) error {
	// Generar UUID para el ID del mensaje
	msg.Id = uuid.New().String()

	if msg.Date.IsZero() {
		msg.Date = time.Now().UTC()
	}

	// Si no se especifica TypeMessageId, usar 1 (tipo texto por defecto)
	if msg.TypeMessageId == 0 {
		msg.TypeMessageId = 1 // Asumiendo que 1 es el tipo "texto" por defecto
	}

	// Si no se especifica StatusMessage, usar 1 (enviado)
	if msg.StatusMessage == 0 {
		msg.StatusMessage = 1
	}

	queryMySQL := `INSERT INTO Message (Id, TypeMessageId, Text, MediaId, Date, StatusMessage, UserId, ChatId, ChatIdGroup, ResponseTo)
	               VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Exec(queryMySQL,
		msg.Id,
		msg.TypeMessageId,
		msg.Text,
		nil, // MediaId (NULL para mensajes de texto)
		msg.Date,
		msg.StatusMessage,
		msg.UserId,
		msg.ChatId,
		nil, // ChatIdGroup (NULL para chats individuales)
		nil, // ResponseTo (NULL por defecto)
	)

	if err != nil {
		return fmt.Errorf("error insertando mensaje: %w", err)
	}

	return nil
}

// CreateMessageFromChatParams crea un mensaje usando parámetros de chat (fromUserID, toUserID, content)
func CreateMessageFromChatParams(db *sql.DB, fromUserID, toUserID int64, content string) (*models.Message, error) {
	// Buscar el ChatId basado en los usuarios
	chatId, err := getChatIdBetweenUsers(db, fromUserID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo ChatId: %w", err)
	}

	// Crear el mensaje
	msg := &models.Message{
		Text:          content,
		UserId:        fromUserID,
		ChatId:        chatId,
		Date:          time.Now().UTC(),
		TypeMessageId: 1, // Tipo texto por defecto
		StatusMessage: 1, // Enviado
	}

	err = CreateMessage(db, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// getChatIdBetweenUsers busca o establece el ChatId entre dos usuarios
func getChatIdBetweenUsers(db *sql.DB, userID1, userID2 int64) (string, error) {
	var chatId string
	query := `SELECT ChatId FROM Contact 
	          WHERE (User1Id = ? AND User2Id = ?) OR (User1Id = ? AND User2Id = ?) 
	          AND Status = 'accepted' 
	          LIMIT 1`

	err := db.QueryRow(query, userID1, userID2, userID2, userID1).Scan(&chatId)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no existe contacto aceptado entre usuarios %d y %d", userID1, userID2)
		}
		return "", fmt.Errorf("error consultando ChatId: %w", err)
	}

	return chatId, nil
}

// GetAcceptedContacts recupera todos los contactos aceptados para un userID.
func GetAcceptedContacts(db *sql.DB, userID int64) ([]models.Contact, error) {
	query := `SELECT ContactId, User1Id, User2Id, Status, ChatId
	          FROM Contact
	          WHERE (User1Id = ? OR User2Id = ?) AND Status = 'accepted'`

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
	query := `SELECT Id, FirstName, LastName, UserName, Picture, RoleId FROM User WHERE Id = ?`

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

// GetLastMessageBetweenUsers recupera el último mensaje entre dos usuarios.
func GetLastMessageBetweenUsers(db *sql.DB, userID1 int64, userID2 int64) (*models.Message, error) {
	// Primero obtener el ChatId
	chatId, err := getChatIdBetweenUsers(db, userID1, userID2)
	if err != nil {
		return nil, err
	}

	query := `SELECT Id, TypeMessageId, Text, MediaId, Date, StatusMessage, UserId, ChatId, ChatIdGroup, ResponseTo
	          FROM Message
	          WHERE ChatId = ? AND ChatIdGroup IS NULL
	          ORDER BY Date DESC
	          LIMIT 1`

	msg := &models.Message{}
	var mediaId, chatIdGroup, responseTo sql.NullString

	err = db.QueryRow(query, chatId).Scan(
		&msg.Id,
		&msg.TypeMessageId,
		&msg.Text,
		&mediaId,
		&msg.Date,
		&msg.StatusMessage,
		&msg.UserId,
		&msg.ChatId,
		&chatIdGroup,
		&responseTo,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error obteniendo el último mensaje entre %d y %d: %w", userID1, userID2, err)
	}

	// Asignar valores NULL
	if mediaId.Valid {
		msg.MediaId = mediaId.String
	}
	if responseTo.Valid {
		msg.ResponseTo = responseTo.String
	}

	return msg, nil
}

// GetUnreadMessageCount cuenta los mensajes no leídos para toUserID de fromUserID.
// Asume que StatusMessage = 3 (o cualquier valor que signifique 'leído')
const ReadStatusMessage = 3 // Suponiendo que 3 significa "leído"

func GetUnreadMessageCount(db *sql.DB, toUserID int64, fromUserID int64) (int, error) {
	// Obtener ChatId
	chatId, err := getChatIdBetweenUsers(db, fromUserID, toUserID)
	if err != nil {
		return 0, err
	}

	query := `SELECT COUNT(*) FROM Message
	          WHERE ChatId = ? AND UserId = ? AND StatusMessage < ?`
	// Contamos mensajes con StatusMessage < ReadStatusMessage

	var count int
	err = db.QueryRow(query, chatId, fromUserID, ReadStatusMessage).Scan(&count)
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
		queryMysql := `INSERT INTO Online (UserOnlineId, CreateAt, Status)
		               VALUES (?, ?, 1)
		               ON DUPLICATE KEY UPDATE CreateAt = VALUES(CreateAt), Status = 1`
		// Nota: VALUES(ColumnName) en la cláusula UPDATE se refiere al valor que se habría insertado.
		_, err := db.Exec(queryMysql, userID, now)
		if err != nil {
			return fmt.Errorf("error estableciendo estado online para userID %d: %w", userID, err)
		}
	} else {
		// Marcar como offline (Status = 0) y actualizar CreateAt (interpretado como LastSeenAt).
		queryMysql := `UPDATE Online SET Status = 0, CreateAt = ? WHERE UserOnlineId = ?`
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
	                WHEN User1Id = ? THEN User2Id
	                ELSE User1Id
	            END AS OtherUserId
	          FROM Contact
	          WHERE (User1Id = ? OR User2Id = ?) AND Status = 'accepted'`

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
	query := `INSERT INTO Event (Description, UserId, OtherUserId, ProyectId, CreateAt, GroupId)
	          VALUES (?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query,
		event.Description,
		event.UserId,
		event.OtherUserId, // sql.NullInt64 se maneja directamente por el driver
		event.ProyectId,   // sql.NullInt64 se maneja directamente por el driver
		event.CreateAt,
		nil, // GroupId como NULL por defecto
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
	queryStr := `SELECT Id, Description, UserId, OtherUserId, ProyectId, CreateAt, GroupId
	            FROM Event
	            WHERE UserId = ?`
	args = append(args, userID)

	// Nota: onlyUnread no se puede aplicar porque la tabla Event no tiene columna IsRead
	// Por ahora, ignoramos el parámetro onlyUnread y devolvemos todas las notificaciones
	if onlyUnread {
		// Ignoramos el filtro por ahora, ya que no existe la columna IsRead
		// TODO: Agregar columna IsRead a la tabla Event si se necesita
	}

	queryStr += ` ORDER BY CreateAt DESC`

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
		var groupId sql.NullInt64
		if err := rows.Scan(
			&event.Id,
			&event.Description,
			&event.UserId,
			&event.OtherUserId,
			&event.ProyectId,
			&event.CreateAt,
			&groupId,
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
		if groupId.Valid {
			payload["groupId"] = groupId.Int64
		}

		notifications = append(notifications, wsmodels.NotificationInfo{
			ID:        fmt.Sprintf("%d", event.Id),
			Type:      "general",      // Tipo por defecto ya que no existe EventType
			Title:     "Notificación", // Título por defecto ya que no existe EventTitle
			Message:   event.Description,
			Timestamp: event.CreateAt,
			IsRead:    false, // Por defecto false ya que no existe columna IsRead
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
	// NOTA: La tabla Event no tiene columna IsRead, por lo que esta función no hace nada por ahora
	// TODO: Agregar columna IsRead a la tabla Event si se necesita funcionalidad de marcar como leído
	return nil // No hacer nada por ahora
}

// MarkAllNotificationsAsRead marca todas las notificaciones como leídas para un usuario.
func MarkAllNotificationsAsRead(db *sql.DB, userID int64) (int64, error) {
	// NOTA: La tabla Event no tiene columna IsRead, por lo que esta función no hace nada por ahora
	// TODO: Agregar columna IsRead a la tabla Event si se necesita funcionalidad de marcar como leído
	return 0, nil // No hacer nada por ahora
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
	        FROM User u
	        LEFT JOIN Nationality n ON u.NationalityId = n.Id
	        LEFT JOIN Degree d ON u.DegreeId = d.Id
	        LEFT JOIN University un ON u.UniversityId = un.Id
	        LEFT JOIN Role r ON u.RoleId = r.Id
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
	          FROM Education e
	          LEFT JOIN Nationality n ON e.CountryId = n.Id
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
	          FROM WorkExperience w
	          LEFT JOIN Nationality n ON w.CountryId = n.Id
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
	          FROM Certifications WHERE PersonId = ? ORDER BY DateObtained DESC, Id DESC`
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
	query := `SELECT Id, PersonId, Skill, Level FROM Skills WHERE PersonId = ? ORDER BY Skill ASC, Id DESC`
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
	query := `SELECT Id, PersonId, Language, Level FROM Languages WHERE PersonId = ? ORDER BY Language ASC, Id DESC`
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
	          FROM Project WHERE PersonID = ? ORDER BY StartDate DESC, Id DESC`
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
