package config

import (
	"fmt"

	"github.com/spf13/viper" // Usaremos viper para facilitar la gestión de config
)

// Config holds the application configuration
type Config struct {
	DatabaseDSN string `mapstructure:"DB_DSN"`
	ApiPort     string `mapstructure:"API_PORT"`
	WsPort      string `mapstructure:"WS_PORT"`
	ProxyPort   string `mapstructure:"PROXY_PORT"`
	JwtSecret   string `mapstructure:"JWT_SECRET"`
	// TODO: Añadir configuración para Google Cloud Storage (bucket, credentials path, etc.)
	GCSBucketName        string `mapstructure:"GCS_BUCKET_NAME"`
	GCSServiceAccountKey string `mapstructure:"GCS_SERVICE_ACCOUNT_KEY_PATH"` // Ruta al archivo JSON de credenciales

}

// LoadConfig loads configuration from environment variables or a config file.
func LoadConfig() (*Config, error) {
	viper.SetConfigName(".env") // Nombre del archivo de configuración (sin extensión)
	viper.SetConfigType("env")  // Tipo del archivo de configuración

	// Rutas de búsqueda: directorio actual (para tests?), y directorios cmd/*
	viper.AddConfigPath(".")       // Directorio actual (e.g. /internal/config)
	viper.AddConfigPath("cmd/api") // Para ejecución desde la raíz del proyecto
	viper.AddConfigPath("cmd/websocket")
	// Buscar también relativa a donde se ejecuta el binario
	// (Importante cuando se construye y ejecuta desde /cmd/api/ o /cmd/websocket/)
	viper.AddConfigPath(".") // Directorio de ejecución del binario

	// Añadir configuración para buscar automáticamente variables de entorno
	viper.AutomaticEnv()

	// Establecer valores por defecto (opcional, pero recomendado)
	viper.SetDefault("API_PORT", "8080")
	viper.SetDefault("WS_PORT", "8081")
	viper.SetDefault("PROXY_PORT", "8000")
	viper.SetDefault("DB_HOST", "127.0.0.1")
	viper.SetDefault("DB_PORT", "3306")
	viper.SetDefault("JWT_SECRET", "un-secreto-muy-seguro-cambiar-en-produccion") // ¡CAMBIAR ESTO!

	// Intentar leer el archivo de configuración
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Archivo .env no encontrado, no es un error fatal si las variables de entorno están seteadas
			fmt.Println("Warning: .env file not found. Relying on environment variables and defaults.")
		} else {
			// Otro error al leer el archivo
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// Validar configuración esencial
	if cfg.DatabaseDSN == "" {
		// Intentar construir DSN desde variables individuales si DSN completo no está
		dbUser := viper.GetString("DB_USER")
		dbPassword := viper.GetString("DB_PASSWORD")
		dbHost := viper.GetString("DB_HOST")
		dbPort := viper.GetString("DB_PORT")
		dbName := viper.GetString("DB_NAME")

		if dbUser == "" || dbName == "" {
			return nil, fmt.Errorf("DB_USER and DB_NAME are required if DB_DSN is not set")
		}
		if dbHost == "" {
			dbHost = "127.0.0.1" // Default host if not set but others are
		}
		if dbPort == "" {
			dbPort = "3306" // Default port if not set but others are
		}

		cfg.DatabaseDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			dbUser, dbPassword, dbHost, dbPort, dbName)

	} else {
		// Si se proporciona DB_DSN, usarlo directamente
		fmt.Println("Using provided DB_DSN")
	}

	if cfg.JwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	if cfg.GCSBucketName == "" {
		fmt.Println("Warning: GCS_BUCKET_NAME is not set. File uploads will fail if GCS is intended.")
	}

	return &cfg, nil
}
