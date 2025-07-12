package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

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

	/*
	* ===================================================
	* REGLAS DE VALIDACIÓN POR TIPO DE PUBLICACIÓN (PostType)
	* ===================================================
	*
	* - EVENTO: Requiere 'title', 'description', 'eventDate' y 'location'.
	* - NOTICIA/ARTICULO: Requiere 'title' y 'description'. 'contentUrl' es recomendado.
	* - ANUNCIO: Requiere 'title' y 'description'.
	* - DESAFIO: Requiere 'title', 'description' y 'challengeEndDate'.
	* - DISCUSION: Requiere 'title' (como la pregunta principal). 'description' es opcional.
	* - MULTIMEDIA: No requiere 'title', pero sí 'description' y al menos uno de ('imageUrl' o 'contentUrl').
	*
	 */

	// --- Validación Dinámica ---
	if req.PostType == "" {
		http.Error(w, "El campo 'post_type' es requerido", http.StatusBadRequest)
		return
	}

	switch req.PostType {
	case "EVENTO":
		if req.Title == "" || req.Description == nil || *req.Description == "" || req.EventDate == nil || req.Location == nil || *req.Location == "" {
			http.Error(w, "Para 'EVENTO', se requieren: title, description, eventDate y location", http.StatusBadRequest)
			return
		}
	case "NOTICIA", "ARTICULO", "ANUNCIO":
		if req.Title == "" || req.Description == nil || *req.Description == "" {
			http.Error(w, "Para este tipo de post, se requieren: title y description", http.StatusBadRequest)
			return
		}
	case "DESAFIO":
		if req.Title == "" || req.Description == nil || *req.Description == "" || req.ChallengeEndDate == nil {
			http.Error(w, "Para 'DESAFIO', se requieren: title, description y challengeEndDate", http.StatusBadRequest)
			return
		}
	case "DISCUSION":
		if req.Title == "" {
			http.Error(w, "Para 'DISCUSION', se requiere un 'title' que actúe como pregunta", http.StatusBadRequest)
			return
		}
	case "MULTIMEDIA":
		if (req.Description == nil || *req.Description == "") || (req.ImageUrl == nil || *req.ImageUrl == "") && (req.ContentUrl == nil || *req.ContentUrl == "") {
			http.Error(w, "Para 'MULTIMEDIA', se requiere 'description' y al menos uno de ('imageUrl' o 'contentUrl')", http.StatusBadRequest)
			return
		}
	default:
		// Para cualquier otro tipo no definido, se requiere al menos un título.
		if req.Title == "" {
			logger.Warnf("COMMUNITY_EVENT_HANDLER", "CreateCommunityEvent: Faltan campos requeridos para PostType no estándar: %s", req.PostType)
			http.Error(w, "El campo 'title' es requerido", http.StatusBadRequest)
			return
		}
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

// GetMyCommunityEvents maneja la solicitud para obtener los eventos publicados por el usuario autenticado.
func (h *CommunityEventHandler) GetMyCommunityEvents(w http.ResponseWriter, r *http.Request) {
	// Extraer userID y roleID del contexto
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok {
		logger.Warn("COMMUNITY_EVENT_HANDLER", "GetMyCommunityEvents: UserID no encontrado en el contexto")
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
		return
	}

	roleID, ok := r.Context().Value(middleware.RoleIDContextKey).(int64)
	if !ok {
		logger.Warnf("COMMUNITY_EVENT_HANDLER", "GetMyCommunityEvents: RoleID no encontrado en el contexto para el usuario %d", userID)
		http.Error(w, "Rol de usuario no encontrado", http.StatusForbidden)
		return
	}

	// Verificar que el rol es 3 (Empresa) o 1 (Estudiante)
	if roleID != 3 && roleID != 1 && roleID != 2 {
		logger.Warnf("COMMUNITY_EVENT_HANDLER", "GetMyCommunityEvents: Usuario %d con rol %d intentó acceder a un recurso restringido", userID, roleID)
		http.Error(w, "Acceso denegado. Este recurso es solo para empresas o estudiantes.", http.StatusForbidden)
		return
	}

	// Parsear parámetros de paginación de la query string
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 10 // Valor por defecto
	}
	if pageSize > 100 {
		pageSize = 100 // Límite máximo
	}

	// Llamar al servicio
	paginatedResponse, err := h.Service.GetMyCommunityEvents(userID, page, pageSize)
	if err != nil {
		// El error ya está logueado en el servicio/queries
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(paginatedResponse); err != nil {
		logger.Errorf("COMMUNITY_EVENT_HANDLER", "GetMyCommunityEvents: Error codificando la respuesta JSON: %v", err)
	}
}
