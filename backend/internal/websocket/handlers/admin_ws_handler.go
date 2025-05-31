package handlers

import (
	"fmt"

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

	// Aquí iría la lógica para obtener los datos del dashboard:
	// 1. Obtener activeUsers (podría ser desde conn.Manager() o admin.GetCollector() si se adapta)
	// 2. Consultar totalRegisteredUsers, administrativeUsers, businessAccounts, alumniStudents desde la BD.
	// 3. Consultar usersByCampus desde la BD.
	// 4. Consultar monthlyActivity desde la BD.
	// 5. Calcular averageUsageTime (esto es más complejo, podría ser un placeholder por ahora).

	// Ejemplo de datos (reemplazar con datos reales de la BD y el sistema)
	dashboardData := wsmodels.DashboardDataPayload{
		ActiveUsers:          10,       // Placeholder
		TotalRegisteredUsers: 100,      // Placeholder
		AdministrativeUsers:  5,        // Placeholder
		BusinessAccounts:     15,       // Placeholder
		AlumniStudents:       70,       // Placeholder
		AverageUsageTime:     "1h 30m", // Placeholder
		UsersByCampus: []wsmodels.UserByCampus{
			{Name: "Campus Central", Users: 50},
			{Name: "Campus Norte", Users: 20},
		},
		MonthlyActivity: wsmodels.MonthlyActivity{
			Labels: []string{"Ene", "Feb", "Mar"},
			Data:   []int64{30, 45, 60},
		},
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
		// No necesariamente retornamos error al cliente aquí, ya que la solicitud original (data_request)
		// podría haber recibido ya un ACK. El error es en el envío asíncrono de los datos.
		return fmt.Errorf("error enviando datos del dashboard: %w", err)
	}

	logger.Successf(adminWsHandlerLogComponent, "Datos del dashboard enviados a UserID %d", conn.ID)
	// El `data_request` original debería ser ack'd por `handelDataRequest.go` si el despacho fue exitoso.
	// Este handler solo envía los datos solicitados.
	return nil
}
