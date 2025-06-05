package handlers

import (
	"database/sql"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	// Importar paquete
)

// MediaHandler maneja las peticiones relacionadas con multimedia.
type MediaHandler struct {
	DB  *sql.DB
	Cfg *config.Config
}

// NewMediaHandler crea una nueva instancia de MediaHandler
func NewMediaHandler(db *sql.DB, cfg *config.Config) *MediaHandler {
	return &MediaHandler{DB: db, Cfg: cfg}
}

// UploadMedia maneja la subida de archivos multimedia.
func (h *MediaHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {

}
