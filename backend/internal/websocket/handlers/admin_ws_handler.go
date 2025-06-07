package handlers

import (
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/db"
	"github.com/davidM20/micro-service-backend-go.git/internal/services/admin"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	// Asumimos que necesitaremos acceso a la BD, así que importamos el paquete db
	// y el inicializador del servicio de admin para acceder al collector o DB si es necesario.
	// "github.com/davidM20/micro-service-backend-go.git/internal/db"
	// "github.com/davidM20/micro-service-backend-go.git/internal/websocket/admin"
)

const adminWsHandlerLogComponent = "HANDLER_ADMIN_WS"

// HandleGetDashboardInfo maneja la solicitud de información para el dashboard de administración.
func HandleGetDashboardInfo(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof(adminWsHandlerLogComponent, "Solicitud de información del dashboard recibida de UserID %d, PID: %s", conn.ID, msg.PID)

	// Obtener la conexión a la BD
	database := db.GetDB()
	if database == nil {
		err := fmt.Errorf("la conexión a la base de datos no está disponible")
		logger.Errorf(adminWsHandlerLogComponent, "Error en HandleGetDashboardInfo: %v", err)
		// Opcional: enviar un mensaje de error al cliente
		conn.SendErrorNotification(msg.PID, 500, "Error interno del servidor: no se pudo conectar a la base de datos.")
		return err
	}

	// Obtener el número de usuarios activos desde el connection manager
	activeUsers := conn.Manager().GetUserCount()

	// Obtener los datos del dashboard desde el servicio
	dashboardData, err := admin.GetDashboardData(activeUsers)
	if err != nil {
		logger.Errorf(adminWsHandlerLogComponent, "Error obteniendo datos del dashboard: %v", err)
		conn.SendErrorNotification(msg.PID, 500, "Error obteniendo la información del dashboard.")
		return err
	}

	responsePayload := map[string]interface{}{
		"origin":    "get-info-dashboard",
		"dashboard": dashboardData,
	}

	responseMsg := types.ServerToClientMessage{
		Type:       types.MessageTypeDataEvent,
		FromUserID: 0, // 0 indica que es un mensaje del sistema/servidor
		Payload:    responsePayload,
		PID:        conn.Manager().Callbacks().GeneratePID(), // Generar un nuevo PID para la respuesta
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf(adminWsHandlerLogComponent, "Error enviando datos del dashboard a UserID %d: %v", conn.ID, err)
		return fmt.Errorf("error enviando datos del dashboard: %w", err)
	}

	logger.Successf(adminWsHandlerLogComponent, "Datos del dashboard enviados a UserID %d", conn.ID)
	return nil
}
