package websocket

import (
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/admin"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// Este archivo contendrá la implementación de los Callbacks de customws.

// OnConnect se ejecuta cuando un usuario se conecta al WebSocket
func OnConnect(conn *customws.Connection[wsmodels.WsUserData]) error {
	logger.Infof("CONNECTION", "Usuario conectado: ID %d, Username: %s",
		conn.ID, conn.UserData.Username)

	// Registrar conexión en métricas
	collector := admin.GetCollector()
	if collector != nil {
		collector.RecordConnection(conn.ID)
	}

	// Procesar lógica de conexión
	return services.HandleUserConnect(conn.ID, conn.UserData.Username, conn.Manager())
}

// OnDisconnect se ejecuta cuando un usuario se desconecta del WebSocket
func OnDisconnect(conn *customws.Connection[wsmodels.WsUserData], err error) {
	logger.Infof("CONNECTION", "Usuario desconectado: ID %d, Username: %s",
		conn.ID, conn.UserData.Username)
	if err != nil {
		logger.Warnf("CONNECTION", "Desconexión con error para UserID %d: %v", conn.ID, err)
	}

	// Registrar desconexión en métricas
	collector := admin.GetCollector()
	if collector != nil {
		collector.RecordDisconnection(conn.ID)
	}

	// Procesar lógica de desconexión
	services.HandleUserDisconnect(conn.ID, conn.UserData.Username, conn.Manager(), err)
}

// GeneratePID genera un ID único para cada mensaje
func GeneratePID() string {
	return "server-msg-" + time.Now().Format("20060102150405.000000")
}
