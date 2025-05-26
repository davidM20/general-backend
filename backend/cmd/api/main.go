package main

import (
	"log"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/db"
	"github.com/davidM20/micro-service-backend-go.git/internal/routes"
	"github.com/davidM20/micro-service-backend-go.git/pkg/cloudclient"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno desde .env (opcional, pero recomendado)
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not load .env file. Using environment variables directly.")
	}

	// Cargar configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Inicializar cliente GCS
	if cfg.GCSBucketName != "" && cfg.GCSServiceAccountKey != "" {
		if err := cloudclient.Open(cfg.GCSBucketName, cfg.GCSServiceAccountKey); err != nil {
			log.Fatalf("Failed to initialize Google Cloud Storage client: %v", err)
		} else {
			log.Println("Google Cloud Storage client initialized successfully.")
		}
	} else {
		log.Println("GCS_BUCKET_NAME or GCS_SERVICE_ACCOUNT_KEY_PATH not set, GCS client not initialized.")
	}

	// Conectar e inicializar la base de datos
	dbConn, err := db.Connect(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	if err := db.InitializeDatabase(dbConn); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Configurar el router principal
	mainRouter := mux.NewRouter()

	// Configurar las rutas de la API
	routes.SetupApiRoutes(mainRouter, dbConn, cfg)

	// CORS manejado por el proxy - no aplicar aquí para evitar duplicación
	httpHandler := mainRouter

	// Configurar servidor HTTP
	serverAddr := cfg.ApiPort
	log.Printf("API Server starting on port %s (CORS handled by proxy)...", serverAddr)

	srv := &http.Server{
		Handler: httpHandler,
		Addr:    ":" + serverAddr,
	}

	// Iniciar servidor
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", serverAddr, err)
	}

	log.Println("API Server stopped.")
}
