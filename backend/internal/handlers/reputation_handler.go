package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	logger.Info(reputationHandlerComponent, "Reseña creada exitosamente.")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Reseña creada exitosamente"})
}
