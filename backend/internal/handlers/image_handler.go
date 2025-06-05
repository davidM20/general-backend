package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/middleware"
	"github.com/davidM20/micro-service-backend-go.git/internal/services"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

/*
 * ===================================================
 * HANDLER PARA LA SUBIDA DE IMÁGENES
 * ===================================================
 *
 * Este handler gestiona las solicitudes HTTP para subir imágenes.
 * Extrae el archivo de la solicitud, obtiene el ID de usuario autenticado
 * y llama al ImageUploadService para procesar y guardar la imagen.
 */

// ImageHandler maneja las solicitudes de subida de imágenes.
type ImageHandler struct {
	imageService *services.ImageUploadService
	// cfg *config.Config // Configuración general si es necesaria directamente en el handler
}

// NewImageHandler crea una nueva instancia de ImageHandler.
func NewImageHandler(imageService *services.ImageUploadService) *ImageHandler {
	return &ImageHandler{imageService: imageService}
}

// UploadImage es el método que maneja la petición POST para subir una imagen.
func (h *ImageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Obtener userID del contexto (inyectado por AuthMiddleware)
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok || userID == 0 {
		logger.Warn("UploadImage.Auth", "No se pudo obtener userID del contexto o es inválido.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Usuario no autenticado o ID de usuario inválido."})
		return
	}

	// Parsear el formulario multipart, limitando el tamaño total a, por ejemplo, 10MB.
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logger.Errorf("UploadImage.ParseForm", "Error parseando multipart form: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Solicitud inválida: " + err.Error()})
		return
	}

	file, handler, err := r.FormFile("image") // "image" es el nombre del campo en el form-data
	if err != nil {
		logger.Errorf("UploadImage.FormFile", "Error obteniendo el archivo 'image' del formulario: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al recibir el archivo: " + err.Error()})
		return
	}
	defer file.Close()

	logger.Infof("UploadImage", "Recibida solicitud de subida de imagen del usuario %d, archivo: %s, tamaño: %d", userID, handler.Filename, handler.Size)

	uploadDetails, err := h.imageService.ProcessAndUploadImage(r.Context(), userID, file, handler)
	if err != nil {
		logger.Errorf("UploadImage.ServiceCall", "Error procesando la imagen para el usuario %d: %v", userID, err)
		w.Header().Set("Content-Type", "application/json")
		// Determinar el código de estado basado en el tipo de error podría ser más granular aquí
		w.WriteHeader(http.StatusInternalServerError) // Podría ser BadRequest dependiendo del error del servicio
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al procesar la imagen: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(uploadDetails)
}
