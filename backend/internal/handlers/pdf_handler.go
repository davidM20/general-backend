package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"
	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/middleware"
	"github.com/davidM20/micro-service-backend-go.git/internal/services"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/gorilla/mux"
)

/*
 * ===================================================
 * HANDLER PARA LA SUBIDA Y VISUALIZACIÓN DE ARCHIVOS PDF
 * ===================================================
 *
 * Este handler gestiona las solicitudes HTTP para subir y visualizar archivos PDF.
 * Extrae el archivo de la solicitud, obtiene el ID de usuario autenticado
 * y llama al PDFUploadService para procesar y guardar el PDF.
 */

// PDFHandler maneja las solicitudes de subida y visualización de PDF.
type PDFHandler struct {
	pdfService *services.PDFUploadService
	cfg        *config.Config // Añadido para JWT y GCS config
}

// NewPDFHandler crea una nueva instancia de PDFHandler.
func NewPDFHandler(pdfService *services.PDFUploadService, cfg *config.Config) *PDFHandler {
	return &PDFHandler{pdfService: pdfService, cfg: cfg}
}

// UploadPDF es el método que maneja la petición POST para subir un archivo PDF.
func (h *PDFHandler) UploadPDF(w http.ResponseWriter, r *http.Request) {
	// Obtener userID del contexto (inyectado por AuthMiddleware)
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok || userID == 0 {
		logger.Warn("UploadPDF.Auth", "No se pudo obtener userID del contexto o es inválido.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Usuario no autenticado o ID de usuario inválido."})
		return
	}

	// Parsear el formulario multipart, limitando el tamaño total (ej. 10MB para PDFs, según MaxPDFSize en el servicio).
	// El servicio ya valida el tamaño exacto del archivo, aquí es una primera barrera.
	if err := r.ParseMultipartForm(services.MaxPDFSize + (1 << 20)); err != nil { // Límite un poco mayor que MaxPDFSize para dar margen
		logger.Errorf("UploadPDF.ParseForm", "Error parseando multipart form: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Solicitud inválida o demasiado grande: " + err.Error()})
		return
	}

	file, handler, err := r.FormFile("pdf") // "pdf" es el nombre del campo en el form-data
	if err != nil {
		logger.Errorf("UploadPDF.FormFile", "Error obteniendo el archivo 'pdf' del formulario: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al recibir el archivo PDF: " + err.Error()})
		return
	}
	defer file.Close()

	logger.Infof("UploadPDF", "Recibida solicitud de subida de PDF del usuario %d, archivo: %s, tamaño: %d", userID, handler.Filename, handler.Size)

	uploadDetails, err := h.pdfService.ProcessAndUploadPDF(r.Context(), userID, file, handler)
	if err != nil {
		logger.Errorf("UploadPDF.ServiceCall", "Error procesando el PDF para el usuario %d: %v", userID, err)
		w.Header().Set("Content-Type", "application/json")
		// Determinar el código de estado basado en el tipo de error podría ser más granular
		// Por ejemplo, si es un error de validación del servicio (tipo/tamaño), podría ser BadRequest.
		w.WriteHeader(http.StatusInternalServerError) // Asumir error de servidor si no es explícitamente de validación
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al procesar el archivo PDF: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(uploadDetails)
}

// ViewPDF maneja la solicitud GET para ver/descargar un PDF, autenticando con token en query param.
func (h *PDFHandler) ViewPDF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	if filename == "" {
		logger.Warn("ViewPDF.Params", "Nombre de archivo no proporcionado en la ruta.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Nombre de archivo requerido."})
		return
	}

	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		logger.Warn("ViewPDF.Auth", "Token no proporcionado en query params.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Token de autenticación requerido."})
		return
	}

	claims, err := auth.ValidateJWT(tokenStr, []byte(h.cfg.JwtSecret))
	if err != nil {
		logger.Warnf("ViewPDF.Auth", "Token inválido: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Token inválido o expirado."})
		return
	}

	logger.Infof("ViewPDF.Auth", "Acceso autorizado para UserID: %s a PDF: %s", claims.Subject, filename)

	if h.cfg.GCSBucketName == "" {
		logger.Error("ViewPDF.Config", "El nombre del bucket GCS no está configurado.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error de configuración del servidor."})
		return
	}

	gcsURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", h.cfg.GCSBucketName, filename)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(r.Context(), "GET", gcsURL, nil)
	if err != nil {
		logger.Errorf("ViewPDF.GCSRequestError", "Error creando request para GCS %s: %v", gcsURL, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al solicitar el PDF."})
		return
	}

	gcsResponse, err := client.Do(req)
	if err != nil {
		logger.Errorf("ViewPDF.GCSDownloadError", "Error descargando PDF de GCS %s: %v", gcsURL, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": "No se pudo obtener el PDF del almacenamiento."})
		return
	}
	defer gcsResponse.Body.Close()

	if gcsResponse.StatusCode != http.StatusOK {
		logger.Warnf("ViewPDF.GCSStatusError", "GCS devolvió estado no OK (%d) para %s", gcsResponse.StatusCode, gcsURL)
		w.Header().Set("Content-Type", "application/json")
		if gcsResponse.StatusCode == http.StatusNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "PDF no encontrado."})
		} else {
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error al obtener el PDF del almacenamiento."})
		}
		return
	}

	// Para PDFs, el Content-Type es generalmente application/pdf
	contentType := gcsResponse.Header.Get("Content-Type")
	if contentType == "" || contentType != "application/pdf" {
		logger.Warnf("ViewPDF.GCSContentTypeMismatch", "GCS devolvió Content-Type '%s' para PDF %s. Forzando a application/pdf.", contentType, gcsURL)
		contentType = "application/pdf"
	}

	w.Header().Set("Content-Type", contentType)
	if gcsResponse.ContentLength > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", gcsResponse.ContentLength))
	}
	// Content-Disposition ayuda al navegador a decidir si mostrar en línea o descargar.
	// 'inline' sugiere mostrarlo. 'attachment; filename="filename.pdf"' sugeriría descargarlo.
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename)) // Sugerir mostrar en línea

	_, err = io.Copy(w, gcsResponse.Body)
	if err != nil {
		logger.Errorf("ViewPDF.ResponseWriteError", "Error escribiendo PDF al cliente: %v", err)
		return
	}
}
