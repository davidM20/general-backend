package middleware

import (
	"context"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"
	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// AuthMiddleware valida el token JWT de las peticiones entrantes desde el parámetro URL 'token'
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Obtener token del query parameter
			token := r.URL.Query().Get("token")
			if token == "" {
				logger.Warn("AUTH", "AuthMiddleware: Missing 'token' query parameter")
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

			// Agregar información del usuario al contexto
			ctx := context.WithValue(r.Context(), "userID", claims.UserID)
			ctx = context.WithValue(ctx, "roleID", claims.RoleID)

			logger.Infof("AUTH", "AuthMiddleware: User %d authenticated via token param with Role %d", claims.UserID, claims.RoleID)

			// Continuar con el siguiente handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
