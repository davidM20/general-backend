package handlers

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// AdminHandler maneja las peticiones para las rutas de administrador.
type AdminHandler struct {
	DB  *sql.DB
	Cfg *config.Config
}

// NewAdminHandler crea una nueva instancia de AdminHandler.
func NewAdminHandler(db *sql.DB, cfg *config.Config) *AdminHandler {
	return &AdminHandler{
		DB:  db,
		Cfg: cfg,
	}
}

// GetDashboard es un handler de ejemplo para el dashboard de administrador.
func (h *AdminHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Bienvenido al Dashboard de Administrador"))
}

// ListUsers responde con una lista paginada de todos los usuarios del sistema.
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// 1. Parsear parámetros de paginación de la query string
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1 // Valor por defecto
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		pageSize = 10 // Valor por defecto
	}

	// 2. Obtener el conteo total de usuarios
	totalUsers, err := queries.CountTotalUsers()
	if err != nil {
		logger.Errorf("ADMIN_HANDLER", "Failed to count users: %v", err)
		http.Error(w, "Error al obtener la lista de usuarios", http.StatusInternalServerError)
		return
	}

	if totalUsers == 0 {
		// Respuesta para cuando no hay usuarios
		response := models.PaginatedUserResponse{
			CurrentPage:  1,
			PageSize:     pageSize,
			TotalPages:   0,
			TotalRecords: 0,
			Users:        []models.UserDTO{},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 3. Obtener la lista paginada de usuarios
	users, err := queries.GetUsersPaginated(page, pageSize)
	if err != nil {
		logger.Errorf("ADMIN_HANDLER", "Failed to get paginated users: %v", err)
		http.Error(w, "Error al obtener la lista de usuarios", http.StatusInternalServerError)
		return
	}

	// 4. Construir la respuesta paginada
	response := models.PaginatedUserResponse{
		CurrentPage:  page,
		PageSize:     pageSize,
		TotalPages:   int(math.Ceil(float64(totalUsers) / float64(pageSize))),
		TotalRecords: totalUsers,
		Users:        users,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ListUnapprovedCompanies responde con una lista paginada de empresas pendientes de aprobación.
func (h *AdminHandler) ListUnapprovedCompanies(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	totalCompanies, err := queries.CountUnapprovedCompanies()
	if err != nil {
		logger.Errorf("ADMIN_HANDLER", "Failed to count unapproved companies: %v", err)
		http.Error(w, "Error al obtener la lista de empresas", http.StatusInternalServerError)
		return
	}

	if totalCompanies == 0 {
		response := models.PaginatedCompanyApprovalResponse{
			CurrentPage:  1,
			PageSize:     pageSize,
			TotalPages:   0,
			TotalRecords: 0,
			Companies:    []models.CompanyApprovalDTO{},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	companies, err := queries.GetUnapprovedCompaniesPaginated(page, pageSize)
	if err != nil {
		logger.Errorf("ADMIN_HANDLER", "Failed to get unapproved companies: %v", err)
		http.Error(w, "Error al obtener la lista de empresas", http.StatusInternalServerError)
		return
	}

	response := models.PaginatedCompanyApprovalResponse{
		CurrentPage:  page,
		PageSize:     pageSize,
		TotalPages:   int(math.Ceil(float64(totalCompanies) / float64(pageSize))),
		TotalRecords: totalCompanies,
		Companies:    companies,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
