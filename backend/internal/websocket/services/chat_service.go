package services

import (
	"database/sql"
	// "encoding/json" // No se usa directamente aquí por ahora
	"errors"
	"fmt"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models" // Alias para el paquete que contiene ChatInfo
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

	contacts, err := queries.GetAcceptedContacts(chatDB, userID)
	if err != nil {
		logger.Errorf("SERVICE_CHAT", "Error obteniendo contactos para UserID %d: %v", userID, err)
		return nil, fmt.Errorf("error obteniendo contactos: %w", err)
	}

	var chatList []wsmodels.ChatInfo
	for _, contact := range contacts {
		var otherUserID int64
		if contact.User1Id == userID {
			otherUserID = contact.User2Id
		} else {
			otherUserID = contact.User1Id
		}

		otherUserInfo, err := queries.GetUserBaseInfo(chatDB, otherUserID)
		if err != nil {
			logger.Warnf("SERVICE_CHAT", "Error obteniendo info del usuario %d para la lista de chat de %d: %v", otherUserID, userID, err)
			// Podríamos optar por continuar y mostrar el chat sin nombre/foto o saltarlo
			continue // Por ahora, saltamos este chat si no podemos obtener info del otro usuario
		}

		lastMsg, err := queries.GetLastMessageBetweenUsers(chatDB, userID, otherUserID)
		if err != nil {
			logger.Warnf("SERVICE_CHAT", "Error obteniendo último mensaje entre %d y %d: %v", userID, otherUserID, err)
		}

		var lastMessageText string
		var lastMessageTs int64
		var lastMessageFromUserId int64
		if lastMsg != nil {
			lastMessageText = lastMsg.Text
			lastMessageTs = lastMsg.Date.UnixMilli()
			lastMessageFromUserId = lastMsg.UserId
		}

		unreadCount, err := queries.GetUnreadMessageCount(chatDB, userID, otherUserID) // Mensajes de otherUserID para userID
		if err != nil {
			logger.Warnf("SERVICE_CHAT", "Error obteniendo contador de no leídos para %d de %d: %v", userID, otherUserID, err)
			// No es un error fatal
		}

		isOnline := manager.IsUserOnline(otherUserID)

		chatInfo := wsmodels.ChatInfo{
			ChatID:         contact.ChatId, // Usar el ChatID del contacto
			OtherUserID:    otherUserID,
			OtherUserName:  otherUserInfo.UserName,
			OtherFirstName: otherUserInfo.FirstName,
			OtherLastName:  otherUserInfo.LastName,
			OtherPicture:   otherUserInfo.Picture,
			IsOtherOnline:  isOnline,
			UnreadCount:    unreadCount,
		}

		if lastMessageText != "" {
			chatInfo.LastMessage = lastMessageText
			chatInfo.LastMessageTs = lastMessageTs
			chatInfo.LastMessageFromUserId = lastMessageFromUserId
		}

		chatList = append(chatList, chatInfo)
	}

	logger.Successf("SERVICE_CHAT", "Lista de chats recuperada para UserID: %d. Número de chats: %d", userID, len(chatList))
	return chatList, nil
}

