package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/davidM20/micro-service-backend-go.git/internal/middleware"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/services"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/gorilla/mux"
)

const jobApplicationHandlerComponent = "JOB_APPLICATION_HANDLER"

// JobApplicationHandler maneja las solicitudes HTTP para las postulaciones.
type JobApplicationHandler struct {
	service services.IJobApplication
}

// NewJobApplicationHandler crea una nueva instancia de JobApplicationHandler.
func NewJobApplicationHandler(service services.IJobApplication) *JobApplicationHandler {
	return &JobApplicationHandler{service: service}
}

// ApplyToJob gestiona la postulación de un usuario a una oferta.
func (h *JobApplicationHandler) ApplyToJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.ParseInt(vars["eventID"], 10, 64)
	if err != nil {
		http.Error(w, "ID de evento inválido", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok {
		http.Error(w, "No se pudo obtener el ID del usuario desde el token", http.StatusUnauthorized)
		return
	}

	var req models.JobApplicationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Cuerpo de la solicitud inválido", http.StatusBadRequest)
		return
	}

	if err := h.service.ApplyToJob(eventID, userID, req); err != nil {
		logger.Errorf(jobApplicationHandlerComponent, "Error en el servicio al aplicar al trabajo: %v", err)
		http.Error(w, "Error al procesar la postulación", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Postulación creada exitosamente"})
}

// ListApplicants gestiona la solicitud para listar los postulantes de una oferta.
func (h *JobApplicationHandler) ListApplicants(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.ParseInt(vars["eventID"], 10, 64)
	if err != nil {
		http.Error(w, "ID de evento inválido", http.StatusBadRequest)
		return
	}

	// TODO: Añadir validación para asegurar que quien consulta es el creador de la oferta o un admin.
	// Por ahora, cualquier usuario autenticado puede ver los postulantes.

	applicants, err := h.service.ListApplicants(eventID)
	if err != nil {
		logger.Errorf(jobApplicationHandlerComponent, "Error en el servicio al listar aplicantes: %v", err)
		http.Error(w, "Error al obtener la lista de postulantes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(applicants)
}

// UpdateApplicationStatus gestiona la solicitud para cambiar el estado de una postulación.
func (h *JobApplicationHandler) UpdateApplicationStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.ParseInt(vars["eventID"], 10, 64)
	if err != nil {
		http.Error(w, "ID de evento inválido", http.StatusBadRequest)
		return
	}

	applicantID, err := strconv.ParseInt(vars["applicantID"], 10, 64)
	if err != nil {
		http.Error(w, "ID de aplicante inválido", http.StatusBadRequest)
		return
	}

	// TODO: Aquí se debería validar que el usuario que hace la petición
	// es el creador del evento o un administrador.

	var req models.UpdateApplicationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Cuerpo de la solicitud inválido", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateApplicationStatus(eventID, applicantID, req.Status); err != nil {
		logger.Errorf(jobApplicationHandlerComponent, "Error en el servicio al actualizar estado: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Estado de la postulación actualizado exitosamente"})
}
