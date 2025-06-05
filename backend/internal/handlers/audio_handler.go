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
 * HANDLER PARA LA SUBIDA DE ARCHIVOS DE AUDIO
 * ===================================================
 *
 * Este handler gestiona las solicitudes HTTP para subir archivos de audio.
 * Extrae el archivo de la solicitud, obtiene el ID de usuario autenticado
 * y llama al AudioUploadService para procesar y guardar el audio.
 */

// AudioHandler maneja las solicitudes de subida de audio.
type AudioHandler struct {
	audioService *services.AudioUploadService
}

// NewAudioHandler crea una nueva instancia de AudioHandler.
func NewAudioHandler(audioService *services.AudioUploadService) *AudioHandler {
	return &AudioHandler{audioService: audioService}
}

// UploadAudio es el método que maneja la petición POST para subir un archivo de audio.
func (h *AudioHandler) UploadAudio(w http.ResponseWriter, r *http.Request) {
	// Obtener userID del contexto (inyectado por AuthMiddleware)
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok || userID == 0 {
		logger.Warn("UploadAudio.Auth", "No se pudo obtener userID del contexto o es inválido.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Usuario no autenticado o ID de usuario inválido."})
		return
	}

	// Parsear el formulario multipart, limitando el tamaño total (ej. 20MB para audios).
	if err := r.ParseMultipartForm(20 << 20); err != nil { // Límite aumentado para audios
		logger.Errorf("UploadAudio.ParseForm", "Error parseando multipart form: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Solicitud inválida: " + err.Error()})
		return
	}

	file, handler, err := r.FormFile("audio") // "audio" es el nombre del campo en el form-data
	if err != nil {
		logger.Errorf("UploadAudio.FormFile", "Error obteniendo el archivo 'audio' del formulario: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al recibir el archivo de audio: " + err.Error()})
		return
	}
	defer file.Close()

	logger.Infof("UploadAudio", "Recibida solicitud de subida de audio del usuario %d, archivo: %s, tamaño: %d", userID, handler.Filename, handler.Size)

	uploadDetails, err := h.audioService.ProcessAndUploadAudio(r.Context(), userID, file, handler)
	if err != nil {
		logger.Errorf("UploadAudio.ServiceCall", "Error procesando el audio para el usuario %d: %v", userID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // O BadRequest dependiendo del error
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al procesar el archivo de audio: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(uploadDetails)
}
