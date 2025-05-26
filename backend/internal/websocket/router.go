package websocket

import (
	"errors"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/handlers"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// ProcessClientMessage actúa como un enrutador para los mensajes entrantes del cliente.
// Delega el procesamiento a manejadores específicos basados en msg.Type.
func ProcessClientMessage(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Debugf("ROUTER", "Mensaje recibido de UserID %d: Tipo '%s', PID '%s'", conn.ID, msg.Type, msg.PID)

	switch msg.Type {
	// --- Chat ---
	case types.MessageTypeGetChatList:
		return handlers.HandleGetChatList(conn, msg)
	case types.MessageTypeSendChatMessage:
		return handlers.HandleSendChatMessage(conn, msg)
	// case types.MessageTypeMessagesRead:
	// 	return handlers.HandleMessagesRead(conn, msg)
	// case types.MessageTypeTypingIndicatorOn:
	// 	return handlers.HandleTypingIndicatorOn(conn, msg)
	// case types.MessageTypeTypingIndicatorOff:
	// 	return handlers.HandleTypingIndicatorOff(conn, msg)

	// --- Perfil ---
	case types.MessageTypeGetMyProfile:
		return handlers.HandleGetMyProfile(conn, msg)
	// case types.MessageTypeUpdateMyProfile:
	// 	return handlers.HandleUpdateMyProfile(conn, msg)
	case types.MessageTypeGetUserProfile:
		return handlers.HandleGetUserProfile(conn, msg)

	// --- Notificaciones ---
	case types.MessageTypeGetNotifications:
		return handlers.HandleGetNotifications(conn, msg)
	case types.MessageTypeMarkNotificationRead:
		return handlers.HandleMarkNotificationRead(conn, msg)

	default:
		warnMsg := fmt.Sprintf("Tipo de mensaje no soportado recibido de UserID %d: '%s'", conn.ID, msg.Type)
		logger.Warn("ROUTER", warnMsg)
		// Opcional: enviar un error al cliente si el tipo de mensaje es desconocido y se espera una respuesta
		// if msg.PID != "" { // Si el cliente espera una respuesta (tiene PID)
		// 	 conn.SendErrorNotification(msg.PID, 400, "Tipo de mensaje no soportado: "+string(msg.Type))
		// }
		return errors.New(warnMsg) // Devolver error para que customws pueda registrarlo si es necesario
	}
}
