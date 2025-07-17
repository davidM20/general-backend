package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"
	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// Este archivo contendrá la lógica de autenticación para las conexiones WebSocket.

// Authenticator maneja la autenticación de las solicitudes WebSocket.
// Puede contener dependencias como una conexión a la base de datos.
type Authenticator struct {
	db  *sql.DB
	cfg *config.Config // Añadir la configuración para acceder al secreto del JWT
}

// NewAuthenticator crea una nueva instancia de Authenticator.
func NewAuthenticator(db *sql.DB, cfg *config.Config) *Authenticator {
	return &Authenticator{db: db, cfg: cfg}
}

// AuthenticateAndGetUserData es el callback para customws.
// Valida la petición (ej. token JWT, cookies) y retorna el ID del usuario (int64) y los datos WsUserData.
// Si la autenticación falla, debe retornar un error y ServeHTTP responderá con HTTP Unauthorized.
func (a *Authenticator) AuthenticateAndGetUserData(r *http.Request) (userID int64, userData wsmodels.WsUserData, err error) {
	var token string

	// Lógica de autenticación mejorada - múltiples métodos:
	// 1. Extraer token del header "Authorization: Bearer <token>"
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) == 2 {
			token = splitToken[1]
		} else {
			logger.Warn("AUTH", "Formato de Authorization header inválido")
		}
	}

	// 2. Si no hay token en header, intentar parámetro de URL (para WebSocket desde navegador/React Native)
	if token == "" {
		token = r.URL.Query().Get("token")
	}

	// 3. Si aún no hay token, fallar
	if token == "" {
		logger.Warn("AUTH", "Intento de conexión WS sin token de autorización (header Authorization o parámetro ?token)")
		return 0, wsmodels.WsUserData{}, errors.New("token de autorización requerido")
	}

	// 2. Validar el token JWT
	claims, err := auth.ValidateJWT(token, []byte(a.cfg.JwtSecret))
	if err != nil {
		logger.Warnf("AUTH", "Token JWT inválido para WS: %v", err)
		return 0, wsmodels.WsUserData{}, errors.New("token inválido o expirado")
	}

	// 3. Si el token es válido, obtener datos adicionales del usuario desde la BD
	user, err := queries.GetUserByID(a.db, claims.UserID) // Necesitarás crear esta función
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warnf("AUTH", "Usuario del token JWT no encontrado en BD: UserID %d", claims.UserID)
			return 0, wsmodels.WsUserData{}, errors.New("usuario no encontrado")
		}
		logger.Errorf("AUTH", "Error al obtener datos del usuario desde BD para WS: %v", err)
		return 0, wsmodels.WsUserData{}, errors.New("error interno al verificar usuario")
	}

	// 4. Construir y devolver WsUserData
	logger.Infof("AUTH", "Usuario autenticado exitosamente para WS: ID %d, Username %s",
		user.Id, user.UserName)

	return user.Id, wsmodels.WsUserData{
		UserID:   user.Id,
		Username: user.UserName,
		RoleId:   user.RoleId,
	}, nil
}