// ProcessAndSaveChatMessage se encarga de tomar los datos de un mensaje de chat entrante,
// validarlos (parcialmente, el handler puede hacer más), guardarlos en la BD
// y devolver el mensaje guardado para su posterior envío.
// También busca al usuario destino y le envía el mensaje si está en línea.
func ProcessAndSaveChatMessage(userID int64, payload map[string]interface{}, messageID string, manager *customws.ConnectionManager[wsmodels.WsUserData]) (*models.Message, error) {
	if chatDB == nil {
		return nil, errors.New("servicio de chat no inicializado con conexión a BD")
	}
	if manager == nil {
		return nil, errors.New("ConnectionManager no proporcionado a ProcessAndSaveChatMessage")
	}

	// Extraer y validar campos del payload
	chatId, ok := payload["chatId"].(string)
	if !ok || chatId == "" {
		return nil, errors.New("ChatId es requerido y debe ser un string")
	}

	text, _ := payload["text"].(string) // El texto puede ser opcional si hay MediaId
	mediaId, _ := payload["mediaId"].(string)
	responseTo, _ := payload["responseTo"].(string)

	if text == "" && mediaId == "" { // Un mensaje debe tener texto o media
		return nil, errors.New("el mensaje no puede estar vacío, debe contener texto o media")
	}

	// Determinar TypeMessageId basado en si hay MediaId o no.
	var typeMessageID int64 = 1 // Por defecto, texto
	if mediaId != "" {
		typeMessageID = 2 // Asumimos 2 para mensajes con media. Ajusta según tu tabla TypeMessage.
		// Podrías tener una lógica más compleja aquí, por ejemplo, leer el tipo de la tabla Multimedia.
	}

	newMessage := &models.Message{
		Id:            messageID, // Usar el ID generado en el handler
		ChatId:        chatId,
		Text:          text,
		UserId:        userID, // ID del remitente
		Date:          time.Now().UTC(),
		TypeMessageId: typeMessageID,
		StatusMessage: queries.StatusMessageSent, // Estado inicial: Enviado
		MediaId:       mediaId,                   // Puede ser string vacío
		ResponseTo:    responseTo,                // Puede ser string vacío
	}

	// Guardar el mensaje en la base de datos
	createdMsgID, err := queries.CreateMessage(chatDB, newMessage)
	if err != nil {
		logger.Errorf("SERVICE_CHAT", "Error guardando mensaje para UserID %d, ChatID %s: %v", userID, chatId, err)
		return nil, fmt.Errorf("error guardando mensaje en DB: %w", err)
	}

	newMessage.Id = createdMsgID // Asegurarse de que el ID en el struct es el devuelto por la BD

	logger.Infof("SERVICE_CHAT", "Mensaje guardado (ID: %s) para UserID %d en ChatID %s", createdMsgID, userID, chatId)

	// --- Lógica para encontrar destinatario y enviar si está en línea ---

	// 1. Obtener información del chat/contacto para identificar al destinatario
	contact, err := queries.GetContactByChatID(chatDB, chatId)
	if err != nil {
		logger.Errorf("SERVICE_CHAT", "Error obteniendo información del contacto para ChatID %s después de guardar mensaje: %v", chatId, err)
		return newMessage, fmt.Errorf("mensaje guardado pero error obteniendo datos del chat para envío: %w", err)
	}

	// 2. Identificar al usuario destinatario
	var recipientUserID int64
	if userID == contact.User1Id {
		recipientUserID = contact.User2Id
	} else if userID == contact.User2Id {
		recipientUserID = contact.User1Id
	} else {
		logger.Errorf("SERVICE_CHAT", "El remitente del mensaje (UserID %d) no coincide con los participantes del ContactID %s (User1: %d, User2: %d)", userID, contact.ContactId, contact.User1Id, contact.User2Id)
		return newMessage, fmt.Errorf("mensaje guardado pero remitente no coincide con participantes del chat")
	}

	// 3. Verificar si el destinatario está en línea
	isRecipientOnline := manager.IsUserOnline(recipientUserID)
	logger.Infof("SERVICE_CHAT", "Destinatario UserID %d para ChatID %s está en línea: %v", recipientUserID, chatId, isRecipientOnline)

	// 4. Si está en línea, enviar el mensaje
	if isRecipientOnline {
		messageToSend := wsmodels.MessageDB{
			Id:            newMessage.Id,
			ChatId:        newMessage.ChatId,
			FromUserId:    newMessage.UserId,
			TargetUserId:  recipientUserID,
			Text:          newMessage.Text,
			Timestamp:     newMessage.Date.UTC().Format(time.RFC3339Nano),
			Status:        MapStatusMessageToString(newMessage.StatusMessage),
			TypeMessageId: newMessage.TypeMessageId,
			MediaId:       newMessage.MediaId,
			ResponseTo:    newMessage.ResponseTo,
		}

		// MODIFICADO: Usar types.ServerToClientMessage
		serverMessage := customwsTypes.ServerToClientMessage{
			Type:       customwsTypes.MessageTypeNewChatMessage,
			FromUserID: newMessage.UserId, // El remitente original del mensaje
			Payload:    messageToSend,     // messageToSend es wsmodels.MessageDB
		}

		err := manager.SendMessageToUser(recipientUserID, serverMessage) // <--- MODIFICADO: Pasar serverMessage
		if err != nil {
			logger.Errorf("SERVICE_CHAT", "Error enviando mensaje (ID: %s) a UserID %d: %v", newMessage.Id, recipientUserID, err)
		} else {
			logger.Successf("SERVICE_CHAT", "Mensaje (ID: %s) enviado exitosamente a UserID %d", newMessage.Id, recipientUserID)
		}
	} else {
		logger.Infof("SERVICE_CHAT", "Destinatario UserID %d no está en línea, mensaje (ID: %s) guardado pero no enviado inmediatamente.", recipientUserID, newMessage.Id)
	}

	return newMessage, nil
}

