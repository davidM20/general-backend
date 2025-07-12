package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/middleware"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/services"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const jobApplicationHandlerComponent = "JOB_APPLICATION_HANDLER"

// JobApplicationHandler maneja las solicitudes HTTP para las postulaciones.
type JobApplicationHandler struct {
	service services.IJobApplication
	DB      *sql.DB
}

// NewJobApplicationHandler crea una nueva instancia de JobApplicationHandler.
func NewJobApplicationHandler(service services.IJobApplication, db *sql.DB) *JobApplicationHandler {
	return &JobApplicationHandler{
		service: service,
		DB:      db,
	}
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

	// --- Validación: Evitar que un usuario se postule a su propio evento ---
	creatorID, err := queries.GetEventCreatorID(eventID)
	if err != nil {
		if err.Error() == "evento no encontrado" {
			http.Error(w, "El evento al que intentas postularte no existe.", http.StatusNotFound)
		} else {
			http.Error(w, "Error al verificar el creador del evento.", http.StatusInternalServerError)
		}
		return
	}

	if userID == creatorID {
		http.Error(w, "No puedes postularte a tu propio evento.", http.StatusForbidden)
		return
	}
	// --- Fin de la validación ---

	var req models.JobApplicationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Cuerpo de la solicitud inválido", http.StatusBadRequest)
		return
	}

	if err := h.service.ApplyToJob(eventID, userID, req); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			logger.Warnf(jobApplicationHandlerComponent, "Intento de postulación duplicada para el evento %d por el usuario %d", eventID, userID)
			http.Error(w, "Ya te has postulado a esta oferta de trabajo.", http.StatusConflict)
			return
		}

		logger.Errorf(jobApplicationHandlerComponent, "Error en el servicio al aplicar al trabajo: %v", err)
		http.Error(w, "Error al procesar la postulación", http.StatusInternalServerError)
		return
	}

	// Crear notificación para la empresa
	go h.createApplicationNotification(eventID, userID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Postulación creada exitosamente"})
}

// createApplicationNotification es una función auxiliar para crear y enviar notificaciones de forma asíncrona.
func (h *JobApplicationHandler) createApplicationNotification(eventID, applicantID int64) {
	// 1. Obtener detalles del evento (oferta de trabajo)
	event, err := queries.GetCommunityEventByID(h.DB, eventID)
	if err != nil {
		logger.Errorf(jobApplicationHandlerComponent, "Error al obtener detalles del evento %d para notificación: %v", eventID, err)
		return
	}

	// 2. Obtener nombre del postulante
	firstName, lastName, err := queries.GetUserNameByID(applicantID)
	if err != nil {
		logger.Errorf(jobApplicationHandlerComponent, "Error al obtener nombre del postulante %d para notificación: %v", applicantID, err)
		return
	}
	applicantName := fmt.Sprintf("%s %s", firstName, lastName)
	if applicantName == " " {
		applicantName = "un postulante"
	}

	// 3. Crear el objeto de notificación/evento
	companyUserID := event.CreatedByUserId
	notification := models.Event{
		EventType:      "NEW_JOB_APPLICATION",
		EventTitle:     fmt.Sprintf("Nuevo postulante para '%s'", event.Title),
		Description:    fmt.Sprintf("%s se ha postulado a tu oferta.", applicantName),
		UserId:         companyUserID,                                  // Notificación PARA la empresa
		OtherUserId:    sql.NullInt64{Int64: applicantID, Valid: true}, // Notificación SOBRE el postulante
		ActionRequired: true,                                           // La empresa debe revisar la postulación
	}

	// Adjuntar metadata útil como el ID del evento
	metadata := map[string]int64{"communityEventId": eventID, "applicantId": applicantID}
	metadataJSON, err := json.Marshal(metadata)
	if err == nil {
		notification.Metadata = metadataJSON
	} else {
		logger.Warnf(jobApplicationHandlerComponent, "No se pudo serializar metadata para notificación del evento %d: %v", eventID, err)
	}

	// 4. Guardar la notificación en la base de datos
	if err := queries.CreateEvent(&notification); err != nil {
		logger.Errorf(jobApplicationHandlerComponent, "No se pudo crear la notificación para la empresa %d sobre el evento %d: %v", companyUserID, eventID, err)
	}

	logger.Successf(jobApplicationHandlerComponent, "Notificación de postulación creada para la empresa %d sobre el evento %d", companyUserID, eventID)
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
