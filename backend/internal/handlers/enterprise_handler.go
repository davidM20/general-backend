package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"
	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
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
// Campos mínimos requeridos:
// - CompanyName: Nombre de la empresa
// - RIF: Registro de Información Fiscal
// - Sector: Sector empresarial
// - FirstName: Nombre del contacto
// - Email: Correo electrónico
// - Phone: Teléfono de contacto
// - Password: Contraseña para acceso
func (h *EnterpriseHandler) RegisterEnterprise(w http.ResponseWriter, r *http.Request) {
	var req models.EnterpriseRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("ENTERPRISE", "Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validaciones básicas de campos requeridos
	if req.CompanyName == "" || req.RIF == "" || req.Sector == "" ||
		req.FirstName == "" || req.Email == "" || req.Phone == "" || req.Password == "" {
		http.Error(w, "Missing required fields (companyName, rif, sector, contactName, email, phone, password)", http.StatusBadRequest)
		return
	}

	// TODO: Validar formato de RIF, email, teléfono y otros campos

	// Verificar si ya existe un usuario con ese email
	existsEmail, err := queries.MeasureQueryWithResult(func() (bool, error) {
		return queries.CheckEmailExists(h.DB, req.Email)
	})
	if err != nil {
		logger.Errorf("ENTERPRISE", "Error checking email existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existsEmail {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	// Verificar si ya existe una empresa con ese RIF
	existsRIF, err := queries.MeasureQueryWithResult(func() (bool, error) {
		return queries.CheckRIFExists(h.DB, req.RIF)
	})
	if err != nil {
		logger.Errorf("ENTERPRISE", "Error checking RIF existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existsRIF {
		http.Error(w, "Enterprise with this RIF already exists", http.StatusConflict)
		return
	}

	// Hashear la contraseña (debería usar una función de hash adecuada)
	// TODO: Implementar hasheo de contraseña adecuado
	hashedPassword := req.Password // Reemplazar con implementación real de hash
	req.Password = hashedPassword

	// Registrar la empresa en la base de datos
	userId, err := queries.MeasureQueryWithResult(func() (int64, error) {
		return queries.RegisterEnterprise(h.DB, &req)
	})
	if err != nil {
		logger.Errorf("ENTERPRISE", "Error registering enterprise: %v", err)
		http.Error(w, "Failed to register enterprise", http.StatusInternalServerError)
		return
	}

	// Preparar respuesta de éxito (sin incluir la contraseña)
	response := models.EnterpriseResponse{
		ID:          userId,
		CompanyName: req.CompanyName,
		RIF:         req.RIF,
		Email:       req.Email,
		Message:     "Enterprise registered successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateEnterpriseProfile maneja la actualización del perfil de la empresa autenticada.
func (h *EnterpriseHandler) UpdateEnterpriseProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Obtener el ID de usuario del token JWT
	claims, ok := r.Context().Value(auth.ClaimsKey).(*auth.Claims)
	if !ok || claims == nil {
		http.Error(w, "Authentication error: missing claims", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// 2. Verificar que el rol es de empresa (opcional pero recomendado si hay lógica específica)
	// Aquí asumimos que el middleware ya podría haberlo verificado o que la ruta es solo para empresas.

	// 3. Decodificar el cuerpo de la petición
	var req models.EnterpriseProfileUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("ENTERPRISE_UPDATE", "Error decoding request for user %d: %v", userID, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 4. Llamar a la función de la base de datos para actualizar el perfil
	err := queries.UpdateEnterpriseProfile(h.DB, userID, &req)
	if err != nil {
		logger.Errorf("ENTERPRISE_UPDATE", "Error updating profile for user %d: %v", userID, err)
		http.Error(w, "Failed to update enterprise profile", http.StatusInternalServerError)
		return
	}

	// 5. Enviar respuesta de éxito
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

// TODO: Implementar GetEnterprises (Listar/Buscar, podría ser WS)
// TODO: Implementar GetEnterpriseByID (Ver detalle, podría ser WS)
// TODO: Implementar UpdateEnterprise (Actualizar, podría ser WS)