// GetChatHistory recupera el historial de mensajes para un chat específico.
// Implementa paginación basada en beforeMessageID y limit.
func GetChatHistory(chatID string, userID int64, limit int, beforeMessageID string, manager *customws.ConnectionManager[wsmodels.WsUserData]) ([]wsmodels.MessageDB, error) {
	if chatDB == nil {
		return nil, errors.New("GetChatHistory: chat service no inicializado con conexión a BD")
	}

	logger.Infof("SERVICE_CHAT", "Recuperando historial para ChatID: %s, UserID: %d, Limit: %d, BeforeMessageID: %s", chatID, userID, limit, beforeMessageID)

	// Obtener participantes del chat para determinar TargetUserId en cada mensaje
	contact, err := queries.GetContactByChatID(chatDB, chatID) // Asumiendo que esta función existe o la creas
	if err != nil {
		logger.Errorf("SERVICE_CHAT", "Error obteniendo información del contacto para ChatID %s: %v", chatID, err)
		return nil, fmt.Errorf("error obteniendo datos del chat: %w", err)
	}

	var args []interface{}
	query := `SELECT Id, UserId, Text, Date, StatusMessage, TypeMessageId, MediaId FROM Message WHERE ChatId = ?`
	args = append(args, chatID)

	if beforeMessageID != "" {
		// Obtener el timestamp y el ID del mensaje ancla para la paginación
		var anchorDate time.Time
		var anchorID string // Asumimos que el ID también se usa para desempatar si los timestamps son idénticos
		// La consulta para el ancla debe ser precisa
		row := chatDB.QueryRow("SELECT Date, Id FROM Message WHERE Id = ? AND ChatId = ?", beforeMessageID, chatID)
		err := row.Scan(&anchorDate, &anchorID)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Warnf("SERVICE_CHAT", "beforeMessageID %s no encontrado para ChatID %s", beforeMessageID, chatID)
				return []wsmodels.MessageDB{}, nil // No hay más mensajes antes de un ID inexistente
			}
			logger.Errorf("SERVICE_CHAT", "Error obteniendo mensaje ancla %s: %v", beforeMessageID, err)
			return nil, fmt.Errorf("error con paginación: %w", err)
		}
		// Para orden DESC (más nuevos primero), queremos mensajes "menores que" el ancla
		// (Date < anchorDate) O (Date == anchorDate AND Id < anchorID)
		// Si los IDs no son directamente comparables para orden, ajustar esta lógica.
		// Asumiendo que los IDs son ULIDs o similar, donde una comparación lexicográfica es válida para la secuencia.
		query += " AND (Date < ? OR (Date = ? AND Id < ?))"
		args = append(args, anchorDate, anchorDate, anchorID)
	}

	query += " ORDER BY Date DESC, Id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := chatDB.Query(query, args...)
	if err != nil {
		logger.Errorf("SERVICE_CHAT", "Error consultando historial de mensajes para ChatID %s: %v", chatID, err)
		return nil, fmt.Errorf("error al obtener mensajes: %w", err)
	}
	defer rows.Close()

	var messages []wsmodels.MessageDB
	for rows.Next() {
		var dbMsg models.Message // Usamos models.Message para el escaneo inicial
		var typeMessageIdSc sql.NullInt64
		var mediaIdSc sql.NullString
		var textSc sql.NullString // Variable para escanear el campo Text
		// ResponseTo y ChatIdGroup no están en el SELECT actual, añadir si es necesario.

		// Los campos a escanear deben coincidir con la consulta SELECT actual:
		// Id, UserId, Text, Date, StatusMessage, TypeMessageId, MediaId
		err := rows.Scan(
			&dbMsg.Id,
			&dbMsg.UserId, // FromUserId
			&textSc,       // Text (puede ser NULL)
			&dbMsg.Date,   // Se convertirá a Timestamp string
			&dbMsg.StatusMessage,
			&typeMessageIdSc, // TypeMessageId (puede ser NULL)
			&mediaIdSc,       // MediaId (puede ser NULL)
		)
		if err != nil {
			logger.Errorf("SERVICE_CHAT", "Error escaneando mensaje: %v", err)
			continue
		}

		var targetUserID int64
		if dbMsg.UserId == contact.User1Id {
			targetUserID = contact.User2Id
		} else if dbMsg.UserId == contact.User2Id { // Asegurar que User1Id y User2Id sean los correctos del contacto
			targetUserID = contact.User1Id
		} else {
			logger.Warnf("SERVICE_CHAT", "El UserId %d del mensaje no coincide con ninguno de los participantes del ContactID %s", dbMsg.UserId, contact.ContactId)
			// Decidir cómo manejar esto: ¿Omitir mensaje? ¿Establecer targetUserID a un valor por defecto o error?
			// Por ahora, lo dejaremos como estaba, pero esto podría ser un problema si los datos son inconsistentes.
			targetUserID = 0 // O alguna otra lógica de manejo de errores
		}

		m := wsmodels.MessageDB{
			Id:           dbMsg.Id,
			ChatId:       chatID,
			FromUserId:   dbMsg.UserId,
			TargetUserId: targetUserID,
			Text:         textSc.String, // Usar textSc.String, será "" si Text era NULL
			Timestamp:    dbMsg.Date.UTC().Format(time.RFC3339Nano),
			Status:       MapStatusMessageToString(dbMsg.StatusMessage),
		}

		if typeMessageIdSc.Valid {
			m.TypeMessageId = typeMessageIdSc.Int64
		}
		if mediaIdSc.Valid {
			m.MediaId = mediaIdSc.String
		}

		messages = append(messages, m)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf("SERVICE_CHAT", "Error después de iterar mensajes: %v", err)
		return nil, fmt.Errorf("error procesando resultados de mensajes: %w", err)
	}

	// Invertir el slice 'messages' para que el más antiguo de la página actual esté primero.
	// for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
	// 	messages[i], messages[j] = messages[j], messages[i]
	// }

	logger.Successf("SERVICE_CHAT", "Historial para ChatID %s recuperado y ordenado (más antiguo primero en la página). %d mensajes.", chatID, len(messages))
	return messages, nil
}

// GetChatParticipants recupera los IDs de los dos participantes de un chat.
// Retorna user1ID, user2ID, error.
func GetChatParticipants(chatID string) (int64, int64, error) {
	if chatDB == nil {
		return 0, 0, errors.New("GetChatParticipants: servicio de chat no inicializado con conexión a BD")
	}
	contact, err := queries.GetContactByChatID(chatDB, chatID)
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

// MapStatusMessageToString convierte el estado int de la BD a una cadena para el cliente.
// Renombrada para ser exportada.
func MapStatusMessageToString(statusInt int) string {
	switch statusInt {
	case 1: // Asumiendo 1 = Enviado
		return "sent"
	case 2: // Asumiendo 2 = Entregado (al dispositivo)
		return "delivered_device"
	case 3: // Asumiendo 3 = Leído
		return "read"
	default:
		logger.Warnf("SERVICE_CHAT", "Estado de mensaje desconocido: %d", statusInt)
		return "unknown"
	}
}

// TODO: Implementar GetMessagesForChat, MarkMessagesAsRead, SetUserTypingStatus
