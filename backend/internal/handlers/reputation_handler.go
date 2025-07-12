package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/middleware"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/services"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

const reputationHandlerComponent = "REPUTATION_HANDLER"

// ReputationHandler maneja las solicitudes HTTP para el sistema de reputación.
type ReputationHandler struct {
	service services.IReputationService
}

// NewReputationHandler crea una nueva instancia de ReputationHandler.
func NewReputationHandler(service services.IReputationService) *ReputationHandler {
	return &ReputationHandler{service: service}
}

// CreateReview gestiona la creación de una nueva reseña.
func (h *ReputationHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	logger.Error(reputationHandlerComponent, "-> EJECUTANDO CÓDIGO ACTUALIZADO <- SI NO VE ESTE MENSAJE, EL SERVIDOR NO SE HA REINICIADO.")

	// Obtener el ID del usuario que emite la reseña desde el contexto (token JWT).
	reviewerID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok {
		logger.Error(reputationHandlerComponent, "Error: No se pudo obtener el ID del revisor del token.")
		http.Error(w, "No se pudo obtener el ID del usuario desde el token", http.StatusUnauthorized)
		return
	}
	logger.Infof(reputationHandlerComponent, "Reviewer ID from token: %d", reviewerID)

	// Verificar si el cuerpo de la petición está vacío antes de intentar decodificar.
	if r.Body == nil {
		errorMessage := "El cuerpo de la solicitud es nulo (nil)."
		logger.Error(reputationHandlerComponent, errorMessage)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": errorMessage})
		return
	}

	if r.ContentLength == 0 {
		errorMessage := "El cuerpo de la solicitud está vacío (ContentLength es 0)."
		logger.Error(reputationHandlerComponent, errorMessage)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": errorMessage})
		return
	}

	logger.Info(reputationHandlerComponent, "El cuerpo de la solicitud no está vacío. Procediendo a decodificar.")

	// Decodificar el cuerpo de la solicitud.
	var req models.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorMessage := "Cuerpo de la solicitud inválido o malformado."
		logger.Errorf(reputationHandlerComponent, "%s Error: %v", errorMessage, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   errorMessage,
			"details": err.Error(),
		})
		return
	}

	logger.Infof(reputationHandlerComponent, "Solicitud de revisión decodificada exitosamente: %+v", req)

	// Validar que el usuario no intente calificarse a sí mismo.
	if reviewerID == req.RevieweeID {
		errorMessage := fmt.Sprintf("Un usuario no puede calificarse a sí mismo. ReviewerID: %d, RevieweeID: %d", reviewerID, req.RevieweeID)
		logger.Error(reputationHandlerComponent, errorMessage)
		http.Error(w, "Un usuario no puede calificarse a sí mismo", http.StatusBadRequest)
		return
	}

	logger.Info(reputationHandlerComponent, "La validación de autocalificación pasó. Llamando al servicio...")

	// Llamar al servicio para procesar la lógica de negocio.
	if err := h.service.CreateReview(reviewerID, req); err != nil {
		logger.Errorf(reputationHandlerComponent, "Error en el servicio al crear la reseña: %v", err)
		// Aquí se podría devolver un error más específico basado en el tipo de error del servicio.
		http.Error(w, "Error al procesar la reseña", http.StatusInternalServerError)
		return
	}

	// Obtener el nombre de la empresa que hace la reseña.
	companyName, err := queries.GetCompanyNameByID(reviewerID)
	if err != nil {
		// Si no se encuentra el nombre, no es un error fatal, pero se debe loguear.
		// Se usará un texto genérico para la notificación.
		logger.Warnf(reputationHandlerComponent, "No se pudo obtener el nombre de la empresa para el revisor %d: %v", reviewerID, err)
		companyName = "Una empresa"
	}

	// Crear la notificación para que el usuario calificado pueda valorar a la empresa.
	notification := models.Event{
		EventType:      "COMPANY_REVIEW_PENDING",
		EventTitle:     fmt.Sprintf("Valora tu experiencia con %s", companyName),
		Description:    "Ahora puedes calificar a la empresa que te ha evaluado. Tu opinión es importante.",
		UserId:         req.RevieweeID,                                // Notificación PARA el usuario calificado.
		OtherUserId:    sql.NullInt64{Int64: reviewerID, Valid: true}, // Notificación se refiere a esta empresa.
		ActionRequired: true,                                          // El usuario debe realizar una acción.
	}

	if err := queries.CreateEvent(&notification); err != nil {
		// Loguear el error pero no devolver un error al cliente, ya que la operación principal (crear reseña) fue exitosa.
		logger.Errorf(reputationHandlerComponent, "No se pudo crear la notificación de reseña para el usuario %d: %v", req.RevieweeID, err)
	}

	logger.Info(reputationHandlerComponent, "Reseña creada exitosamente.")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Reseña creada exitosamente"})
}

