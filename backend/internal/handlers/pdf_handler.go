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
 * HANDLER PARA LA SUBIDA DE ARCHIVOS PDF
 * ===================================================
 *
 * Este handler gestiona las solicitudes HTTP para subir archivos PDF.
 * Extrae el archivo de la solicitud, obtiene el ID de usuario autenticado
 * y llama al PDFUploadService para procesar y guardar el PDF.
 */

// PDFHandler maneja las solicitudes de subida de PDF.
type PDFHandler struct {
	pdfService *services.PDFUploadService
}

// NewPDFHandler crea una nueva instancia de PDFHandler.
func NewPDFHandler(pdfService *services.PDFUploadService) *PDFHandler {
	return &PDFHandler{pdfService: pdfService}
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
