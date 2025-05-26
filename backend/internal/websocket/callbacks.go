package websocket

import (
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// Este archivo contendrá la implementación de los Callbacks de customws.

// OnConnect se llama cuando un nuevo cliente establece una conexión exitosa.
func OnConnect(conn *customws.Connection[wsmodels.WsUserData]) error {
	logger.Infof("CALLBACK", "User connected via WebSocket: ID %d, Username %s", conn.ID, conn.UserData.Username)

	// Llamar al servicio para manejar la lógica de conexión (actualizar BD, notificar a otros)
	err := services.HandleUserConnect(conn.ID, conn.UserData.Username, conn.Manager())
	if err != nil {
		logger.Errorf("CALLBACK", "Error in OnConnect handler for user %d: %v", conn.ID, err)
		// Decidir si el error es crítico y debe cerrar la conexión.
		// Por ahora, solo lo registramos.
	}

	// Ejemplo de envío de mensaje de bienvenida al usuario conectado
	// welcomeMsg := types.ServerToClientMessage{
	// 	PID:     conn.Manager.Callbacks().GeneratePID(), // Asegúrate que GeneratePID no sea nil en la config de customws
	// 	Type:    "welcome_message", // Define este tipo de mensaje
	// 	Payload: map[string]interface{}{
	// 		"message": "Bienvenido al servidor WebSocket, " + conn.UserData.Username + "!",
	// 		"userId":  conn.ID,
	// 	},
	// }
	// if err := conn.SendMessage(welcomeMsg); err != nil {
	// 	logger.Warnf("CALLBACK", "Failed to send welcome message to user %d: %v", conn.ID, err)
	// }

	return nil
}

// OnDisconnect se llama cuando un cliente se desconecta.
func OnDisconnect(conn *customws.Connection[wsmodels.WsUserData], discErr error) {
	if discErr != nil {
		logger.Warnf("CALLBACK", "User disconnected with error: ID %d, Username %s. Error: %v", conn.ID, conn.UserData.Username, discErr)
	} else {
		logger.Infof("CALLBACK", "User disconnected gracefully: ID %d, Username %s", conn.ID, conn.UserData.Username)
	}

	// Llamar al servicio para manejar la lógica de desconexión
	services.HandleUserDisconnect(conn.ID, conn.UserData.Username, conn.Manager(), discErr)
}
