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
package queries

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/google/uuid"
)

var DB *sql.DB

// InitDB inicializa la conexión a la base de datos
func InitDB(database *sql.DB) {
	DB = database
}

const (
	StatusMessageSent       = 1
	StatusMessageDelivered  = 2
	StatusMessageRead       = 3
	StatusMessageError      = 4
	StatusMessagePending    = 0  // Para cuando se está procesando o en cola
	StatusMessageNotSentYet = -1 // Estado inicial antes de intentar enviar
)

// GetUserBySessionToken busca un usuario basado en un token de sesión.
func GetUserBySessionToken(token string) (*models.User, error) {
	// Paso 1: Buscar sesión activa por token
	var userId int64
	var roleId int
	err := DB.QueryRow(`
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
	err = DB.QueryRow(`
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
// Retorna el ID del mensaje creado o un error.
func CreateMessage(msg *models.Message) (string, error) {
	// Generar UUID para el ID del mensaje si no se ha proporcionado uno
	if msg.Id == "" {
		msg.Id = uuid.New().String()
	}

	if msg.Date.IsZero() {
		msg.Date = time.Now().UTC()
	}

	// Si no se especifica TypeMessageId, usar 1 (tipo texto por defecto)
	if msg.TypeMessageId == 0 {
		msg.TypeMessageId = 1 // Asumiendo que 1 es el tipo "texto" por defecto
	}

	// Si no se especifica StatusMessage, usar StatusMessageSent por defecto
	if msg.StatusMessage == 0 {
		msg.StatusMessage = StatusMessageSent
	}

	// Manejar valores opcionales que pueden ser NULL en la BD
	var mediaID sql.NullString
	if msg.MediaId != "" {
		mediaID = sql.NullString{String: msg.MediaId, Valid: true}
	} else {
		mediaID = sql.NullString{Valid: false}
	}

	// ChatIdGroup no está en models.Message, se asume NULL o se maneja directamente en la consulta si es necesario.
	// Para este ejemplo, se insertará NULL para ChatIdGroup.

	var responseTo sql.NullString
	if msg.ResponseTo != "" { // Ajustar si ResponseTo es un tipo diferente
		responseTo = sql.NullString{String: msg.ResponseTo, Valid: true}
	} else {
		responseTo = sql.NullString{Valid: false}
	}

	query := `INSERT INTO Message (Id, TypeMessageId, Text, MediaId, Date, StatusMessage, UserId, ChatId, ChatIdGroup, ResponseTo)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)` // El 10º placeholder es para ChatIdGroup

	_, err := DB.Exec(query,
		msg.Id,
		msg.TypeMessageId,
		msg.Text,
		mediaID,
		msg.Date,
		msg.StatusMessage,
		msg.UserId,
		msg.ChatId,
		nil, // ChatIdGroup se inserta como NULL
		responseTo,
	)

	if err != nil {
		return "", fmt.Errorf("error insertando mensaje: %w", err)
	}

	return msg.Id, nil
}

// CreateMessageFromChatParams crea un mensaje usando parámetros de chat (fromUserID, toUserID, content)
func CreateMessageFromChatParams(fromUserID, toUserID int64, content string) (*models.Message, error) {
	// Buscar el ChatId basado en los usuarios
	chatId, err := getChatIdBetweenUsers(fromUserID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo ChatId: %w", err)
	}

	// Crear el mensaje
	msg := &models.Message{
		Text:          content,
		UserId:        fromUserID,
		ChatId:        chatId,
		Date:          time.Now().UTC(),
		TypeMessageId: 1,                 // Tipo texto por defecto
		StatusMessage: StatusMessageSent, // Usar la constante definida
	}

	// CreateMessage ahora devuelve (string, error)
	msgId, err := CreateMessage(msg)
	if err != nil {
		return nil, err
	}
	msg.Id = msgId // Asignar el ID devuelto al struct

	return msg, nil
}

// getChatIdBetweenUsers busca o establece el ChatId entre dos usuarios
func getChatIdBetweenUsers(userID1, userID2 int64) (string, error) {
	var chatId string
	query := `SELECT ChatId FROM Contact 
	          WHERE (User1Id = ? AND User2Id = ?) OR (User1Id = ? AND User2Id = ?) 
	          AND Status = 'accepted' 
	          LIMIT 1`

	err := DB.QueryRow(query, userID1, userID2, userID2, userID1).Scan(&chatId)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no existe contacto aceptado entre usuarios %d y %d", userID1, userID2)
		}
		return "", fmt.Errorf("error consultando ChatId: %w", err)
	}

	return chatId, nil
}

