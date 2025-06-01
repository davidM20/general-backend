package websocket

import (
	"errors"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/admin"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/handlers"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// handleUnsupportedResource maneja el error de recurso no soportado
func handleUnsupportedResource(conn *customws.Connection[wsmodels.WsUserData], pid, action, resource string) error {
	errMsg := fmt.Sprintf("Recurso '%s' no soportado para la acción '%s'", resource, action)
	logger.Warn("ROUTER_DATA", errMsg)
	conn.SendErrorNotification(pid, 400, errMsg)
	return errors.New(errMsg)
}

// ProcessClientMessage enruta los mensajes del cliente a los handlers apropiados
func ProcessClientMessage(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Debugf("ROUTER", "Mensaje recibido de UserID %d: Tipo '%s', PID '%s'",
		conn.ID, msg.Type, msg.PID)

	// Registrar métricas
	collector := admin.GetCollector()
	if collector != nil {
		collector.RecordMessage(string(msg.Type))
	}

	var err error

	switch msg.Type {
	// --- Solicitud de datos genérica ---
	case types.MessageTypeDataRequest:
		err = HandleDataRequest(conn, msg)

	// --- Chat ---
	case types.MessageTypeGetChatList:
		err = handlers.HandleGetChatList(conn, msg)
	case types.MessageTypeChatHistory:
		err = handlers.HandleGetChatHistory(conn, msg)
	case types.MessageTypeSendChatMessage:
		err = handlers.HandleSendChatMessage(conn, msg)

	// --- Notificaciones ---
	case types.MessageTypeGetNotifications:
		err = handlers.HandleGetNotifications(conn, msg)
	case types.MessageTypeMarkNotificationRead:
		err = handlers.HandleMarkNotificationRead(conn, msg)

	// --- Contactos ---
	case types.MessageTypeAcceptFriendRequest:
		err = handlers.HandleAcceptFriendRequest(conn, msg)
	case types.MessageTypeRejectFriendRequest:
		err = handlers.HandleRejectFriendRequest(conn, msg)

	// --- Perfil ---
	case types.MessageTypeGetMyProfile:
		err = handlers.HandleGetMyProfile(conn, msg)
	case types.MessageTypeGetUserProfile:
		err = handlers.HandleGetUserProfile(conn, msg)

	default:
		warnMsg := fmt.Sprintf("Tipo de mensaje no soportado: '%s'", msg.Type)
		logger.Warn("ROUTER", warnMsg)
		err = errors.New(warnMsg)
	}

	// Registrar error si ocurrió
	if err != nil && collector != nil {
		collector.RecordError(string(msg.Type) + "_error")
	}

	return err
}
