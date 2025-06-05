package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"
	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/middleware"
	"github.com/davidM20/micro-service-backend-go.git/internal/services"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/gorilla/mux"
)

/*
 * ===================================================
 * HANDLER PARA LA SUBIDA Y VISUALIZACIÓN DE IMÁGENES
 * ===================================================
 *
 * Este handler gestiona las solicitudes HTTP para subir y visualizar imágenes.
 * Extrae el archivo de la solicitud, obtiene el ID de usuario autenticado
 * y llama al ImageUploadService para procesar y guardar la imagen.
 */

// ImageHandler maneja las solicitudes de subida y visualización de imágenes.
type ImageHandler struct {
	imageService *services.ImageUploadService
	cfg          *config.Config // Añadido para acceder a la configuración (ej. JWT secret, GCS bucket)
}

// NewImageHandler crea una nueva instancia de ImageHandler.
func NewImageHandler(imageService *services.ImageUploadService, cfg *config.Config) *ImageHandler {
	return &ImageHandler{imageService: imageService, cfg: cfg}
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

// ViewImage maneja la solicitud GET para ver una imagen, autenticando con token en query param.
func (h *ImageHandler) ViewImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	if filename == "" {
		logger.Warn("ViewImage.Params", "Nombre de archivo no proporcionado en la ruta.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Nombre de archivo requerido."})
		return
	}

	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		logger.Warn("ViewImage.Auth", "Token no proporcionado en query params.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Token de autenticación requerido."})
		return
	}

	// Validar el token
	claims, err := auth.ValidateJWT(tokenStr, []byte(h.cfg.JwtSecret))
	if err != nil {
		logger.Warnf("ViewImage.Auth", "Token inválido: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Token inválido o expirado."})
		return
	}

	// Log opcional del usuario autenticado
	logger.Infof("ViewImage.Auth", "Acceso autorizado para UserID: %s a imagen: %s", claims.Subject, filename)

	// Construir la URL de GCS
	if h.cfg.GCSBucketName == "" {
		logger.Error("ViewImage.Config", "El nombre del bucket GCS no está configurado en el servidor.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error de configuración del servidor."})
		return
	}

	gcsURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", h.cfg.GCSBucketName, filename)

	// Descargar la imagen desde GCS
	client := &http.Client{}
	req, err := http.NewRequestWithContext(r.Context(), "GET", gcsURL, nil)
	if err != nil {
		logger.Errorf("ViewImage.GCSRequestError", "Error creando request para GCS %s: %v", gcsURL, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al solicitar la imagen."})
		return
	}

	gcsResponse, err := client.Do(req)
	if err != nil {
		logger.Errorf("ViewImage.GCSDownloadError", "Error descargando imagen de GCS %s: %v", gcsURL, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway) // 502 si GCS falla
		json.NewEncoder(w).Encode(map[string]string{"error": "No se pudo obtener la imagen del almacenamiento."})
		return
	}
	defer gcsResponse.Body.Close()

	if gcsResponse.StatusCode != http.StatusOK {
		logger.Warnf("ViewImage.GCSStatusError", "GCS devolvió estado no OK (%d) para %s", gcsResponse.StatusCode, gcsURL)
		// Devolver un error genérico o mapear el código de estado de GCS si es necesario
		w.Header().Set("Content-Type", "application/json")
		if gcsResponse.StatusCode == http.StatusNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Imagen no encontrada."})
		} else {
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error al obtener la imagen del almacenamiento."})
		}
		return
	}

	// Obtener Content-Type y Content-Length de la respuesta de GCS
	contentType := gcsResponse.Header.Get("Content-Type")
	if contentType == "" {
		logger.Warnf("ViewImage.GCSContentTypeMissing", "GCS no devolvió Content-Type para %s. Intentando deducir.", gcsURL)
		// Intento básico de deducir por extensión, aunque es menos fiable
		if strings.HasSuffix(strings.ToLower(filename), ".webp") {
			contentType = "image/webp"
		} else if strings.HasSuffix(strings.ToLower(filename), ".jpeg") || strings.HasSuffix(strings.ToLower(filename), ".jpg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(strings.ToLower(filename), ".png") {
			contentType = "image/png"
		} else {
			contentType = "application/octet-stream" // Genérico si no se puede deducir
		}
	}

	// Configurar headers de la respuesta al cliente
	w.Header().Set("Content-Type", contentType)
	if gcsResponse.ContentLength > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", gcsResponse.ContentLength))
	}

	// Escribir los bytes de la imagen en la respuesta
	// http.ServeContent podría ser una opción más robusta aquí si tuviéramos un io.ReadSeeker y un modtime.
	// Por ahora, copiamos directamente el stream.
	_, err = io.Copy(w, gcsResponse.Body)
	if err != nil {
		logger.Errorf("ViewImage.ResponseWriteError", "Error escribiendo imagen al cliente: %v", err)
		// Es posible que los headers ya se hayan enviado, por lo que es difícil enviar un error JSON aquí.
		// El cliente podría recibir una respuesta truncada.
		return
	}
}
