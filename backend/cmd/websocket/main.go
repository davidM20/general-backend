package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/db"
	internalWs "github.com/davidM20/micro-service-backend-go.git/internal/websocket"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/admin"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/auth"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/handlers"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar .env (opcional)
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not load .env file. Using environment variables directly.")
	}

	// Cargar configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Conectar a la base de datos
	dbConn, err := db.Connect(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if dbConn != nil {
			dbConn.Close()
			log.Println("Database connection closed.")
		}
	}()

	if err := db.InitializeDatabase(dbConn); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully.")

	// Inicializar servicios que dependen de la BD
	services.InitializeChatService(dbConn)
	services.InitializeNotificationService(dbConn)
	services.InitializeProfileService(dbConn)

	// Inicializar FeedService y FeedHandler
	feedSvc := services.NewFeedService(dbConn) // Crear y asignar la instancia
	handlers.InitializeFeedHandler(feedSvc)    // Pasar la instancia al inicializador del handler
	logger.Info("MAIN", "FeedService y FeedHandler inicializados.")

	// Configurar el paquete customws
	wsConfig := types.DefaultConfig()
	wsConfig.AllowedOrigins = []string{"*"}
	wsConfig.WriteWait = 15 * time.Second
	wsConfig.PongWait = 60 * time.Second
	wsConfig.PingPeriod = (wsConfig.PongWait * 9) / 10
	wsConfig.MaxMessageSize = 4096
	wsConfig.SendChannelBuffer = 256
	wsConfig.AckTimeout = 10 * time.Second
	wsConfig.RequestTimeout = 20 * time.Second

	// Instanciar el authenticator
	wsAuthenticator := auth.NewAuthenticator(dbConn)

	// Configurar callbacks
	callbacks := customws.Callbacks[wsmodels.WsUserData]{
		AuthenticateAndGetUserData: wsAuthenticator.AuthenticateAndGetUserData,
		OnConnect: func(conn *customws.Connection[wsmodels.WsUserData]) error {
			log.Printf("User connected: ID %d, Username %s", conn.ID, conn.UserData.Username)
			// Llamar a OnConnect de callbacks.go
			return internalWs.OnConnect(conn)
		},
		OnDisconnect: func(conn *customws.Connection[wsmodels.WsUserData], err error) {
			// Llamar a OnDisconnect de callbacks.go
			internalWs.OnDisconnect(conn, err)
		},
		ProcessClientMessage: internalWs.ProcessClientMessage,
		GeneratePID: func() string { // Opcional: custom PID generation
			// return uuid.NewString()
			return "server-msg-" + time.Now().Format("20060102150405.000000")
		},
	}

	// Crear el ConnectionManager
	connManager := customws.NewConnectionManager(wsConfig, callbacks)

	// Inicializar PresenceService después de crear el ConnectionManager
	services.InitializePresenceService(dbConn, connManager)

	// Inicializar sistema de administración
	adminUser := os.Getenv("ADMIN_USERNAME")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminUser == "" {
		adminUser = "admin" // valor por defecto
	}
	if adminPass == "" {
		adminPass = "admin123" // valor por defecto
	}

	adminHandler := admin.InitializeAdmin(connManager, dbConn, adminUser, adminPass)
	logger.Infof("MAIN", "Sistema de administración inicializado - Usuario: %s", adminUser)

	// Configurar rutas HTTP
	mux := http.NewServeMux()

	// Ruta principal de WebSocket
	mux.HandleFunc("/ws", connManager.ServeHTTP)

	// Ruta de health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","timestamp":%d}`, time.Now().Unix())
	})

	// Registrar rutas administrativas
	adminHandler.RegisterAdminRoutes(mux)

	serverAddr := cfg.WsPort
	if serverAddr == "" {
		serverAddr = "8082" // Puerto por defecto si no está en la configuración
	}

	srv := &http.Server{
		Addr:         ":" + serverAddr,
		Handler:      mux,
		ReadTimeout:  wsConfig.WriteWait + (5 * time.Second), // Un poco más que el WriteWait de WS
		WriteTimeout: wsConfig.WriteWait + (5 * time.Second),
		IdleTimeout:  wsConfig.PongWait + (10 * time.Second), // Un poco más que el PongWait
	}

	go func() {
		log.Printf("WebSocket Server (using customws) starting on port %s...", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v", serverAddr, err)
		}
	}()

	// Manejo de cierre ordenado
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	<-stopChan // Esperar señal

	log.Println("Shutting down server...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 35*time.Second) // Dar tiempo a las conexiones WS para cerrar
	defer cancelShutdown()

	if err := connManager.Shutdown(shutdownCtx); err != nil {
		log.Printf("CustomWS ConnectionManager shutdown error: %v", err)
	} else {
		log.Println("CustomWS ConnectionManager shutdown complete.")
	}

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		log.Println("HTTP server shutdown complete.")
	}

	log.Println("Server gracefully stopped.")
}
