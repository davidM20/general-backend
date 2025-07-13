package services

import (
	"database/sql"
	// "encoding/json" // No se usa directamente aquí por ahora
	"errors"
	"fmt"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries" // Alias para el paquete que contiene ChatInfo
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	customwsTypes "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

var chatDB *sql.DB // Renombrado para evitar colisión si otros servicios usan 'db'

// InitializeChatService permite inyectar la dependencia de la base de datos.
// Esta función debería ser llamada desde main.go después de conectar a la BD.
func InitializeChatService(database *sql.DB) {
	chatDB = database
	logger.Info("SERVICE_CHAT", "ChatService inicializado con conexión a BD.")
}

// GetChatListForUser recupera la lista de chats para un usuario dado.
// Esto implicaría consultar la base de datos para encontrar todos los chats
// en los que el usuario participa, el último mensaje de cada chat, etc.
func GetChatListForUser(userID int64, manager *customws.ConnectionManager[wsmodels.WsUserData]) ([]wsmodels.ChatInfo, error) {
	if chatDB == nil {
		return nil, errors.New("chat service no inicializado con conexión a BD")
	}
	logger.Infof("SERVICE_CHAT", "Recuperando lista de chats para UserID: %d", userID)

	// Usar la nueva consulta optimizada
	results, err := queries.GetChatList(userID)
	if err != nil {
		logger.Errorf("SERVICE_CHAT", "Error obteniendo la lista de chats optimizada para UserID %d: %v", userID, err)
		return nil, fmt.Errorf("error obteniendo lista de chats: %w", err)
	}

	var chatList []wsmodels.ChatInfo
	for _, r := range results {
		isOnline := manager.IsUserOnline(r.OtherUserID)

		chatType := ""
		if r.OtherUserRoleID == 3 {
			chatType = "company"
		} else if r.OtherUserRoleID == 2 {
			chatType = "graduate" // egresado
		} else {
			chatType = "student"
		}
		chatInfo := wsmodels.ChatInfo{
			ChatID:        r.ChatID,
			OtherUserID:   r.OtherUserID,
			OtherPicture:  r.OtherPicture.String,
			IsOtherOnline: isOnline,
			UnreadCount:   r.UnreadCount,
			Type:          chatType,
		}

		if r.OtherUserRoleID == 3 {
			// Para empresas, usar CompanyName. Si está vacío, usar UserName como fallback.
			displayName := r.OtherCompanyName.String
			if displayName == "" {
				displayName = r.OtherUserName.String
			}
			chatInfo.OtherFirstName = displayName // Asignar a otherFirstName como solicitaste.
			chatInfo.OtherUserName = displayName  // Asignar también a otherUserName para asegurar visibilidad.
			chatInfo.OtherLastName = ""
		} else {
			// Para usuarios normales, usar su nombre y apellido.
			chatInfo.OtherUserName = r.OtherUserName.String
			chatInfo.OtherFirstName = r.OtherFirstName.String
			chatInfo.OtherLastName = r.OtherLastName.String
		}

		if r.LastMessage.Valid {
			chatInfo.LastMessage = r.LastMessage.String
			chatInfo.LastMessageTs = r.LastMessageTs.Time.UnixMilli()
			chatInfo.LastMessageFromUserId = r.LastMessageFromUserId.Int64
		}

		chatList = append(chatList, chatInfo)
	}

	logger.Successf("SERVICE_CHAT", "Lista de chats recuperada para UserID: %d. Número de chats: %d", userID, len(chatList))
	return chatList, nil
}

func ProcessAndSaveChatMessage(userID int64, payload map[string]interface{}, messageID string, manager *customws.ConnectionManager[wsmodels.WsUserData]) (*wsmodels.MessageDB, error) {
	if chatDB == nil {
		return nil, errors.New("servicio de chat no inicializado con conexión a BD")
	}
	if manager == nil {
		return nil, errors.New("ConnectionManager no proporcionado a ProcessAndSaveChatMessage")
	}

	// Extraer y validar campos del payload
	chatId, _ := payload["chatId"].(string)
	chatIdGroup, _ := payload["chatIdGroup"].(string)

	// Un mensaje debe pertenecer a un chat privado o a un grupo, no a ambos.
	if (chatId == "" && chatIdGroup == "") || (chatId != "" && chatIdGroup != "") {
		return nil, errors.New("se debe proporcionar un chatId o un chatIdGroup, pero no ambos")
	}

	content, _ := payload["content"].(string)
	mediaId, _ := payload["mediaId"].(string) // Este es el FileName
	replyToMessageId, _ := payload["replyToMessageId"].(string)

	var realMediaId string
	var err error
	if mediaId != "" {
		// Buscar el ID real del multimedia a partir del FileName
		query := "SELECT Id FROM Multimedia WHERE FileName = ?"
		err = chatDB.QueryRow(query, mediaId).Scan(&realMediaId)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Warnf("SERVICE_CHAT", "Multimedia con FileName %s no encontrado para UserID %d", mediaId, userID)
				return nil, fmt.Errorf("multimedia no encontrado con FileName: %s", mediaId)
			}
			logger.Errorf("SERVICE_CHAT", "Error buscando media por FileName para UserID %d: %v", userID, err)
			return nil, fmt.Errorf("error interno al buscar multimedia: %w", err)
		}
	}

	// Un mensaje debe tener contenido de texto o un adjunto.
	if content == "" && realMediaId == "" {
		return nil, errors.New("el mensaje no puede estar vacío, debe contener contenido o media")
	}

	// Determinar TypeMessageId basado en si hay MediaId o no.
	var typeMessageID int64 = 1 // Por defecto, texto
	if realMediaId != "" {
		typeMessageID = 2 // Asumimos 2 para mensajes con media.
	}

	// --- Guardar el mensaje en la base de datos con el nuevo esquema ---
	sentAt := time.Now().UTC()
	status := "sent"

	// Usamos sql.NullString para campos que podrían estar vacíos
	dbChatId := sql.NullString{String: chatId, Valid: chatId != ""}
	dbChatIdGroup := sql.NullString{String: chatIdGroup, Valid: chatIdGroup != ""}
	dbContent := sql.NullString{String: content, Valid: content != ""}
	dbMediaId := sql.NullString{String: realMediaId, Valid: realMediaId != ""}
	dbReplyToId := sql.NullString{String: replyToMessageId, Valid: replyToMessageId != ""}

	query := `INSERT INTO Message (Id, ChatId, ChatIdGroup, SenderId, Content, Status, TypeMessageId, MediaId, ReplyToMessageId, SentAt) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = chatDB.Exec(query, messageID, dbChatId, dbChatIdGroup, userID, dbContent, status, typeMessageID, dbMediaId, dbReplyToId, sentAt)
	if err != nil {
		logContext := fmt.Sprintf("UserID %d", userID)
		if chatId != "" {
			logContext += fmt.Sprintf(", ChatID %s", chatId)
		}
		if chatIdGroup != "" {
			logContext += fmt.Sprintf(", ChatIdGroup %s", chatIdGroup)
		}
		logger.Errorf("SERVICE_CHAT", "Error guardando mensaje para %s: %v", logContext, err)
		return nil, fmt.Errorf("error guardando mensaje en DB: %w", err)
	}

	logger.Infof("SERVICE_CHAT", "Mensaje guardado (ID: %s) de UserID %d", messageID, userID)

	// --- Construir el objeto de mensaje para la transmisión y retorno ---
	var contentPtr, mediaIdPtr, replyToPtr *string
	if dbContent.Valid {
		contentPtr = &dbContent.String
	}
	if dbMediaId.Valid {
		mediaIdPtr = &dbMediaId.String
	}
	if dbReplyToId.Valid {
		replyToPtr = &dbReplyToId.String
	}

	var chatIdPtr, chatIdGroupPtr *string
	if dbChatId.Valid {
		chatIdPtr = &dbChatId.String
	}
	if dbChatIdGroup.Valid {
		chatIdGroupPtr = &dbChatIdGroup.String
	}

	messageToSend := &wsmodels.MessageDB{
		Id:               messageID,
		ChatId:           chatIdPtr,
		ChatIdGroup:      chatIdGroupPtr,
		SenderId:         userID,
		Content:          contentPtr,
		SentAt:           sentAt.Format(time.RFC3339Nano),
		Status:           status,
		TypeMessageId:    typeMessageID,
		MediaId:          mediaIdPtr,
		ReplyToMessageId: replyToPtr,
	}

	// --- Lógica para encontrar destinatario(s) y enviar si están en línea ---
	if chatId != "" {
		// Lógica para chat privado (1 a 1)
		contact, err := queries.GetContactByChatID(chatId)
		if err != nil {
			logger.Errorf("SERVICE_CHAT", "Error obteniendo información del contacto para ChatID %s después de guardar mensaje: %v", chatId, err)
			return messageToSend, fmt.Errorf("mensaje guardado pero error obteniendo datos del chat para envío: %w", err)
		}

		var recipientUserID int64
		switch userID {
		case contact.User1Id:
			recipientUserID = contact.User2Id
		case contact.User2Id:
			recipientUserID = contact.User1Id
		default:
			logger.Errorf("SERVICE_CHAT", "El remitente del mensaje (UserID %d) no coincide con los participantes del ContactID %s (User1: %d, User2: %d)", userID, contact.ContactId, contact.User1Id, contact.User2Id)
			return messageToSend, fmt.Errorf("mensaje guardado pero remitente no coincide con participantes del chat")
		}

		if manager.IsUserOnline(recipientUserID) {
			serverMessage := customwsTypes.ServerToClientMessage{
				Type:       customwsTypes.MessageTypeNewChatMessage,
				FromUserID: userID,
				Payload:    messageToSend,
				PID:        manager.Callbacks().GeneratePID(),
			}

			if err := manager.SendMessageToUser(recipientUserID, serverMessage); err != nil {
				logger.Errorf("SERVICE_CHAT", "Error enviando mensaje (ID: %s) a UserID %d: %v", messageToSend.Id, recipientUserID, err)
			} else {
				logger.Successf("SERVICE_CHAT", "Mensaje (ID: %s) enviado exitosamente a UserID %d", messageToSend.Id, recipientUserID)
			}
		} else {
			logger.Infof("SERVICE_CHAT", "Destinatario UserID %d no está en línea, mensaje (ID: %s) guardado pero no enviado inmediatamente.", recipientUserID, messageToSend.Id)
		}

	} else if chatIdGroup != "" {
		// Lógica para chat de grupo
		// Asumiendo que existe una función `GetGroupMembersByChatID` que retorna los miembros del grupo.
		groupMembers, err := queries.GetGroupMembersByChatID(chatIdGroup)
		if err != nil {
			logger.Errorf("SERVICE_CHAT", "Error obteniendo miembros del grupo para ChatIdGroup %s: %v", chatIdGroup, err)
			return messageToSend, fmt.Errorf("mensaje guardado pero no se pudieron obtener los miembros del grupo: %w", err)
		}

		serverMessage := customwsTypes.ServerToClientMessage{
			Type:       customwsTypes.MessageTypeNewChatMessage,
			FromUserID: userID,
			Payload:    messageToSend,
			PID:        manager.Callbacks().GeneratePID(),
		}

		for _, member := range groupMembers {
			// No enviar el mensaje al propio remitente.
			if member.UserID == userID {
				continue
			}

			if manager.IsUserOnline(member.UserID) {
				if err := manager.SendMessageToUser(member.UserID, serverMessage); err != nil {
					logger.Errorf("SERVICE_CHAT", "Error enviando mensaje de grupo (ID: %s) a miembro %d: %v", messageToSend.Id, member.UserID, err)
				} else {
					logger.Successf("SERVICE_CHAT", "Mensaje de grupo (ID: %s) enviado exitosamente a miembro %d", messageToSend.Id, member.UserID)
				}
			}
		}
	}

	return messageToSend, nil
}

func GetChatHistory(chatID string, userID int64, limit int, beforeMessageID string, manager *customws.ConnectionManager[wsmodels.WsUserData]) ([]wsmodels.MessageDB, error) {
	if chatDB == nil {
		return nil, errors.New("GetChatHistory: chat service no inicializado con conexión a BD")
	}

	logger.Infof("SERVICE_CHAT", "Recuperando historial para ChatID: %s, UserID: %d, Limit: %d, BeforeMessageID: %s", chatID, userID, limit, beforeMessageID)

	// Consulta base
	query := `
        SELECT Id, SenderId, Content, SentAt, Status, TypeMessageId, MediaId, ReplyToMessageId, EditedAt, ChatIdGroup
        FROM Message
        WHERE ChatId = ?
    `
	args := []interface{}{chatID}

	// Si se requiere paginación con beforeMessageID, se obtiene la fecha e ID del mensaje ancla.
	if beforeMessageID != "" {
		var anchorDate time.Time
		var anchorID string
		err := chatDB.QueryRow(
			"SELECT SentAt, Id FROM Message WHERE Id = ? AND ChatId = ?",
			beforeMessageID, chatID,
		).Scan(&anchorDate, &anchorID)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Warnf("SERVICE_CHAT", "beforeMessageID %s no encontrado para ChatID %s", beforeMessageID, chatID)
				return []wsmodels.MessageDB{}, nil
			}
			logger.Errorf("SERVICE_CHAT", "Error obteniendo mensaje ancla %s: %v", beforeMessageID, err)
			return nil, fmt.Errorf("error con paginación: %w", err)
		}

		// Se agrega la condición de paginación.
		query += " AND (SentAt < ? OR (SentAt = ? AND Id < ?))"
		args = append(args, anchorDate, anchorDate, anchorID)
	}

	query += " ORDER BY SentAt DESC, Id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := chatDB.Query(query, args...)
	if err != nil {
		logger.Errorf("SERVICE_CHAT", "Error consultando historial de mensajes para ChatID %s: %v", chatID, err)
		return nil, fmt.Errorf("error al obtener mensajes: %w", err)
	}
	defer rows.Close()

	var messages []wsmodels.MessageDB
	for rows.Next() {
		var m wsmodels.MessageDB
		var content, mediaId, replyToMessageId, chatIdGroup sql.NullString
		var editedAt sql.NullTime
		var sentAt time.Time

		err := rows.Scan(
			&m.Id,
			&m.SenderId,
			&content,
			&sentAt,
			&m.Status,
			&m.TypeMessageId,
			&mediaId,
			&replyToMessageId,
			&editedAt,
			&chatIdGroup,
		)
		if err != nil {
			logger.Errorf("SERVICE_CHAT", "Error escaneando mensaje: %v", err)
			continue
		}

		// Asignar chatID a la estructura utilizando un puntero.
		m.ChatId = new(string)
		*m.ChatId = chatID

		if content.Valid {
			m.Content = &content.String
		}
		if mediaId.Valid {
			m.MediaId = &mediaId.String
		}
		if replyToMessageId.Valid {
			m.ReplyToMessageId = &replyToMessageId.String
		}
		if chatIdGroup.Valid {
			m.ChatIdGroup = &chatIdGroup.String
		}

		// Formateo de los timestamps a ISO8601.
		m.SentAt = sentAt.UTC().Format(time.RFC3339Nano)
		if editedAt.Valid {
			editedAtStr := editedAt.Time.UTC().Format(time.RFC3339Nano)
			m.EditedAt = &editedAtStr
		}

		messages = append(messages, m)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf("SERVICE_CHAT", "Error después de iterar mensajes: %v", err)
		return nil, fmt.Errorf("error procesando resultados de mensajes: %w", err)
	}

	logger.Successf("SERVICE_CHAT", "Historial para ChatID %s recuperado. %d mensajes.", chatID, len(messages))
	return messages, nil
}

// GetChatParticipants recupera los IDs de los dos participantes de un chat.
// Retorna user1ID, user2ID, error.
func GetChatParticipants(chatID string) (int64, int64, error) {
	if chatDB == nil {
		return 0, 0, errors.New("GetChatParticipants: servicio de chat no inicializado con conexión a BD")
	}
	contact, err := queries.GetContactByChatID(chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warnf("SERVICE_CHAT", "GetChatParticipants: No se encontró contacto para ChatID %s", chatID)
			return 0, 0, fmt.Errorf("no se encontró chat con ID %s: %w", chatID, err)
		}
		logger.Errorf("SERVICE_CHAT", "GetChatParticipants: Error obteniendo contacto para ChatID %s: %v", chatID, err)
		return 0, 0, fmt.Errorf("error obteniendo datos del chat %s: %w", chatID, err)
	}
	// Asumiendo que la estructura de contact tiene User1Id y User2Id
	return contact.User1Id, contact.User2Id, nil
}

func MarkMessageAsRead(userID int64, messageID string, manager *customws.ConnectionManager[wsmodels.WsUserData]) (int64, error) {
	if chatDB == nil {
		return 0, errors.New("servicio de chat no inicializado")
	}

	// 1. Obtener el SenderId del mensaje para saber a quién notificar.
	var senderID int64
	var currentStatus string
	queryGet := `SELECT SenderId, Status FROM Message WHERE Id = ?`
	err := chatDB.QueryRow(queryGet, messageID).Scan(&senderID, &currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("mensaje con ID %s no encontrado", messageID)
		}
		return 0, fmt.Errorf("error obteniendo el remitente del mensaje %s: %w", messageID, err)
	}

	// 2. Actualizar el estado del mensaje a 'read' solo si no lo está ya.
	// Esto evita notificaciones y escrituras innecesarias.
	if currentStatus == "read" {
		return senderID, nil // El mensaje ya está leído, no hacer nada más.
	}

	queryUpdate := `UPDATE Message SET Status = 'read' WHERE Id = ?`
	result, err := chatDB.Exec(queryUpdate, messageID)
	if err != nil {
		return 0, fmt.Errorf("error actualizando el estado del mensaje %s a 'read': %w", messageID, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Esto podría ocurrir si el mensaje fue eliminado justo después de leerlo.
		return 0, fmt.Errorf("no se actualizó ninguna fila para el mensaje ID %s (puede que no exista)", messageID)
	}

	// 3. Devolver el ID del remitente para que el handler pueda notificarle.
	return senderID, nil
}

// TODO: Implementar GetMessagesForChat, MarkMessagesAsRead, SetUserTypingStatus
