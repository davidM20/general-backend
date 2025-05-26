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
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
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

		lastMsg, err := queries.GetLastChatMessageBetweenUsers(chatDB, userID, otherUserID)
		if err != nil {
			logger.Warnf("SERVICE_CHAT", "Error obteniendo último mensaje entre %d y %d: %v", userID, otherUserID, err)
			// No es un error fatal, el chat puede no tener mensajes
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

		if lastMsg != nil {
			chatInfo.LastMessage = lastMsg.Content
			chatInfo.LastMessageTs = lastMsg.CreatedAt.UnixMilli()
		}

		chatList = append(chatList, chatInfo)
	}

	logger.Successf("SERVICE_CHAT", "Lista de chats recuperada para UserID: %d. Número de chats: %d", userID, len(chatList))
	return chatList, nil
}

// ProcessAndSaveChatMessage procesa un mensaje de chat entrante, lo guarda en la BD,
// y lo reenvía al destinatario si está conectado.
func ProcessAndSaveChatMessage(fromUserID, toUserID int64, text string, manager *customws.ConnectionManager[wsmodels.WsUserData]) (*models.ChatMessage, error) {
	if chatDB == nil {
		return nil, errors.New("chat service no inicializado con conexión a BD")
	}
	logger.Infof("SERVICE_CHAT", "Procesando mensaje de UserID %d para UserID %d.", fromUserID, toUserID)

	// Determinar el ChatID. Podría basarse en los IDs de los usuarios o recuperarse de la tabla Contact.
	// Por ahora, crearemos un ChatID simple si no existe un contacto directo con ChatID.
	// Idealmente, `GetAcceptedContacts` o una función similar proporcionaría el ChatID para una pareja.
	// Vamos a intentar obtener el contacto para usar su ChatID.
	// Esto es una simplificación; la gestión de ChatID puede ser más compleja.
	var chatID string
	contacts, err := queries.GetAcceptedContacts(chatDB, fromUserID) // Podríamos filtrar para encontrar el específico con toUserID
	if err == nil {
		for _, c := range contacts {
			if (c.User1Id == fromUserID && c.User2Id == toUserID) || (c.User1Id == toUserID && c.User2Id == fromUserID) {
				chatID = c.ChatId
				break
			}
		}
	}
	if chatID == "" {
		// Si no se encuentra un ChatID de contacto, se podría generar uno o manejarlo como un error,
		// dependiendo de la lógica de la aplicación. Para el ejemplo, se dejará vacío, pero
		// la BD podría requerirlo si ChatMessage.ChatID es FK a Contact.ChatId y no es nullable.
		// El DDL muestra ChatMessage.ChatId VARCHAR(255) y Contact.ChatId VARCHAR(255) UNIQUE.
		// Y ChatMessage.ChatId FK a Contact.ChatId.
		// Por lo tanto, un ChatID válido de un Contacto existente es necesario.
		logger.Errorf("SERVICE_CHAT", "No se encontró ChatID para la conversación entre %d y %d.", fromUserID, toUserID)
		return nil, fmt.Errorf("no se pudo determinar el ChatID para la conversación entre %d y %d. Asegúrese de que sean contactos.", fromUserID, toUserID)
	}

	chatMsg := models.ChatMessage{
		// ID se establecerá después de la inserción si es AUTO_INCREMENT, o ya está (UUID).
		ChatID:     chatID,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Content:    text,
		CreatedAt:  time.Now().UTC(),
		StatusID:   1,
	}

	if err := queries.CreateChatMessage(chatDB, &chatMsg); err != nil {
		logger.Errorf("SERVICE_CHAT", "Error guardando mensaje en BD de %d a %d: %v", fromUserID, toUserID, err)
		return nil, err
	}

	logger.Successf("SERVICE_CHAT", "Mensaje de %d a %d guardado en BD. ID: %d, ChatID: %s", fromUserID, toUserID, chatMsg.ID, chatMsg.ChatID)

	fromUserConn, fromUserExists := manager.GetConnection(fromUserID)
	var fromUsername = "UsuarioDesconocido"
	if fromUserExists {
		fromUsername = fromUserConn.UserData.Username
	} else {
		// Como fallback, intentar cargar desde BD si el usuario no está conectado (ej. para notificaciones push futuras)
		userInfo, dbErr := queries.GetUserBaseInfo(chatDB, fromUserID)
		if dbErr == nil && userInfo != nil {
			fromUsername = userInfo.UserName
		} else {
			logger.Warnf("SERVICE_CHAT", "No se pudo obtener el nombre de usuario del remitente (ID: %d) desde la conexión WS activa ni BD. Error BD: %v", fromUserID, dbErr)
		}
	}

	messageToRecipientPayload := map[string]interface{}{
		"id":           fmt.Sprintf("%d", chatMsg.ID), // ChatMessage.ID
		"chatId":       chatMsg.ChatID,
		"fromUserId":   fromUserID,
		"fromUsername": fromUsername,
		"toUserId":     toUserID,
		"text":         text,
		"timestamp":    chatMsg.CreatedAt.UnixMilli(),
	}

	serverMessage := types.ServerToClientMessage{
		PID:     manager.Callbacks().GeneratePID(),
		Type:    types.MessageTypeNewChatMessage,
		Payload: messageToRecipientPayload,
	}

	if manager.IsUserOnline(toUserID) {
		errSend := manager.SendMessageToUser(toUserID, serverMessage) // SendMessageToUser devuelve un solo error
		if errSend != nil {                                           // Comprobar si hay un error, no un mapa de errores
			logger.Warnf("SERVICE_CHAT", "Error enviando mensaje a UserID %d: %v", toUserID, errSend)
		}
		logger.Infof("SERVICE_CHAT", "Mensaje (ID: %d) enviado a UserID %d (online)", chatMsg.ID, toUserID)
		// TODO: Actualizar estado a "delivered_to_recipient_device" (StatusID = 2)
	} else {
		logger.Infof("SERVICE_CHAT", "Usuario %d no está online. Mensaje (ID: %d) guardado.", toUserID, chatMsg.ID)
		// TODO: Implementar notificación push si el usuario está offline
	}

	return &chatMsg, nil
}

// TODO: Implementar GetMessagesForChat, MarkMessagesAsRead, SetUserTypingStatus
