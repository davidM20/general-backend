package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// Este archivo contendrá la lógica de autenticación para las conexiones WebSocket.

// Authenticator maneja la autenticación de las solicitudes WebSocket.
// Puede contener dependencias como una conexión a la base de datos.
type Authenticator struct {
	db *sql.DB
	// Aquí podrías añadir otras dependencias, como un validador de JWT, etc.
}

// NewAuthenticator crea una nueva instancia de Authenticator.
func NewAuthenticator(db *sql.DB) *Authenticator {
	return &Authenticator{db: db}
}

// AuthenticateAndGetUserData es el callback para customws.
// Valida la petición (ej. token JWT, cookies) y retorna el ID del usuario (int64) y los datos WsUserData.
// Si la autenticación falla, debe retornar un error y ServeHTTP responderá con HTTP Unauthorized.
func (a *Authenticator) AuthenticateAndGetUserData(r *http.Request) (userID int64, userData wsmodels.WsUserData, err error) {
	// Lógica de autenticación de ejemplo:
	// 1. Extraer token (ej. del header "Authorization: Bearer <token>")
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		logger.Warn("AUTH", "Intento de conexión WS sin Authorization header")
		return 0, wsmodels.WsUserData{}, errors.New("authorization header requerido")
	}

	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		logger.Warn("AUTH", "Formato de Authorization header inválido")
		return 0, wsmodels.WsUserData{}, errors.New("formato de token inválido")
	}
	token := splitToken[1]

	// 2. Validar el token
	// Aquí iría la lógica para validar el token JWT o de sesión.
	// Esto podría implicar verificar la firma, la expiración, y consultar la BD si es un token de sesión.
	// Por ahora, simularemos una validación simple.

	// Ejemplo: Si el token es "valid-token-for-user-1", se autentica como UserID 1.
	// En una implementación real, esto consultaría la tabla Session o validaría un JWT.
	user, err := queries.GetUserBySessionToken(a.db, token) // Asumimos que tienes esta función en tu paquete db/queries
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warnf("AUTH", "Token de sesión inválido o expirado: %s", token)
			return 0, wsmodels.WsUserData{}, errors.New("token inválido o expirado")
		}
		logger.Errorf("AUTH", "Error al validar token de sesión: %v", err)
		return 0, wsmodels.WsUserData{}, errors.New("error interno validando token")
	}

	// 3. Si el token es válido, obtener UserID y cualquier otro dato necesario para WsUserData.
	// Por ejemplo, podrías querer cargar el Username aquí si no está en el token.
	// Supongamos que 'user.Id' y 'user.UserName' vienen de tu estructura 'models.User' recuperada.
	logger.Infof("AUTH", "Usuario autenticado exitosamente para WS: ID %d, Username %s", user.Id, user.UserName)
	return user.Id, wsmodels.WsUserData{
		UserID:   user.Id,
		Username: user.UserName,
	}, nil
}
