package handlers

import (
	"database/sql"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/config"
)

// AdminHandler maneja las peticiones para las rutas de administrador.
type AdminHandler struct {
	DB  *sql.DB
	Cfg *config.Config
}

// NewAdminHandler crea una nueva instancia de AdminHandler.
func NewAdminHandler(db *sql.DB, cfg *config.Config) *AdminHandler {
	return &AdminHandler{
		DB:  db,
		Cfg: cfg,
	}
}

// GetDashboard es un handler de ejemplo para el dashboard de administrador.
func (h *AdminHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Bienvenido al Dashboard de Administrador"))
}
