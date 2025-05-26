package services

import (
	"database/sql"
	// "encoding/json" // No se usa directamente aquí por ahora
	"errors"
	"fmt"

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

		lastMsg, err := queries.GetLastMessageBetweenUsers(chatDB, userID, otherUserID)
		if err != nil {
			logger.Warnf("SERVICE_CHAT", "Error obteniendo último mensaje entre %d y %d: %v", userID, otherUserID, err)
		}

		var lastMessageText string
		if lastMsg != nil {
			lastMessageText = lastMsg.Text
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
		}

		chatList = append(chatList, chatInfo)
	}

	logger.Successf("SERVICE_CHAT", "Lista de chats recuperada para UserID: %d. Número de chats: %d", userID, len(chatList))
	return chatList, nil
}

// ProcessAndSaveChatMessage procesa un mensaje de chat entrante, lo guarda en la BD,
// y envía el mensaje a destinatarios.
func ProcessAndSaveChatMessage(fromUserID, toUserID int64, text string, manager *customws.ConnectionManager[wsmodels.WsUserData]) (*models.Message, error) {
	logger.Infof("SERVICE_CHAT", "Procesando mensaje de chat de UserID %d a UserID %d", fromUserID, toUserID)

	// Usar la variable global chatDB que ya está inicializada
	if chatDB == nil {
		return nil, errors.New("chat service no inicializado con conexión a BD")
	}

	// Comprobar que ambos usuarios están en contactos aceptados
	// No lo verificamos por ahora, asumimos que la aplicación frontend solo permite esto para contactos válidos.

	// Nota: Si es requerido verificar que fromUserID y toUserID tienen un Contact.Status = 'accepted' entre ellos,
	// la BD podría requerirlo si ChatMessage.ChatID es FK a Contact.ChatId y no es nullable.
	// El DDL muestra ChatMessage.ChatId VARCHAR(255) y Contact.ChatId VARCHAR(255) UNIQUE.
	// Y ChatMessage.ChatId FK a Contact.ChatId.

	// Crear el mensaje usando la nueva función
	msg, err := queries.CreateMessageFromChatParams(chatDB, fromUserID, toUserID, text)
	if err != nil {
		return nil, fmt.Errorf("error creando mensaje en BD: %w", err)
	}

	logger.Successf("SERVICE_CHAT", "Mensaje guardado en BD con ID: %s", msg.Id)

	// Crear payload para websocket (compatible con formato anterior)
	payload := map[string]interface{}{
		"id":         msg.Id,
		"chatId":     msg.ChatId,
		"fromUserId": msg.UserId,
		"toUserId":   toUserID, // Nota: esto se deduce del ChatId pero lo agregamos por compatibilidad
		"content":    msg.Text,
		"createdAt":  msg.Date,
		"statusId":   msg.StatusMessage,
	}

	serverMessage := types.ServerToClientMessage{
		PID:     manager.Callbacks().GeneratePID(),
		Type:    types.MessageTypeNewChatMessage,
		Payload: payload,
	}

	if manager.IsUserOnline(toUserID) {
		errSend := manager.SendMessageToUser(toUserID, serverMessage) // SendMessageToUser devuelve un solo error
		if errSend != nil {                                           // Comprobar si hay un error, no un mapa de errores
			logger.Warnf("SERVICE_CHAT", "Error enviando mensaje a UserID %d: %v", toUserID, errSend)
		}
		logger.Infof("SERVICE_CHAT", "Mensaje (ID: %s) enviado a UserID %d (online)", msg.Id, toUserID)
		// TODO: Actualizar estado a "delivered_to_recipient_device" (StatusID = 2)
	} else {
		logger.Infof("SERVICE_CHAT", "Usuario %d no está online. Mensaje (ID: %s) guardado.", toUserID, msg.Id)
		// TODO: Implementar notificación push si el usuario está offline
	}

	return msg, nil
}

// TODO: Implementar GetMessagesForChat, MarkMessagesAsRead, SetUserTypingStatus
