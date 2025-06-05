package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/middleware"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/services"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// CommunityEventHandler maneja las peticiones HTTP relacionadas con eventos comunitarios.
type CommunityEventHandler struct {
	Service *services.CommunityEventService
	Cfg     *config.Config // Puede ser necesario para otras configuraciones
}

// NewCommunityEventHandler crea una nueva instancia de CommunityEventHandler.
func NewCommunityEventHandler(db *sql.DB, cfg *config.Config) *CommunityEventHandler {
	communityEventService := services.NewCommunityEventService(db)
	return &CommunityEventHandler{
		Service: communityEventService,
		Cfg:     cfg,
	}
}

// CreateCommunityEvent maneja la creación de un nuevo evento comunitario.
func (h *CommunityEventHandler) CreateCommunityEvent(w http.ResponseWriter, r *http.Request) {
	// Obtener el UserID del usuario autenticado desde el contexto
	// El middleware de autenticación ya debería haber puesto esto en el contexto.
	createdByUserID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok {
		logger.Warn("COMMUNITY_EVENT_HANDLER", "CreateCommunityEvent: UserID no encontrado en el contexto o tipo incorrecto")
		http.Error(w, "Usuario no autenticado o error interno", http.StatusUnauthorized)
		return
	}

	var req models.CommunityEventCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warnf("COMMUNITY_EVENT_HANDLER", "CreateCommunityEvent: Error decodificando el cuerpo de la solicitud: %v", err)
		http.Error(w, "Cuerpo de la solicitud inválido", http.StatusBadRequest)
		return
	}

	// Llamar al servicio para crear el evento
	createdEvent, err := h.Service.CreateCommunityEvent(req, createdByUserID)
	if err != nil {
		// Los errores específicos de validación o de base de datos ya deberían estar logueados
		// en la capa de servicio o de queries. Aquí devolvemos un error genérico o uno más específico
		// basado en el tipo de error si es necesario.
		logger.Errorf("COMMUNITY_EVENT_HANDLER", "CreateCommunityEvent: Error creando el evento: %v", err)
		// Podríamos tener un helper para mapear errores de servicio a códigos HTTP
		http.Error(w, err.Error(), http.StatusInternalServerError) // O http.StatusBadRequest si es un error de validación
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdEvent); err != nil {
		logger.Errorf("COMMUNITY_EVENT_HANDLER", "CreateCommunityEvent: Error codificando la respuesta JSON: %v", err)
		// Si llegamos aquí, el evento ya fue creado, pero no podemos enviar la respuesta.
		// Podríamos loguear y no hacer nada más, o intentar enviar un http.Error genérico.
	}
}