// CreateReviewByStudent gestiona la creación de una nueva reseña de un estudiante hacia una empresa.
func (h *ReputationHandler) CreateReviewByStudent(w http.ResponseWriter, r *http.Request) {
	logger.Info(reputationHandlerComponent, "Iniciando CreateReviewByStudent...")

	// Obtener el ID del estudiante que emite la reseña desde el contexto (token JWT).
	studentID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok {
		logger.Error(reputationHandlerComponent, "Error: No se pudo obtener el ID del estudiante del token.")
		http.Error(w, "No se pudo obtener el ID del usuario desde el token", http.StatusUnauthorized)
		return
	}
	logger.Infof(reputationHandlerComponent, "StudentID from token: %d", studentID)

	// Decodificar el cuerpo de la solicitud.
	var req models.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorMessage := "Cuerpo de la solicitud inválido."
		logger.Errorf(reputationHandlerComponent, "%s Error: %v", errorMessage, err)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}
	logger.Infof(reputationHandlerComponent, "Solicitud de reseña de estudiante decodificada: %+v", req)

	// Validar que el estudiante no intente calificarse a sí mismo.
	if studentID == req.RevieweeID {
		errorMessage := "Un estudiante no puede calificarse a sí mismo."
		logger.Error(reputationHandlerComponent, errorMessage)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	// Aquí, la lógica de negocio (como verificar si el estudiante puede calificar a esta empresa)
	// debería estar en la capa de servicio.
	if err := h.service.CreateReview(studentID, req); err != nil {
		logger.Errorf(reputationHandlerComponent, "Error en el servicio al crear la reseña del estudiante: %v", err)
		http.Error(w, "Error al procesar la reseña", http.StatusInternalServerError)
		return
	}

	// Obtener el nombre del estudiante para la notificación.
	firstName, lastName, err := queries.GetUserNameByID(studentID)
	var studentName string
	if err != nil {
		logger.Warnf(reputationHandlerComponent, "No se pudo obtener el nombre del estudiante %d: %v. Usando un nombre genérico.", studentID, err)
		studentName = "Un estudiante"
	} else {
		studentName = strings.TrimSpace(fmt.Sprintf("%s %s", firstName, lastName))
		if studentName == "" {
			studentName = "Un estudiante"
		}
	}

	// Crear la notificación para la empresa que fue calificada.
	notification := models.Event{
		EventType:      "REVIEW_CREATED_BY_STUDENT",
		EventTitle:     fmt.Sprintf("%s ha valorado tu empresa.", studentName),
		Description:    fmt.Sprintf("Has recibido una nueva calificación de %.1f estrellas.", req.Rating),
		UserId:         req.RevieweeID,                               // Notificación PARA la empresa calificada.
		OtherUserId:    sql.NullInt64{Int64: studentID, Valid: true}, // Notificación DESDE el estudiante.
		ActionRequired: false,                                        // Es solo informativa.
	}

	if err := queries.CreateEvent(&notification); err != nil {
		logger.Errorf(reputationHandlerComponent, "No se pudo crear la notificación de reseña para la empresa %d: %v", req.RevieweeID, err)
		// No se retorna error al cliente, ya que la reseña se creó correctamente.
	}

	logger.Info(reputationHandlerComponent, "Reseña de estudiante creada exitosamente.")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Reseña creada exitosamente"})
}
