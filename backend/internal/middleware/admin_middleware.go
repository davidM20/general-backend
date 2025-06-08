package middleware

import (
	"database/sql"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// AdminMiddleware verifica si el usuario tiene rol de admin y si su sesión es válida.
// Debe usarse DESPUÉS de AuthMiddleware.
func AdminMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Obtener datos del contexto (inyectados por AuthMiddleware)
			userID, ok := r.Context().Value(UserIDContextKey).(int64)
			if !ok {
				logger.Warn("ADMIN_AUTH", "AdminMiddleware: UserID no encontrado en el contexto")
				http.Error(w, "Acceso no autorizado", http.StatusForbidden)
				return
			}

			roleID, ok := r.Context().Value(RoleIDContextKey).(int64)
			if !ok {
				logger.Warn("ADMIN_AUTH", "AdminMiddleware: RoleID no encontrado en el contexto")
				http.Error(w, "Acceso no autorizado", http.StatusForbidden)
				return
			}

			// 2. Verificar el rol de administrador
			if roleID != int64(models.RoleAdmin) {
				logger.Warnf("ADMIN_AUTH", "Intento de acceso de no-administrador (UserID: %d, RoleID: %d)", userID, roleID)
				http.Error(w, "Acceso prohibido: se requiere rol de administrador", http.StatusForbidden)
				return
			}

			// 3. Verificar si la sesión del token es válida en la base de datos
			token := r.Header.Get("Authorization")[7:] // Extraer token del header "Bearer <token>"
			valid, err := queries.IsSessionValid(db, userID, token)
			if err != nil {
				// El error ya fue logueado dentro de IsSessionValid
				http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
				return
			}
			if !valid {
				logger.Warnf("ADMIN_AUTH", "Intento de acceso con token inválido o sesión cerrada (UserID: %d)", userID)
				http.Error(w, "Sesión inválida o expirada", http.StatusUnauthorized)
				return
			}

			logger.Infof("ADMIN_AUTH", "Acceso de administrador verificado y válido para UserID: %d", userID)
			next.ServeHTTP(w, r)
		})
	}
}
