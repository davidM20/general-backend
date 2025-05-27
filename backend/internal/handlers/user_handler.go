package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// UserHandler maneja las peticiones relacionadas con los usuarios
type UserHandler struct {
	DB *sql.DB
	// Cfg *config.Config // Añadir si se necesita configuración
}

// NewUserHandler crea una nueva instancia de UserHandler
func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{DB: db}
}

// GetMyProfile devuelve el perfil del usuario autenticado
func (h *UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	// Obtener UserID del contexto (puesto por AuthMiddleware)
	userID, exists := r.Context().Value("userID").(int64)
	if !exists {
		logger.Error("USER", "GetMyProfile Error: User ID not found in context")
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var user models.User
	err := h.DB.QueryRow(`
        SELECT Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin
        FROM User WHERE Id = ?
    `, userID).Scan(
		&user.Id, &user.FirstName, &user.LastName, &user.UserName, &user.Email, &user.Phone, &user.Sex, &user.DocId, &user.NationalityId, &user.Birthdate, &user.Picture, &user.DegreeId, &user.UniversityId, &user.RoleId, &user.StatusAuthorizedId, &user.Summary, &user.Address, &user.Github, &user.Linkedin,
	)

	if err != nil {
		logger.Errorf("USER", "Error fetching user profile (ID: %d): %v", userID, err)
		http.Error(w, "Error fetching profile", http.StatusInternalServerError)
		return
	}

	logger.Successf("USER", "Profile fetched successfully for UserID: %d", userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user.ToUserDTO()) // Usar ToUserDTO para limpiar la respuesta
}

// TODO: Implementar GetUserProfile (para ver perfiles de otros si es permitido)
// TODO: Implementar UpdateMyProfile (parcial o total, podría ser WS)
