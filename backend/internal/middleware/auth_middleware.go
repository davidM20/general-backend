package middleware

import (
	"context"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"
	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// Tipos personalizados para claves de contexto para evitar colisiones
type contextKey string

const (
	UserIDContextKey contextKey = "userID"
	RoleIDContextKey contextKey = "roleID"
)

// AuthMiddleware valida el token JWT de las peticiones entrantes
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			// Intentar obtener el token del encabezado Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				// Formato esperado: "Bearer <token>"
				if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
					token = authHeader[7:]
				}
			}

			// Si no se encuentra en el encabezado, intentar obtenerlo del query parameter
			if token == "" {
				token = r.URL.Query().Get("token")
			}

			// Si no se encuentra el token en ningún lugar
			if token == "" {
				logger.Warn("AUTH", "AuthMiddleware: No token found in Authorization header or query parameter")
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}

			// Validar el token
			claims, err := auth.ValidateJWT(token, []byte(cfg.JwtSecret))
			if err != nil {
				logger.Warnf("AUTH", "AuthMiddleware: Invalid token: %v", err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Agregar información del usuario al contexto usando claves tipadas
			ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
			ctx = context.WithValue(ctx, RoleIDContextKey, int64(claims.RoleID))

			logger.Infof("AUTH", "AuthMiddleware: User %d authenticated with Role %d", claims.UserID, claims.RoleID)

			// Continuar con el siguiente handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