// GetAcceptedContacts recupera todos los contactos aceptados para un userID.
func GetAcceptedContacts(userID int64) ([]models.Contact, error) {
	query := `SELECT ContactId, User1Id, User2Id, Status, ChatId
	          FROM Contact
	          WHERE (User1Id = ? OR User2Id = ?) AND Status = 'accepted'`

	rows, err := DB.Query(query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("error consultando contactos para userID %d: %w", userID, err)
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var contact models.Contact
		if err := rows.Scan(
			&contact.ContactId,
			&contact.User1Id,
			&contact.User2Id,
			&contact.Status,
			&contact.ChatId,
		); err != nil {
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
func GetUserBaseInfo(userID int64) (*models.UserBaseInfo, error) {
	user := &models.UserBaseInfo{}
	query := `SELECT Id, FirstName, LastName, UserName, Picture, RoleId FROM User WHERE Id = ?`

	var firstName, lastName, picture sql.NullString

	err := DB.QueryRow(query, userID).Scan(
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
func GetLastMessageBetweenUsers(userID1 int64, userID2 int64) (*models.Message, error) {
	// Primero obtener el ChatId
	chatId, err := getChatIdBetweenUsers(userID1, userID2)
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

	err = DB.QueryRow(query, chatId).Scan(
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

func GetUnreadMessageCount(toUserID int64, fromUserID int64) (int, error) {
	// Obtener ChatId
	chatId, err := getChatIdBetweenUsers(fromUserID, toUserID)
	if err != nil {
		return 0, err
	}

	query := `SELECT COUNT(*) FROM Message
	          WHERE ChatId = ? AND UserId = ? AND StatusMessage < ?`
	// Contamos mensajes con StatusMessage < ReadStatusMessage

	var count int
	err = DB.QueryRow(query, chatId, fromUserID, ReadStatusMessage).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error contando mensajes no leídos para %d de %d: %w", toUserID, fromUserID, err)
	}
	return count, nil
}

// SetUserOnlineStatus actualiza el estado online de un usuario en la tabla 'Online'.
// Si isOnline es true, inserta o actualiza el registro.
// Si isOnline es false, actualiza el estado a offline y la marca de tiempo (como last_seen).
// La tabla 'Online' usa 'CreateAt' como el timestamp.
func SetUserOnlineStatus(userID int64, isOnline bool) error {
	now := time.Now().UTC()
	if isOnline {
		// Inserta o actualiza el registro del usuario en la tabla "Online".
		// Se utiliza ON DUPLICATE KEY UPDATE para la operación de "upsert" en MySQL.
		// El estado se establece en 1 (online) y CreateAt se actualiza.
		queryMysql := `INSERT INTO Online (UserOnlineId, CreateAt, Status)
		               VALUES (?, ?, 1)
		               ON DUPLICATE KEY UPDATE CreateAt = VALUES(CreateAt), Status = 1`
		// Nota: VALUES(ColumnName) en la cláusula UPDATE se refiere al valor que se habría insertado.
		_, err := DB.Exec(queryMysql, userID, now)
		if err != nil {
			return fmt.Errorf("error estableciendo estado online para userID %d: %w", userID, err)
		}
	} else {
		// Marcar como offline (Status = 0) y actualizar CreateAt (interpretado como LastSeenAt).
		queryMysql := `UPDATE Online SET Status = 0, CreateAt = ? WHERE UserOnlineId = ?`
		res, err := DB.Exec(queryMysql, now, userID)
		if err != nil {
			return fmt.Errorf("error estableciendo estado offline para userID %d: %w", userID, err)
		}
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			// Si el usuario no estaba en la tabla Online (por ejemplo, nunca se conectó o fue purgado),
			// no es necesariamente un error. Podríamos insertar un registro offline si la lógica lo requiriera.
			// Por ahora, se considera que no hacer nada está bien si no hay filas afectadas.
			logger.Warnf("DB_QUERIES", "SetUserOnlineStatus: UserID %d no encontrado en tabla Online al intentar marcar como offline.", userID)
		}
	}
	return nil
}

// GetUserContactIDs recupera los IDs de los contactos aceptados de un usuario.
func GetUserContactIDs(userID int64) ([]int64, error) {
	// Selecciona el ID del otro usuario en el contacto
	query := `SELECT
	            CASE
	                WHEN User1Id = ? THEN User2Id
	                ELSE User1Id
	            END AS OtherUserId
	          FROM Contact
	          WHERE (User1Id = ? OR User2Id = ?) AND Status = 'accepted'`

	rows, err := DB.Query(query, userID, userID, userID)
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
func CreateEvent(event *models.Event) error {
	if event.CreateAt.IsZero() {
		event.CreateAt = time.Now().UTC()
	}

	query := `INSERT INTO Event (
		EventType, EventTitle, Description, UserId, OtherUserId, 
		ProyectId, CreateAt, IsRead, GroupId, Status, 
		ActionRequired, ActionTakenAt, Metadata
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := DB.Exec(query,
		event.EventType,
		event.EventTitle,
		event.Description,
		event.UserId,
		event.OtherUserId,
		event.ProyectId,
		event.CreateAt,
		event.IsRead,
		event.GroupId,
		event.Status,
		event.ActionRequired,
		event.ActionTakenAt,
		event.Metadata,
	)
	if err != nil {
		return fmt.Errorf("error insertando evento: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error obteniendo LastInsertId para evento: %w", err)
	}
	event.Id = id
	return nil
}

// GetNotificationsForUser recupera todas las notificaciones para un usuario.
// También popula la información del perfil del usuario que originó la notificación (OtherUser) usando un JOIN.
// NOTA: EventType, EventTitle, e IsRead se omiten temporalmente de la consulta a la tabla Event,
//
//	asumiendo que se añadirán a la BD más adelante.
func GetNotificationsForUser(userID int64, onlyUnread bool, limit int, offset int) ([]wsmodels.NotificationInfo, error) {
	var rows *sql.Rows
	var err error

	// Campos a seleccionar: Omitiendo e.EventType, e.EventTitle, e.IsRead temporalmente
	queryFields := `
		e.Id, e.Description, e.CreateAt,
		e.OtherUserId, e.ProyectId,
		u.Id AS ProfileId,
		u.FirstName AS ProfileFirstName,
		u.LastName AS ProfileLastName,
		u.UserName AS ProfileUserName,
		u.Picture AS ProfilePicture,
		u.Email AS ProfileEmail
	`
	// La tabla Event se aliasa como 'e', User como 'u'
	baseQuery := fmt.Sprintf("SELECT %s FROM Event e LEFT JOIN User u ON e.OtherUserId = u.Id WHERE e.UserId = ?", queryFields)
	args := []interface{}{userID}

	if onlyUnread {
		// TODO: Cuando IsRead se añada a la BD y a la consulta, esta condición deberá usar e.IsRead
		// baseQuery += " AND e.IsRead = false"
		// Por ahora, si onlyUnread es true, podría no devolver nada o devolver todo,
		// dependiendo de cómo se quiera manejar la ausencia de IsRead.
		// Para ser seguro y evitar errores, si onlyUnread es true y la columna no existe,
		// podríamos añadir una condición que siempre sea falsa si se quiere simular "no hay no leídas",
		// o simplemente ignorar el filtro por ahora. Ignorémoslo temporalmente.
		fmt.Println("[GetNotificationsForUser] ADVERTENCIA: onlyUnread=true pero la columna IsRead no se está consultando.")
	}
	baseQuery += " ORDER BY e.CreateAt DESC"
	if limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		baseQuery += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err = DB.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("error consultando notificaciones para userID %d: %w", userID, err)
	}
	defer rows.Close()

	var notifications []wsmodels.NotificationInfo
	for rows.Next() {
		var notification wsmodels.NotificationInfo
		var rawCreateAt []byte // Para escanear el timestamp directamente

		// Variables para los campos del perfil que pueden ser NULL
		var profileID sql.NullInt64
		var profileFirstName sql.NullString
		var profileLastName sql.NullString
		var profileUserName sql.NullString
		var profilePicture sql.NullString
		var profileEmail sql.NullString

		// Variables para OtherUserId y ProyectId de la tabla Event
		var otherUserID sql.NullInt64
		var projectID sql.NullInt64

		// Escanear los campos disponibles. notification.Type, notification.Title, notification.IsRead
		// quedarán con sus zero values (string vacío, false).
		err := rows.Scan(
			&notification.ID, &notification.Message, &rawCreateAt, // EventType, EventTitle, IsRead omitidos
			&otherUserID, &projectID,
			&profileID, &profileFirstName, &profileLastName, &profileUserName, &profilePicture, &profileEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando fila de notificación: %w", err)
		}

		// notification.Type = "" (valor por defecto)
		// notification.Title = "" (valor por defecto)
		// notification.IsRead = false (valor por defecto)

		parsedTime, parseErr := time.Parse("2006-01-02 15:04:05", string(rawCreateAt))
		if parseErr != nil {
			parsedTime, parseErr = time.Parse(time.RFC3339, string(rawCreateAt))
			if parseErr != nil {
				return nil, fmt.Errorf("error parseando CreateAt de notificación (%s): %w", string(rawCreateAt), parseErr)
			}
		}
		notification.Timestamp = parsedTime

		payloadMap := make(map[string]interface{})
		if otherUserID.Valid {
			payloadMap["otherUserId"] = otherUserID.Int64
		}
		if projectID.Valid {
			payloadMap["projectId"] = projectID.Int64
		}
		notification.Payload = payloadMap

		if otherUserID.Valid && profileID.Valid {
			notification.Profile = wsmodels.ProfileData{
				ID:        profileID.Int64,
				FirstName: profileFirstName.String,
				LastName:  profileLastName.String,
				UserName:  profileUserName.String,
				Picture:   profilePicture.String,
				Email:     profileEmail.String,
			}
		}

		notifications = append(notifications, notification)
	}

	if errRows := rows.Err(); errRows != nil {
		return nil, fmt.Errorf("error después de iterar sobre filas de eventos: %w", errRows)
	}

	return notifications, nil
}

// MarkNotificationAsRead marca una notificación específica como leída para un usuario.
func MarkNotificationAsRead(notificationID string, userID int64) error {
	// Convertir notificationID de string a int64
	notifID, err := strconv.ParseInt(notificationID, 10, 64)
	if err != nil {
		return fmt.Errorf("ID de notificación inválido: %s", notificationID)
	}

	query := `UPDATE Event SET IsRead = ? WHERE Id = ? AND UserId = ?`
	result, err := DB.Exec(query, true, notifID, userID)
	if err != nil {
		return fmt.Errorf("error marcando notificación %d como leída para usuario %d: %w", notifID, userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error obteniendo filas afectadas al marcar notificación como leída: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notificación %d no encontrada para usuario %d o ya marcada como leída", notifID, userID)
	}

	return nil
}

// MarkAllNotificationsAsRead marca todas las notificaciones como leídas para un usuario.
func MarkAllNotificationsAsRead(userID int64) (int64, error) {
	query := `UPDATE Event SET IsRead = ? WHERE UserId = ? AND IsRead = ?`
	result, err := DB.Exec(query, true, userID, false) // true para marcar como leída, false para seleccionar no leídas
	if err != nil {
		return 0, fmt.Errorf("error marcando todas las notificaciones como leídas para usuario %d: %w", userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error obteniendo filas afectadas al marcar todas las notificaciones como leídas: %w", err)
	}
	return rowsAffected, nil
}

// --- Perfil de Usuario ---

// GetUserFullProfileData recupera los datos principales del perfil de un usuario desde la tabla User.
func GetUserFullProfileData(userID int64) (*models.User, error) {
	user := &models.User{}
	// Unir con tablas relacionadas para obtener nombres en lugar de solo IDs
	query := `SELECT
	            u.Id, u.FirstName, u.LastName, u.UserName, u.Email, u.Phone, u.Sex, u.DocId,
	            u.NationalityId, n.CountryName AS NationalityName, u.Birthdate, u.Picture,
	            u.DegreeId, d.DegreeName AS DegreeName, u.UniversityId, un.Name AS UniversityName,
	            u.RoleId, r.Name AS RoleName, u.StatusAuthorizedId, u.Summary, u.Address, u.Github, u.Linkedin
	        FROM User u
	        LEFT JOIN Nationality n ON u.NationalityId = n.Id
	        LEFT JOIN Degree d ON u.DegreeId = d.Id
	        LEFT JOIN University un ON u.UniversityId = un.Id
	        LEFT JOIN Role r ON u.RoleId = r.Id
	        WHERE u.Id = ?`

	err := DB.QueryRow(query, userID).Scan(
		&user.Id, &user.FirstName, &user.LastName, &user.UserName, &user.Email, &user.Phone, &user.Sex, &user.DocId,
		&user.NationalityId, &user.NationalityName, &user.Birthdate, &user.Picture,
		&user.DegreeId, &user.DegreeName, &user.UniversityId, &user.UniversityName,
		&user.RoleId, &user.RoleName, &user.StatusAuthorizedId, &user.Summary, &user.Address, &user.Github, &user.Linkedin,
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
func GetEducationItemsForUser(personID int64) ([]models.Education, error) {
	query := `SELECT e.Id, e.PersonId, e.Institution, e.Degree, e.Campus, e.GraduationDate, e.CountryId, n.CountryName AS CountryName, e.IsCurrentlyStudying
	          FROM Education e
	          LEFT JOIN Nationality n ON e.CountryId = n.Id
	          WHERE e.PersonId = ? ORDER BY e.GraduationDate DESC, e.Id DESC`
	rows, err := DB.Query(query, personID)
	if err != nil {
		return nil, fmt.Errorf("error consultando items de educación para PersonID %d: %w", personID, err)
	}
	defer rows.Close()

	var items []models.Education
	for rows.Next() {
		var item models.Education
		var countryName sql.NullString
		if err := rows.Scan(&item.Id, &item.PersonId, &item.Institution, &item.Degree, &item.Campus, &item.GraduationDate, &item.CountryId, &countryName, &item.IsCurrentlyStudying); err != nil {
			return nil, fmt.Errorf("error escaneando item de educación: %w", err)
		}
		item.CountryName = countryName // Asignar el nombre del país
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetWorkExperienceItemsForUser recupera los items de experiencia laboral para un usuario.
func GetWorkExperienceItemsForUser(personID int64) ([]models.WorkExperience, error) {
	query := `SELECT w.Id, w.PersonId, w.Company, w.Position, w.StartDate, w.EndDate, w.Description, w.CountryId, n.CountryName AS CountryName, w.IsCurrentJob
	          FROM WorkExperience w
	          LEFT JOIN Nationality n ON w.CountryId = n.Id
	          WHERE w.PersonId = ? ORDER BY w.EndDate DESC, w.StartDate DESC, w.Id DESC`
	rows, err := DB.Query(query, personID)
	if err != nil {
		return nil, fmt.Errorf("error consultando items de experiencia laboral para PersonID %d: %w", personID, err)
	}
	defer rows.Close()

	var items []models.WorkExperience
	for rows.Next() {
		var item models.WorkExperience
		var countryName sql.NullString
		if err := rows.Scan(&item.Id, &item.PersonId, &item.Company, &item.Position, &item.StartDate, &item.EndDate, &item.Description, &item.CountryId, &countryName, &item.IsCurrentJob); err != nil {
			return nil, fmt.Errorf("error escaneando item de experiencia laboral: %w", err)
		}
		item.CountryName = countryName // Asignar el nombre del país
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetCertificationItemsForUser recupera las certificaciones para un usuario.
func GetCertificationItemsForUser(personID int64) ([]models.Certifications, error) {
	query := `SELECT Id, PersonId, Certification, Institution, DateObtained
	          FROM Certifications WHERE PersonId = ? ORDER BY DateObtained DESC, Id DESC`
	rows, err := DB.Query(query, personID)
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
func GetSkillItemsForUser(personID int64) ([]models.Skills, error) {
	query := `SELECT Id, PersonId, Skill, Level FROM Skills WHERE PersonId = ? ORDER BY Skill ASC, Id DESC`
	rows, err := DB.Query(query, personID)
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
func GetLanguageItemsForUser(personID int64) ([]models.Languages, error) {
	query := `SELECT Id, PersonId, Language, Level FROM Languages WHERE PersonId = ? ORDER BY Language ASC, Id DESC`
	rows, err := DB.Query(query, personID)
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
func GetProjectItemsForUser(personID int64) ([]models.Project, error) {
	query := `SELECT Id, PersonID, Title, Role, Description, Company, Document, ProjectStatus, StartDate, ExpectedEndDate, IsOngoing
	          FROM Project WHERE PersonID = ? ORDER BY StartDate DESC, Id DESC`
	rows, err := DB.Query(query, personID)
	if err != nil {
		return nil, fmt.Errorf("error consultando proyectos para PersonID %d: %w", personID, err)
	}
	defer rows.Close()

	var items []models.Project
	for rows.Next() {
		var item models.Project
		if err := rows.Scan(
			&item.Id, &item.PersonID, &item.Title, &item.Role, &item.Description, &item.Company,
			&item.Document, &item.ProjectStatus, &item.StartDate, &item.ExpectedEndDate, &item.IsOngoing,
		); err != nil {
			return nil, fmt.Errorf("error escaneando proyecto: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// TODO: Funciones para Crear, Actualizar, Eliminar items de perfil (Educación, Experiencia, etc.)
// TODO: Funciones para Actualizar campos del perfil principal (User.Summary, User.Picture, etc.)

// GetContactByChatID recupera la información de un contacto por su ChatId.
func GetContactByChatID(chatID string) (*models.Contact, error) {
	contact := &models.Contact{}
	query := `SELECT ContactId, User1Id, User2Id, Status, ChatId
	          FROM Contact
	          WHERE ChatId = ? LIMIT 1`

	err := DB.QueryRow(query, chatID).Scan(
		&contact.ContactId,
		&contact.User1Id,
		&contact.User2Id,
		&contact.Status,
		&contact.ChatId,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contacto con ChatID %s no encontrado", chatID)
		}
		return nil, fmt.Errorf("error consultando contacto por ChatID %s: %w", chatID, err)
	}
	return contact, nil
}

// GetUserOnlineStatus obtiene el estado online de un usuario desde la tabla Online.
// Retorna true si el usuario está en línea (Status = 1) y false si está offline (Status = 0).
func GetUserOnlineStatus(userID int64) (bool, error) {
	var status int
	query := `SELECT Status FROM Online WHERE UserOnlineId = ? LIMIT 1`

	err := DB.QueryRow(query, userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Si no hay registro, asumimos que está offline
		}
		return false, fmt.Errorf("error obteniendo estado online para userID %d: %w", userID, err)
	}

	return status == 1, nil // Status 1 = online, 0 = offline
}

// GetEvents obtiene los eventos/notificaciones de un usuario
func GetEvents(userId int64, onlyUnread bool, limit, offset int) ([]models.Event, error) {
	query := `SELECT 
		Id, EventType, EventTitle, Description, UserId, OtherUserId, 
		ProyectId, CreateAt, IsRead, GroupId, Status, 
		ActionRequired, ActionTakenAt, Metadata
		FROM Event 
		WHERE UserId = ?`

	args := []interface{}{userId}
	if onlyUnread {
		query += " AND IsRead = false"
	}
	query += " ORDER BY CreateAt DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error consultando eventos: %w", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.Id,
			&event.EventType,
			&event.EventTitle,
			&event.Description,
			&event.UserId,
			&event.OtherUserId,
			&event.ProyectId,
			&event.CreateAt,
			&event.IsRead,
			&event.GroupId,
			&event.Status,
			&event.ActionRequired,
			&event.ActionTakenAt,
			&event.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando evento: %w", err)
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando eventos: %w", err)
	}

	return events, nil
}

// MarkEventAsRead marca un evento como leído
func MarkEventAsRead(eventId int64) error {
	query := `UPDATE Event SET IsRead = true WHERE Id = ?`
	result, err := DB.Exec(query, eventId)
	if err != nil {
		return fmt.Errorf("error marcando evento como leído: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error obteniendo filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no se encontró el evento con ID %d", eventId)
	}

	return nil
}

// MarkAllEventsAsRead marca todos los eventos de un usuario como leídos
func MarkAllEventsAsRead(userId int64) (int64, error) {
	query := `UPDATE Event SET IsRead = true WHERE UserId = ? AND IsRead = false`
	result, err := DB.Exec(query, userId)
	if err != nil {
		return 0, fmt.Errorf("error marcando eventos como leídos: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error obteniendo filas afectadas: %w", err)
	}

	return rowsAffected, nil
}

func MarkAllEventsAsReadForUser(userID int64) (int64, error) {
	query := `
		UPDATE Event
		SET IsRead = true, ActionTakenAt = CURRENT_TIMESTAMP
		WHERE UserId = ? AND IsRead = false;`

	result, err := DB.Exec(query, userID)
	if err != nil {
		return 0, fmt.Errorf("MarkAllEventsAsReadForUser: error al ejecutar update: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("MarkAllEventsAsReadForUser: error al obtener filas afectadas: %w", err)
	}

	return rowsAffected, nil
}

// UpdateEventStatus actualiza el estado de un evento y marca la acción tomada
func UpdateEventStatus(eventId int64, status string, metadata interface{}) error {
	query := `UPDATE Event 
		SET Status = ?, 
			ActionRequired = false, 
			ActionTakenAt = CURRENT_TIMESTAMP,
			Metadata = ?
		WHERE Id = ?`

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("error serializando metadata: %w", err)
	}

	result, err := DB.Exec(query, status, metadataJSON, eventId)
	if err != nil {
		return fmt.Errorf("error actualizando estado del evento: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error obteniendo filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no se encontró el evento con ID %d", eventId)
	}

	return nil
}

// GetEventsByUserID recupera los eventos para un usuario específico con paginación y filtro opcional de no leídos.
func GetEventsByUserID(userID int64, onlyUnread bool, limit int, offset int) ([]models.Event, error) {
	var args []interface{}
	query := `
		SELECT Id, EventType, EventTitle, Description, UserId, OtherUserId, ProyectId, CreateAt, IsRead, GroupId, Status, ActionRequired, ActionTakenAt, Metadata
		FROM Event
		WHERE UserId = ?`
	args = append(args, userID)

	if onlyUnread {
		query += " AND IsRead = false"
	}

	query += " ORDER BY CreateAt DESC" // Ordenar por más reciente primero

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}
	query += ";"

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("GetEventsByUserID: error en db.Query: %w", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		var metadataScanValue []byte // Usar []byte para escanear Metadata

		err := rows.Scan(
			&event.Id,
			&event.EventType,
			&event.EventTitle,
			&event.Description,
			&event.UserId,
			&event.OtherUserId,
			&event.ProyectId,
			&event.CreateAt,
			&event.IsRead,
			&event.GroupId,
			&event.Status,
			&event.ActionRequired,
			&event.ActionTakenAt,
			&metadataScanValue, // Escanear en el []byte
		)
		if err != nil {
			// Loguear el error y continuar podría ser una opción si una fila corrupta no debe detener todo
			return nil, fmt.Errorf("GetEventsByUserID: error en rows.Scan: %w", err)
		}

		if metadataScanValue != nil {
			event.Metadata = json.RawMessage(metadataScanValue)
		} else {
			// Si es NULL en la BD, event.Metadata será nil, lo cual es correcto para json.RawMessage
			// o puedes asignarle un JSON vacío si prefieres: event.Metadata = json.RawMessage("null") o json.RawMessage("{}")
			event.Metadata = nil
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetEventsByUserID: error en rows.Err: %w", err)
	}

	return events, nil
}

// UpdateContactStatus actualiza el estado de un contacto entre dos usuarios.
func UpdateContactStatus(userID, otherUserID int64, status string, timestamp string) error {
	query := `
		UPDATE Contact 
		SET Status = ?, 
			UpdatedAt = ? 
		WHERE (User1Id = ? AND User2Id = ?) 
		   OR (User1Id = ? AND User2Id = ?)`

	updatedAt, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return fmt.Errorf("error parseando timestamp: %w", err)
	}

	result, err := DB.Exec(query, status, updatedAt, userID, otherUserID, otherUserID, userID)
	if err != nil {
		return fmt.Errorf("error actualizando estado del contacto: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error obteniendo filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no se encontró el contacto entre los usuarios %d y %d", userID, otherUserID)
	}

	logger.Infof("QUERY_CONTACT", "Estado de contacto actualizado para usuarios %d y %d a '%s'", userID, otherUserID, status)
	return nil
}

// UpdateContactChatId actualiza el chatId de un contacto entre dos usuarios.
func UpdateContactChatId(userID, otherUserID int64, chatId string) error {
	query := "UPDATE Contact SET ChatId = ? WHERE ((User1Id = ? AND User2Id = ?) OR (User1Id = ? AND User2Id = ?))"
	_, err := DB.Exec(query, chatId, userID, otherUserID, otherUserID, userID)
	if err != nil {
		logger.Errorf("QUERY", "Error al actualizar ChatId para los usuarios %d y %d: %v", userID, otherUserID, err)
		return fmt.Errorf("no se pudo actualizar el chatId: %w", err)
	}
	logger.Successf("QUERY", "ChatId actualizado correctamente para los usuarios %d y %d", userID, otherUserID)
	return nil
}

// CreateChat crea un nuevo chat entre dos usuarios y retorna su ID.
func CreateChat(userID, otherUserID int64) (string, error) {
	query := `
		INSERT INTO Chat (Type, CreatedAt, UpdatedAt)
		VALUES ('direct', ?, ?)
		RETURNING ChatId`

	now := time.Now()
	var chatId string

	err := DB.QueryRow(query, now, now).Scan(&chatId)
	if err != nil {
		return "", fmt.Errorf("error creando chat: %w", err)
	}

	// Agregar usuarios al chat
	participantsQuery := `
		INSERT INTO ChatParticipant (ChatId, UserId, CreatedAt, UpdatedAt)
		VALUES (?, ?, ?, ?), (?, ?, ?, ?)`

	_, err = DB.Exec(participantsQuery,
		chatId, userID, now, now,
		chatId, otherUserID, now, now)
	if err != nil {
		return "", fmt.Errorf("error agregando participantes al chat: %w", err)
	}

	logger.Infof("QUERY_CONTACT", "Chat creado entre usuarios %d y %d con ID %s", userID, otherUserID, chatId)
	return chatId, nil
}

// GetNotificationById obtiene una notificación por su ID.
func GetNotificationById(notificationId string) (*models.Notification, error) {
	query := `
		SELECT NotificationId, UserId, Type, Title, Message, 
			   IsRead, CreatedAt, UpdatedAt, OtherUserId,
			   ActionRequired, Status, ActionTakenAt
		FROM Notification
		WHERE NotificationId = ?`

	var notification models.Notification
	var actionTakenAt sql.NullTime

	err := DB.QueryRow(query, notificationId).Scan(
		&notification.NotificationId,
		&notification.UserId,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.IsRead,
		&notification.CreatedAt,
		&notification.UpdatedAt,
		&notification.OtherUserId,
		&notification.ActionRequired,
		&notification.Status,
		&actionTakenAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error obteniendo notificación: %w", err)
	}

	if actionTakenAt.Valid {
		notification.ActionTakenAt = &actionTakenAt.Time
	}

	return &notification, nil
}

func CreateContact(user1ID, user2ID int64, chatID string) error {
	query := "INSERT INTO Contact (User1Id, User2Id, Status, ChatId) VALUES (?, ?, 'pending', ?)"
	_, err := DB.Exec(query, user1ID, user2ID, chatID)
	if err != nil {
		logger.Errorf("QUERY", "Error al crear contacto entre %d y %d: %v", user1ID, user2ID, err)
		return fmt.Errorf("no se pudo crear el contacto: %w", err)
	}
	logger.Successf("QUERY", "Contacto creado exitosamente entre %d y %d con estado 'pending'", user1ID, user2ID)
	return nil
}

// GetChatList recupera la lista de información de chat para un usuario con una única consulta optimizada.
func GetChatList(userID int64) ([]models.ChatInfoQueryResult, error) {
	query := `
WITH LastMessages AS (
    SELECT
        m.ChatId,
        m.Text,
        m.Date,
        m.UserId,
        m.Id,
        ROW_NUMBER() OVER(PARTITION BY m.ChatId ORDER BY m.Date DESC, m.Id DESC) as rn
    FROM Message m
),
UnreadCounts AS (
    SELECT
        m.ChatId,
        m.UserId,
        COUNT(*) as unread
    FROM Message m
    WHERE m.StatusMessage < 3 -- 3 is 'read'
    GROUP BY m.ChatId, m.UserId
)
SELECT
    c.ChatId,
    CASE WHEN c.User1Id = ? THEN c.User2Id ELSE c.User1Id END AS OtherUserID,
    u.RoleId AS OtherUserRoleID,
    u.UserName,
    CASE WHEN u.RoleId = 3 THEN u.CompanyName ELSE u.FirstName END AS OtherFirstName,
    CASE WHEN u.RoleId = 3 THEN '' ELSE u.LastName END AS OtherLastName,
    u.CompanyName AS OtherCompanyName,
    u.Picture,
    lm.Text AS LastMessage,
    lm.Date AS LastMessageTs,
    lm.UserId AS LastMessageFromUserId,
    COALESCE(uc.unread, 0) as UnreadCount
FROM
    Contact c
JOIN
    User u ON u.Id = (CASE WHEN c.User1Id = ? THEN c.User2Id ELSE c.User1Id END)
LEFT JOIN
    LastMessages lm ON lm.ChatId = c.ChatId AND lm.rn = 1
LEFT JOIN
    UnreadCounts uc ON uc.ChatId = c.ChatId AND uc.UserId = u.Id
WHERE
    (c.User1Id = ? OR c.User2Id = ?) AND c.Status = 'accepted'
ORDER BY
    lm.Date DESC
`

	rows, err := DB.Query(query, userID, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying chat list for userID %d: %w", userID, err)
	}
	defer rows.Close()

	var results []models.ChatInfoQueryResult
	for rows.Next() {
		var r models.ChatInfoQueryResult
		err := rows.Scan(
			&r.ChatID,
			&r.OtherUserID,
			&r.OtherUserRoleID,
			&r.OtherUserName,
			&r.OtherFirstName,
			&r.OtherLastName,
			&r.OtherCompanyName,
			&r.OtherPicture,
			&r.LastMessage,
			&r.LastMessageTs,
			&r.LastMessageFromUserId,
			&r.UnreadCount,
		)
		if err != nil {
			logger.Errorf("QUERIES", "Error scanning chat list row: %v", err)
			return nil, fmt.Errorf("error scanning chat list row: %w", err)
		}
		results = append(results, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating chat list rows: %w", err)
	}

	return results, nil
}
