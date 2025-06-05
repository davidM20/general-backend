package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"
	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/middleware"
	"github.com/davidM20/micro-service-backend-go.git/internal/services"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/gorilla/mux"
)

/*
 * ===================================================
 * HANDLER PARA LA SUBIDA Y VISUALIZACIÓN DE ARCHIVOS DE AUDIO
 * ===================================================
 *
 * Este handler gestiona las solicitudes HTTP para subir y visualizar archivos de audio.
 * Extrae el archivo de la solicitud, obtiene el ID de usuario autenticado
 * y llama al AudioUploadService para procesar y guardar el audio.
 */

// AudioHandler maneja las solicitudes de subida y visualización de audio.
type AudioHandler struct {
	audioService *services.AudioUploadService
	cfg          *config.Config // Añadido para JWT y GCS config
}

// NewAudioHandler crea una nueva instancia de AudioHandler.
func NewAudioHandler(audioService *services.AudioUploadService, cfg *config.Config) *AudioHandler {
	return &AudioHandler{audioService: audioService, cfg: cfg}
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

// ViewAudio maneja la solicitud GET para ver/reproducir un audio, autenticando con token en query param.
func (h *AudioHandler) ViewAudio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	if filename == "" {
		logger.Warn("ViewAudio.Params", "Nombre de archivo no proporcionado en la ruta.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Nombre de archivo requerido."})
		return
	}

	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		logger.Warn("ViewAudio.Auth", "Token no proporcionado en query params.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Token de autenticación requerido."})
		return
	}

	claims, err := auth.ValidateJWT(tokenStr, []byte(h.cfg.JwtSecret))
	if err != nil {
		logger.Warnf("ViewAudio.Auth", "Token inválido: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Token inválido o expirado."})
		return
	}

	logger.Infof("ViewAudio.Auth", "Acceso autorizado para UserID: %s a audio: %s", claims.Subject, filename)

	if h.cfg.GCSBucketName == "" {
		logger.Error("ViewAudio.Config", "El nombre del bucket GCS no está configurado.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error de configuración del servidor."})
		return
	}

	gcsURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", h.cfg.GCSBucketName, filename)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(r.Context(), "GET", gcsURL, nil)
	if err != nil {
		logger.Errorf("ViewAudio.GCSRequestError", "Error creando request para GCS %s: %v", gcsURL, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al solicitar el audio."})
		return
	}

	gcsResponse, err := client.Do(req)
	if err != nil {
		logger.Errorf("ViewAudio.GCSDownloadError", "Error descargando audio de GCS %s: %v", gcsURL, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": "No se pudo obtener el audio del almacenamiento."})
		return
	}
	defer gcsResponse.Body.Close()

	if gcsResponse.StatusCode != http.StatusOK {
		logger.Warnf("ViewAudio.GCSStatusError", "GCS devolvió estado no OK (%d) para %s", gcsResponse.StatusCode, gcsURL)
		w.Header().Set("Content-Type", "application/json")
		if gcsResponse.StatusCode == http.StatusNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Audio no encontrado."})
		} else {
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error al obtener el audio del almacenamiento."})
		}
		return
	}

	contentType := gcsResponse.Header.Get("Content-Type")
	if contentType == "" {
		logger.Warnf("ViewAudio.GCSContentTypeMissing", "GCS no devolvió Content-Type para %s. Intentando deducir.", gcsURL)
		// Deducción simple basada en extensión para tipos de audio comunes
		ext := ""
		if dotIndex := strings.LastIndex(filename, "."); dotIndex != -1 {
			ext = strings.ToLower(filename[dotIndex:])
		}
		switch ext {
		case ".mp3":
			contentType = "audio/mpeg"
		case ".wav":
			contentType = "audio/wav"
		case ".m4a":
			contentType = "audio/mp4"
		case ".ogg":
			contentType = "audio/ogg"
		case ".opus":
			contentType = "audio/opus"
		case ".flac":
			contentType = "audio/flac"
		case ".webm": // WebM puede ser video o audio
			contentType = "audio/webm" // Asumimos audio en este contexto
		default:
			contentType = "application/octet-stream"
		}
	}

	w.Header().Set("Content-Type", contentType)
	if gcsResponse.ContentLength > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", gcsResponse.ContentLength))
	}
	// Para audio, es bueno permitir que los clientes soliciten rangos (streaming)
	w.Header().Set("Accept-Ranges", "bytes")

	// Servir el contenido. http.ServeContent es más robusto para servir archivos y maneja rangos.
	// Para usarlo necesitaríamos el modtime del archivo y que gcsResponse.Body sea un io.ReadSeeker.
	// Como gcsResponse.Body es solo un io.ReadCloser, seguimos con io.Copy por simplicidad.
	// Si el streaming/seek es crítico, se necesitaría un buffer intermedio o una librería.
	_, err = io.Copy(w, gcsResponse.Body)
	if err != nil {
		logger.Errorf("ViewAudio.ResponseWriteError", "Error escribiendo audio al cliente: %v", err)
		return
	}
}
