package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"
	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
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
 * Este handler es responsable de la lógica de negocio relacionada con las imágenes.
 */

// ImageHandler maneja las solicitudes de subida y visualización de imágenes.
type ImageHandler struct {
	imageService *services.ImageUploadService
	cfg          *config.Config // Añadido para acceder a la configuración (ej. JWT secret, GCS bucket)
	db           *sql.DB
}

// NewImageHandler crea una nueva instancia de ImageHandler.
func NewImageHandler(imageService *services.ImageUploadService, cfg *config.Config, db *sql.DB) *ImageHandler {
	return &ImageHandler{imageService: imageService, cfg: cfg, db: db}
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

// UpdateProfilePicture maneja la subida de una nueva foto de perfil para el usuario autenticado.
func (h *ImageHandler) UpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	// 1. Obtener userID del contexto
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok || userID == 0 {
		logger.Warn("UpdateProfilePicture.Auth", "No se pudo obtener userID del contexto o es inválido.")
		http.Error(w, `{"error": "Usuario no autenticado o ID de usuario inválido."}`, http.StatusUnauthorized)
		return
	}

	// 2. Parsear el formulario y obtener el archivo
	if err := r.ParseMultipartForm(10 << 20); err != nil { // Límite de 10MB
		logger.Errorf("UpdateProfilePicture.ParseForm", "Error parseando multipart form: %v", err)
		http.Error(w, `{"error": "Solicitud inválida: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("image") // El campo se debe llamar "image"
	if err != nil {
		logger.Errorf("UpdateProfilePicture.FormFile", "Error obteniendo el archivo 'image' del formulario: %v", err)
		http.Error(w, `{"error": "Error al recibir el archivo: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	logger.Infof("UpdateProfilePicture", "Recibida solicitud de actualización de foto de perfil del usuario %d, archivo: %s", userID, handler.Filename)

	// 3. Procesar y subir la imagen usando el servicio
	uploadDetails, err := h.imageService.ProcessAndUploadImage(r.Context(), userID, file, handler)
	if err != nil {
		logger.Errorf("UpdateProfilePicture.ServiceCallUpload", "Error procesando la imagen para el usuario %d: %v", userID, err)
		http.Error(w, `{"error": "Error al procesar la imagen: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// 4. Actualizar la referencia en la tabla de usuarios
	err = h.imageService.UpdateUserProfilePicture(r.Context(), userID, uploadDetails.FileName)
	if err != nil {
		logger.Errorf("UpdateProfilePicture.ServiceCallUpdate", "Error actualizando la foto de perfil en la BD para el usuario %d: %v", userID, err)
		// Aunque la imagen se subió, el perfil no se actualizó. Se debe notificar el error.
		http.Error(w, `{"error": "La imagen fue subida pero no se pudo actualizar el perfil. Por favor, contacte a soporte."}`, http.StatusInternalServerError)
		return
	}

	// 5. Responder con éxito
	response := map[string]interface{}{
		"message":   "Foto de perfil actualizada exitosamente.",
		"fileName":  uploadDetails.FileName,
		"url":       uploadDetails.URL,
		"contentId": uploadDetails.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ViewUserProfilePicture maneja la solicitud GET para ver la foto de perfil de un usuario.
// La autenticación se realiza mediante un token JWT proporcionado como query param "token".
func (h *ImageHandler) ViewUserProfilePicture(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticar con token de query param
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		logger.Warn("ViewUserProfilePicture.Auth", "Token no proporcionado en query params.")
		http.Error(w, `{"error": "Token de autenticación requerido."}`, http.StatusUnauthorized)
		return
	}

	// Validar el token
	claims, err := auth.ValidateJWT(tokenStr, []byte(h.cfg.JwtSecret))
	if err != nil {
		logger.Warnf("ViewUserProfilePicture.Auth", "Token inválido: %v", err)
		http.Error(w, `{"error": "Token inválido o expirado."}`, http.StatusUnauthorized)
		return
	}

	// 2. Obtener userID del path
	vars := mux.Vars(r)
	userIDStr, ok := vars["userID"]
	if !ok {
		logger.Warn("ViewUserProfilePicture.Params", "userID no encontrado en la ruta.")
		http.Error(w, `{"error": "Falta el ID de usuario en la ruta."}`, http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		logger.Warnf("ViewUserProfilePicture.Params", "userID inválido en la ruta: %s", userIDStr)
		http.Error(w, `{"error": "ID de usuario inválido."}`, http.StatusBadRequest)
		return
	}

	logger.Infof("ViewUserProfilePicture.Auth", "Acceso autorizado para UserID: %s para ver perfil de UserID: %d", claims.Subject, userID)

	// 3. Obtener el nombre del archivo de la foto de perfil desde el servicio
	filename, err := h.imageService.GetUserProfilePictureFilename(r.Context(), userID)
	if err != nil {
		logger.Errorf("ViewUserProfilePicture.ServiceCall", "Error obteniendo nombre de archivo para usuario %d: %v", userID, err)
		if strings.Contains(err.Error(), "no encontrado") {
			http.Error(w, `{"error": "Usuario no encontrado."}`, http.StatusNotFound)
		} else if strings.Contains(err.Error(), "no tiene foto de perfil") {
			http.Error(w, `{"error": "El usuario no tiene foto de perfil."}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error": "Error interno al obtener la información de la imagen."}`, http.StatusInternalServerError)
		}
		return
	}

	// 4. Servir la imagen
	if h.cfg.GCSBucketName == "" {
		logger.Error("ViewUserProfilePicture.Config", "El nombre del bucket GCS no está configurado.")
		http.Error(w, `{"error": "Error de configuración del servidor."}`, http.StatusInternalServerError)
		return
	}

	gcsURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", h.cfg.GCSBucketName, filename)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(r.Context(), "GET", gcsURL, nil)
	if err != nil {
		logger.Errorf("ViewUserProfilePicture.GCSRequestError", "Error creando request para GCS %s: %v", gcsURL, err)
		http.Error(w, `{"error": "Error al solicitar la imagen."}`, http.StatusInternalServerError)
		return
	}

	gcsResponse, err := client.Do(req)
	if err != nil {
		logger.Errorf("ViewUserProfilePicture.GCSDownloadError", "Error descargando imagen de GCS %s: %v", gcsURL, err)
		http.Error(w, `{"error": "No se pudo obtener la imagen del almacenamiento."}`, http.StatusBadGateway)
		return
	}
	defer gcsResponse.Body.Close()

	if gcsResponse.StatusCode != http.StatusOK {
		logger.Warnf("ViewUserProfilePicture.GCSStatusError", "GCS devolvió estado no OK (%d) para %s", gcsResponse.StatusCode, gcsURL)
		if gcsResponse.StatusCode == http.StatusNotFound {
			http.Error(w, `{"error": "Imagen no encontrada en el almacenamiento."}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error": "Error al obtener la imagen del almacenamiento."}`, http.StatusBadGateway)
		}
		return
	}

	contentType := gcsResponse.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	if gcsResponse.ContentLength > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", gcsResponse.ContentLength))
	}

	_, err = io.Copy(w, gcsResponse.Body)
	if err != nil {
		logger.Errorf("ViewUserProfilePicture.ResponseWriteError", "Error escribiendo imagen al cliente: %v", err)
	}
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

// GetMultimediaInfo maneja la solicitud GET para obtener los metadatos de un archivo multimedia
// utilizando el ID o el FileName como parámetro de consulta.
func (h *ImageHandler) GetMultimediaInfo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	filename := r.URL.Query().Get("filename")

	if (id == "" && filename == "") || (id != "" && filename != "") {
		logger.Warn("GetMultimediaInfo.Params", "Se debe proporcionar 'id' o 'filename', pero no ambos.")
		http.Error(w, `{"error": "Debe proporcionar 'id' o 'filename' como parámetro de consulta, pero no ambos."}`, http.StatusBadRequest)
		return
	}

	multimedia, err := queries.GetMultimedia(r.Context(), h.db, id, filename)

	if err != nil {
		if strings.Contains(err.Error(), "no encontrado") {
			logger.Warnf("GetMultimediaInfo.NotFound", "No se encontró multimedia: %v", err)
			http.Error(w, `{"error": "Recurso no encontrado."}`, http.StatusNotFound)
		} else {
			logger.Errorf("GetMultimediaInfo.DBError", "Error al obtener multimedia: %v", err)
			http.Error(w, `{"error": "Error interno del servidor."}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(multimedia); err != nil {
		logger.Errorf("GetMultimediaInfo.JSONError", "Error al codificar la respuesta JSON: %v", err)
	}
}
