package middleware

import (
	"context"
	"net/http"

	"log"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"
	"github.com/davidM20/micro-service-backend-go.git/internal/config" // Para obtener JWT Secret
)

// AuthMiddleware valida el token JWT de las peticiones entrantes desde el parámetro URL 'token'
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Obtener el token del parámetro de consulta 'token'
			tokenString := r.URL.Query().Get("token")
			if tokenString == "" {
				log.Println("AuthMiddleware: Missing 'token' query parameter")
				http.Error(w, "'token' query parameter required", http.StatusUnauthorized)
				return
			}

			// 2. Validar el token JWT
			claims, err := auth.ValidateJWT(tokenString, []byte(cfg.JwtSecret))
			if err != nil {
				log.Printf("AuthMiddleware: Invalid token: %v", err)
				// Puedes ser más específico con los errores si quieres (ej. token expirado)
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Añadir información del usuario (solo el ID) al contexto de la petición
			ctx := context.WithValue(r.Context(), auth.UserIDKey, claims.UserID)

			// Llamar al siguiente handler con el contexto modificado
			log.Printf("AuthMiddleware: User %d authenticated via token param with Role %d", claims.UserID, claims.RoleID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
