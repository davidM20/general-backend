package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	// "github.com/davidM20/micro-service-backend-go.git/internal/auth"
	// "github.com/davidM20/micro-service-backend-go.git/internal/config"
	// "github.com/davidM20/micro-service-backend-go.git/pkg/saveimage" // Importar cuando esté listo
	"log"
)

// MediaHandler maneja las peticiones relacionadas con archivos multimedia
type MediaHandler struct {
	DB *sql.DB
	// Cfg *config.Config // Necesario para GCS config
}

// NewMediaHandler crea una nueva instancia de MediaHandler
func NewMediaHandler(db *sql.DB /*, cfg *config.Config */) *MediaHandler {
	return &MediaHandler{DB: db /*, Cfg: cfg*/}
}

// UploadMedia maneja la subida de archivos multimedia
func (h *MediaHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	// TODO: Proteger con middleware de autenticación.
	/*
	   claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	   if !ok || claims == nil {
	       http.Error(w, "Unauthorized", http.StatusUnauthorized)
	       return
	   }
	   userID := claims.UserID
	*/
	userID := int64(1) // Placeholder - obtener userID real del token

	// TODO: Implementar la lógica completa usando pkg/saveimage
	// 1. Parsear el formulario multipart (r.ParseMultipartForm)
	// 2. Obtener el archivo (r.FormFile("image") o el nombre que uses)
	// 3. Leer bytes del archivo
	// 4. Validar tipo de archivo (filetype.Match)
	// 5. Generar nombre único (UUID + extensión)
	// 6. Redimensionar y subir variantes a GCS (usando pkg/saveimage o pkg/cloudClient)
	// 7. Guardar registro en la tabla Multimedia (MediaId, Type, Ratio, UserId, FileName, ContentId(nombre en GCS), ChatId?)
	// 8. Devolver respuesta JSON (ej. { "mediaId": "...", "url": "..."})

	log.Println("Placeholder para UploadMedia ejecutado por usuario:", userID)

	// Respuesta temporal
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // Usar Accepted ya que el proceso puede ser largo
	json.NewEncoder(w).Encode(map[string]string{"message": "Upload request received (placeholder)"})
}
