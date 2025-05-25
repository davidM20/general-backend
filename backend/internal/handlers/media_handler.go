package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"
	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/davidM20/micro-service-backend-go.git/pkg/saveimage" // Importar paquete
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
	// Obtener UserID del contexto (añadido por el middleware de autenticación)
	userID := r.Context().Value(auth.UserIDKey)
	if userID == nil {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}
	userIDTyped, ok := userID.(int64)
	if !ok {
		http.Error(w, "Invalid User ID type in context", http.StatusInternalServerError)
		return
	}

	logger.Infof("MEDIA", "UploadMedia: Received upload request from UserID: %d", userIDTyped)

	// 1. Parsear el formulario multipart
	// Limitar tamaño de la subida (ej. 10 MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logger.Errorf("MEDIA", "UploadMedia Error (UserID: %d): Failed parsing multipart form: %v", userIDTyped, err)
		http.Error(w, fmt.Sprintf("Error parsing multipart form: %v", err), http.StatusBadRequest)
		return
	}

	// 2. Obtener el archivo (asumimos que el campo se llama 'media')
	file, handler, err := r.FormFile("media")
	if err != nil {
		logger.Errorf("MEDIA", "UploadMedia Error (UserID: %d): Error retrieving 'media' file: %v", userIDTyped, err)
		http.Error(w, "Error retrieving 'media' file from form", http.StatusBadRequest)
		return
	}
	defer file.Close() // Asegurarse de cerrar el archivo

	logger.Infof("MEDIA", "UploadMedia Info (UserID: %d): Processing file: Name=%s, Size=%d, Header=%+v",
		userIDTyped, handler.Filename, handler.Size, handler.Header)

	// 3. Configurar saveimage
	// Determinar tipo de almacenamiento (ej. desde config o por defecto)
	storageCfg := saveimage.Config{
		StorageType: "local",         // Cambiar a "gcs" si se configura GCS
		UploadDir:   "./uploads/api", // Directorio específico para subidas de API
		// GCSBucketName: h.Cfg.GCSBucketName,
		// GCSClient: getGCSClient(), // Necesitarías una función para inicializar y obtener el cliente GCS
	}

	// 4. Llamar a saveimage.SaveImage
	// Aquí podrías definir maxWidth/maxHeight si quieres redimensionar siempre
	// ej. maxWidth=1024, maxHeight=1024
	saveResult, err := saveimage.SaveImage(handler, storageCfg, 0, 0) // 0, 0 para no redimensionar
	if err != nil {
		logger.Errorf("MEDIA", "UploadMedia Error (UserID: %d): Failed saving image: %v", userIDTyped, err)
		http.Error(w, fmt.Sprintf("Failed to save uploaded file: %v", err), http.StatusInternalServerError)
		return
	}

	// 5. Guardar información en la tabla Multimedia
	// Determinar el tipo (image, video, etc.) basado en la extensión o MIME type
	fileType := "image" // Simplificación, mejorar esto
	// ratio := calculateAspectRatio(saveResult.FileURL) // Necesitaría leer la imagen guardada
	ratio := float32(1.0) // Placeholder
	now := time.Now().UTC()

	query := `INSERT INTO Multimedia (Id, Type, Ratio, UserId, FileName, CreateAt, ContentId, ChatId) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = h.DB.Exec(query,
		saveResult.MediaID,
		fileType,
		ratio,
		userIDTyped,
		handler.Filename, // Guardar nombre original?
		now,
		saveResult.FileURL, // Usar la URL/identificador devuelto como ContentId
		"",                 // ChatId vacío, ya que no está asociado a un chat específico en esta subida genérica
	)
	if err != nil {
		logger.Errorf("MEDIA", "UploadMedia DB Error (UserID: %d): Failed inserting multimedia record: %v", userIDTyped, err)
		// Considerar eliminar el archivo subido si la inserción en DB falla?
		http.Error(w, "Failed to record uploaded media information", http.StatusInternalServerError)
		return
	}

	// 6. Devolver respuesta exitosa con MediaID y URL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(saveResult) // Devolver el resultado de SaveImage

	logger.Successf("MEDIA", "UploadMedia Success (UserID: %d): Media uploaded. ID: %s, URL: %s", userIDTyped, saveResult.MediaID, saveResult.FileURL)
}
