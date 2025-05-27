package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

var presenceDB *sql.DB // Usar una variable de BD específica para este servicio si es necesario o usar una global del paquete

// InitializePresenceService permite inyectar la dependencia de la base de datos.
// Esta función debería ser llamada desde main.go si este servicio requiere su propia inicialización.
// Si comparte la BD con ChatService, podemos usar chatDB o una variable de BD más genérica para servicios.
// Por ahora, asumiremos que puede usar la misma BD que ChatService, o que se inicializará una similar.
// Para evitar conflictos, la nombraré presenceDB y esperaré que se inicialice.
// O mejor aún, pasamos el db a las funciones Handle directamente si no hay estado en el servicio.
// Por consistencia con InitializeChatService, haré una función de inicialización.

func InitializePresenceService(database *sql.DB) {
	presenceDB = database
	logger.Info("SERVICE_PRESENCE", "PresenceService inicializado con conexión a BD.")
}

// HandleUserConnect se llama cuando un usuario se conecta.
// Debería actualizar el estado del usuario a 'online' en la base de datos
// y potencialmente notificar a los contactos del usuario.
func HandleUserConnect(userID int64, username string, manager *customws.ConnectionManager[wsmodels.WsUserData]) error {
	if presenceDB == nil {
		logger.Error("SERVICE_PRESENCE", "PresenceService no inicializado con conexión a BD")
		return fmt.Errorf("PresenceService no inicializado")
	}
	logger.Infof("SERVICE_PRESENCE", "User connected: ID %d, Username: %s. Processing presence update.", userID, username)

	err := queries.SetUserOnlineStatus(presenceDB, userID, true)
	if err != nil {
		logger.Errorf("SERVICE_PRESENCE", "Error actualizando estado online para UserID %d: %v", userID, err)
		// No devolver el error necesariamente, ya que la conexión WS ya está establecida.
		// Pero es importante registrarlo.
	}

	contactUserIDs, err := queries.GetUserContactIDs(presenceDB, userID)
	if err != nil {
		logger.Errorf("SERVICE_PRESENCE", "Error obteniendo IDs de contacto para UserID %d: %v", userID, err)
		// Continuar incluso si no podemos notificar a los contactos
	} else if len(contactUserIDs) > 0 {
		presenceMsg := types.ServerToClientMessage{
			PID:        manager.Callbacks().GeneratePID(),
			Type:       types.MessageTypePresenceEvent,
			FromUserID: userID,
			Payload: map[string]interface{}{
				"eventType": "user_online",
				"userId":    userID,
				"username":  username, // Podrías querer enviar más info del usuario aquí
			},
		}
		errsMap := manager.BroadcastToUsers(contactUserIDs, presenceMsg)
		if len(errsMap) > 0 {
			logger.Warnf("SERVICE_PRESENCE", "Errores difundiendo estado online para UserID %d a sus contactos: %v", userID, errsMap)
		}
	}

	logger.Successf("SERVICE_PRESENCE", "Actualización de presencia para conexión de usuario %d (%s) manejada.", userID, username)
	return nil
}

// HandleUserDisconnect se llama cuando un usuario se desconecta.
// Debería actualizar el estado del usuario a 'offline' en la base de datos
// y potencialmente notificar a los contactos del usuario.
func HandleUserDisconnect(userID int64, username string, manager *customws.ConnectionManager[wsmodels.WsUserData], discErr error) {
	if presenceDB == nil {
		logger.Errorf("SERVICE_PRESENCE", "PresenceService no inicializado con conexión a BD para desconexión de UserID %d", userID)
		// No podemos hacer mucho si la BD no está disponible aquí.
		return
	}
	logger.Infof("SERVICE_PRESENCE", "User disconnected: ID %d, Username: %s. Error (if any): %v. Processing presence update.", userID, username, discErr)

	err := queries.SetUserOnlineStatus(presenceDB, userID, false)
	if err != nil {
		logger.Errorf("SERVICE_PRESENCE", "Error actualizando estado offline para UserID %d: %v", userID, err)
	}

	lastSeenTimestamp := time.Now().UnixMilli()

	contactUserIDs, err := queries.GetUserContactIDs(presenceDB, userID)
	if err != nil {
		logger.Errorf("SERVICE_PRESENCE", "Error obteniendo IDs de contacto para UserID %d al desconectar: %v", userID, err)
	} else if len(contactUserIDs) > 0 {
		presenceMsg := types.ServerToClientMessage{
			PID:        manager.Callbacks().GeneratePID(),
			Type:       types.MessageTypePresenceEvent,
			FromUserID: userID,
			Payload: map[string]interface{}{
				"eventType": "user_offline",
				"userId":    userID,
				"username":  username,
				"lastSeen":  lastSeenTimestamp,
			},
		}
		errsMap := manager.BroadcastToUsers(contactUserIDs, presenceMsg)
		if len(errsMap) > 0 {
			logger.Warnf("SERVICE_PRESENCE", "Errores difundiendo estado offline para UserID %d a sus contactos: %v", userID, errsMap)
		}
	}

	logger.Successf("SERVICE_PRESENCE", "Actualización de presencia para desconexión de usuario %d (%s) manejada. Desconnect error: %v", userID, username, discErr)
}
