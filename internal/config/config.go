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
	viper.SetConfigName(".env")   // Nombre del archivo de configuración (sin extensión)
	viper.SetConfigType("env")    // Tipo del archivo de configuración
	viper.AddConfigPath(".")      // Ruta para buscar el archivo de configuración (directorio actual)
	viper.AddConfigPath("../")    // También buscar en el directorio padre (útil si se ejecuta desde cmd/*)
	viper.AddConfigPath("../../") // Y dos niveles arriba

	viper.AutomaticEnv() // Leer también variables de entorno

	// Establecer valores por defecto	viper.SetDefault("API_PORT", "8080")
	viper.SetDefault("WS_PORT", "8081")
	viper.SetDefault("PROXY_PORT", "8000")
	viper.SetDefault("JWT_SECRET", "un-secreto-muy-seguro-cambiar-en-produccion") // ¡CAMBIAR ESTO!

	// Intentar leer el archivo de configuración (opcional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Archivo de configuración no encontrado; ignorar y usar env vars/defaults
			fmt.Println("Config file (.env) not found. Using environment variables and defaults.")
		} else {
			// Otro error al leer el archivo de configuración
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
		if dbUser != "" && dbPassword != "" && dbHost != "" && dbPort != "" && dbName != "" {
			cfg.DatabaseDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
				dbUser, dbPassword, dbHost, dbPort, dbName)
		} else {
			return nil, fmt.Errorf("database DSN (DB_DSN or individual DB_ vars) is required")
		}
	}

	if cfg.GCSBucketName == "" {
		fmt.Println("Warning: GCS_BUCKET_NAME is not set. File uploads will fail.")
	}
	// La clave de cuenta de servicio es opcional si se usa ADC (Application Default Credentials)
	// if cfg.GCSServiceAccountKey == "" {
	//     fmt.Println("Warning: GCS_SERVICE_ACCOUNT_KEY_PATH is not set. Assuming Application Default Credentials.")
	// }

	return &cfg, nil
}
