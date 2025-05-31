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

	// Verificar si el usuario ya está marcado como online en la BD
	currentStatus, err := queries.GetUserOnlineStatus(presenceDB, userID)
	if err != nil {
		logger.Errorf("SERVICE_PRESENCE", "Error verificando estado online para UserID %d: %v", userID, err)
	} else if currentStatus {
		logger.Warnf("SERVICE_PRESENCE", "UserID %d ya estaba marcado como online en la BD", userID)
	}

	// Actualizar estado a online
	err = queries.SetUserOnlineStatus(presenceDB, userID, true)
	if err != nil {
		logger.Errorf("SERVICE_PRESENCE", "Error actualizando estado online para UserID %d: %v", userID, err)
		return fmt.Errorf("error actualizando estado online: %w", err)
	}

	// Notificar a contactos
	contactUserIDs, err := queries.GetUserContactIDs(presenceDB, userID)
	if err != nil {
		logger.Errorf("SERVICE_PRESENCE", "Error obteniendo IDs de contacto para UserID %d: %v", userID, err)
	} else if len(contactUserIDs) > 0 {
		// Filtrar solo los contactos que están conectados
		var onlineContactIDs []int64
		for _, contactID := range contactUserIDs {
			if manager.IsUserOnline(contactID) {
				onlineContactIDs = append(onlineContactIDs, contactID)
			}
		}

		if len(onlineContactIDs) > 0 {
			presenceMsg := types.ServerToClientMessage{
				PID:        manager.Callbacks().GeneratePID(),
				Type:       types.MessageTypePresenceEvent,
				FromUserID: userID,
				Payload: map[string]interface{}{
					"eventType": "user_online",
					"userId":    userID,
					"username":  username,
				},
			}
			errsMap := manager.BroadcastToUsers(onlineContactIDs, presenceMsg)
			if len(errsMap) > 0 {
				logger.Warnf("SERVICE_PRESENCE", "Errores difundiendo estado online para UserID %d a sus contactos conectados: %v", userID, errsMap)
			}
		} else {
			logger.Infof("SERVICE_PRESENCE", "Ningún contacto de UserID %d está conectado para notificar", userID)
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
		return
	}
	logger.Infof("SERVICE_PRESENCE", "User disconnected: ID %d, Username: %s. Error (if any): %v. Processing presence update.", userID, username, discErr)

	// Verificar si el usuario ya está marcado como offline en la BD
	currentStatus, err := queries.GetUserOnlineStatus(presenceDB, userID)
	if err != nil {
		logger.Errorf("SERVICE_PRESENCE", "Error verificando estado online para UserID %d: %v", userID, err)
	} else if !currentStatus {
		logger.Warnf("SERVICE_PRESENCE", "UserID %d ya estaba marcado como offline en la BD", userID)
	}

	// Actualizar estado a offline
	err = queries.SetUserOnlineStatus(presenceDB, userID, false)
	if err != nil {
		logger.Errorf("SERVICE_PRESENCE", "Error actualizando estado offline para UserID %d: %v", userID, err)
	}

	lastSeenTimestamp := time.Now().UnixMilli()

	// Notificar a contactos
	contactUserIDs, err := queries.GetUserContactIDs(presenceDB, userID)
	if err != nil {
		logger.Errorf("SERVICE_PRESENCE", "Error obteniendo IDs de contacto para UserID %d al desconectar: %v", userID, err)
	} else if len(contactUserIDs) > 0 {
		// Filtrar solo los contactos que están conectados
		var onlineContactIDs []int64
		for _, contactID := range contactUserIDs {
			if manager.IsUserOnline(contactID) {
				onlineContactIDs = append(onlineContactIDs, contactID)
			}
		}

		if len(onlineContactIDs) > 0 {
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
			errsMap := manager.BroadcastToUsers(onlineContactIDs, presenceMsg)
			if len(errsMap) > 0 {
				logger.Warnf("SERVICE_PRESENCE", "Errores difundiendo estado offline para UserID %d a sus contactos conectados: %v", userID, errsMap)
			}
		} else {
			logger.Infof("SERVICE_PRESENCE", "Ningún contacto de UserID %d está conectado para notificar su desconexión", userID)
		}
	}

	logger.Successf("SERVICE_PRESENCE", "Actualización de presencia para desconexión de usuario %d (%s) manejada. Desconnect error: %v", userID, username, discErr)
}
