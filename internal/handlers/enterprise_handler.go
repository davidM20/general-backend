package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"log"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	// Importar auth si necesitas verificar roles
)

// EnterpriseHandler maneja las peticiones relacionadas con las empresas
type EnterpriseHandler struct {
	DB *sql.DB
	// Cfg *config.Config
}

// NewEnterpriseHandler crea una nueva instancia de EnterpriseHandler
func NewEnterpriseHandler(db *sql.DB) *EnterpriseHandler {
	return &EnterpriseHandler{DB: db}
}

// RegisterEnterprise maneja el registro de una nueva empresa
func (h *EnterpriseHandler) RegisterEnterprise(w http.ResponseWriter, r *http.Request) {
	// TODO: Proteger con middleware de autenticación y verificar roles (¿quién puede registrar empresas?)
	/*
			   claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
			   if !ok || claims == nil {
			       http.Error(w, "Unauthorized", http.StatusUnauthorized)
			       return
			   }
			   // Verificar si claims.RoleID tiene permiso
			   isAdmin := claims.RoleID == 7 || claims.RoleID == 8 // Ejemplo: admin o superadmin
		       isEmpresaRole := claims.RoleID == 9 // Ejemplo: rol Empresa
		       // Lógica de permisos... ¿Un usuario normal puede crear una empresa asociada a él?
		       // ¿O solo admins? ¿O usuarios con rol Empresa?
	*/

	var req models.Enterprise
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validaciones básicas
	if req.RIF == "" || req.CompanyName == "" {
		http.Error(w, "Missing required fields (RIF, CompanyName)", http.StatusBadRequest)
		return
	}
	// TODO: Añadir validaciones más específicas (formato RIF, longitud, etc.)

	// Verificar si el RIF ya existe
	var exists bool
	err := h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM Enterprise WHERE RIF = ?)", req.RIF).Scan(&exists)
	if err != nil {
		log.Printf("Error checking enterprise RIF existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Enterprise with this RIF already exists", http.StatusConflict)
		return
	}

	// Insertar la empresa
	result, err := h.DB.Exec(`
        INSERT INTO Enterprise (RIF, CompanyName, CategoryId, Description, Location, Phone)
        VALUES (?, ?, ?, ?, ?, ?)
    `, req.RIF, req.CompanyName, req.CategoryId, req.Description, req.Location, req.Phone)
	if err != nil {
		log.Printf("Error inserting enterprise: %v", err)
		http.Error(w, "Failed to register enterprise", http.StatusInternalServerError)
		return
	}

	enterpriseID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID for enterprise: %v", err)
		http.Error(w, "Error processing registration", http.StatusInternalServerError)
		return
	}

	// Devolver respuesta
	w.WriteHeader(http.StatusCreated)
	req.Id = enterpriseID // Añadir el ID generado al struct para la respuesta
	json.NewEncoder(w).Encode(req)
}

// TODO: Implementar GetEnterprises (Listar/Buscar, podría ser WS)
// TODO: Implementar GetEnterpriseByID (Ver detalle, podría ser WS)
// TODO: Implementar UpdateEnterprise (Actualizar, podría ser WS)
