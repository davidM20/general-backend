package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"log"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth" // Para obtener UserID del token
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
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

// GetMyProfile obtiene el perfil del usuario autenticado actualmente
func (h *UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del usuario del contexto (establecido por el middleware)
	userIDCtx := r.Context().Value(auth.UserIDKey) // Usar la clave correcta
	if userIDCtx == nil {
		log.Println("GetMyProfile Error: User ID not found in context")
		http.Error(w, "User not authenticated properly", http.StatusUnauthorized)
		return
	}
	// Hacer type assertion a int64
	userID, ok := userIDCtx.(int64)
	if !ok {
		log.Println("GetMyProfile Error: Invalid User ID type in context")
		http.Error(w, "Invalid user information", http.StatusInternalServerError)
		return
	}

	var user models.User
	// Recuperar usuario de la BD usando el userID del token
	// Asegúrate de seleccionar todos los campos necesarios excepto la contraseña
	err := h.DB.QueryRow(`
        SELECT Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin
        FROM User WHERE Id = ?
    `, userID).Scan(
		&user.Id, &user.FirstName, &user.LastName, &user.UserName, &user.Email, &user.Phone, &user.Sex, &user.DocId, &user.NationalityId, &user.Birthdate, &user.Picture, &user.DegreeId, &user.UniversityId, &user.RoleId, &user.StatusAuthorizedId, &user.Summary, &user.Address, &user.Github, &user.Linkedin,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error fetching user profile (ID: %d): %v", userID, err)
		http.Error(w, "Failed to retrieve profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user) // Devuelve el usuario (sin contraseña)
}

// TODO: Implementar GetUserProfile (para ver perfiles de otros si es permitido)
// TODO: Implementar UpdateMyProfile (parcial o total, podría ser WS)
