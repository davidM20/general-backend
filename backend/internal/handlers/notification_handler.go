package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/davidM20/micro-service-backend-go.git/internal/middleware"
	"github.com/davidM20/micro-service-backend-go.git/internal/services"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/gorilla/mux"
)

// NotificationHandler maneja las peticiones HTTP relacionadas con notificaciones.
type NotificationHandler struct {
	Service *services.NotificationService
}

// NewNotificationHandler crea una nueva instancia de NotificationHandler.
func NewNotificationHandler(db *sql.DB) *NotificationHandler {
	notificationService := services.NewNotificationService(db)
	return &NotificationHandler{
		Service: notificationService,
	}
}

// MarkAsRead maneja la solicitud para marcar una notificación como leída.
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	// Obtener el UserID del usuario autenticado desde el contexto.
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok {
		logger.Warn("NOTIFICATION_HANDLER", "MarkAsRead: UserID no encontrado en el contexto")
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
		return
	}

	// Extraer el ID de la notificación de los parámetros de la URL.
	vars := mux.Vars(r)
	notificationID, err := strconv.ParseInt(vars["notificationID"], 10, 64)
	if err != nil {
		logger.Warnf("NOTIFICATION_HANDLER", "MarkAsRead: ID de notificación inválido: %v", err)
		http.Error(w, "ID de notificación inválido", http.StatusBadRequest)
		return
	}

	// Llamar al servicio para marcar la notificación como leída.
	err = h.Service.MarkAsRead(notificationID, userID)
	if err != nil {
		// El servicio devuelve un error específico si la notificación no se encuentra
		// o el usuario no es el propietario.
		if err.Error() == "notificación no encontrada o no tienes permiso para modificarla" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			// Para cualquier otro error, devolvemos un error de servidor.
			http.Error(w, "Error al marcar la notificación como leída", http.StatusInternalServerError)
		}
		// El error ya fue logueado en la capa de servicio.
		return
	}

	// Si todo va bien, devolvemos un estado 204 No Content, que es apropiado
	// para una operación de actualización exitosa que no necesita devolver datos.
	w.WriteHeader(http.StatusNoContent)
}
