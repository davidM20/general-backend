package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/middleware"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/go-sql-driver/mysql"
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
	userID, exists := r.Context().Value(middleware.UserIDContextKey).(int64)
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

// UpdateMyProfile actualiza el perfil del usuario autenticado.
// Solo actualiza los campos proporcionados en el cuerpo de la solicitud.
func (h *UserHandler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Obtener UserID del contexto
	userID, exists := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !exists {
		logger.Error("USER", "UpdateMyProfile Error: User ID not found in context")
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// 2. Decodificar el cuerpo de la solicitud
	var payload models.UpdateProfilePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logger.Errorf("USER", "Error decoding request body for UserID %d: %v", userID, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Construir la consulta de actualización dinámica
	query, args, err := queries.BuildUpdateUserQuery(userID, payload)
	if err != nil {
		// Este error ocurre si no hay campos para actualizar o si el formato de fecha es incorrecto
		logger.Warnf("USER", "Could not build update query for UserID %d: %v", userID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 4. Ejecutar la consulta
	result, err := h.DB.Exec(query, args...)
	if err != nil {
		// Manejar errores de base de datos, como claves únicas duplicadas
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// El error 1062 es para entradas duplicadas
			// Extraer el nombre del campo del mensaje de error
			// El mensaje suele ser: "Duplicate entry 'valor' for key 'nombre_del_indice'"
			parts := strings.Split(mysqlErr.Message, "'")
			var fieldName string
			if len(parts) > 3 {
				// El nombre del campo/índice suele estar en la última parte.
				// ej: "Duplicate entry 'test@test.com' for key 'user.Email'"
				keyName := parts[len(parts)-2]
				if strings.Contains(keyName, "Email") {
					fieldName = "Email"
				} else if strings.Contains(keyName, "UserName") {
					fieldName = "UserName"
				} else if strings.Contains(keyName, "DocId") {
					fieldName = "Document ID"
				} else if strings.Contains(keyName, "RIF") {
					fieldName = "RIF"
				} else {
					fieldName = "a unique field" // Fallback genérico
				}
			} else {
				fieldName = "a unique value"
			}
			errorMessage := fieldName + " is already in use."
			logger.Warnf("USER", "Duplicate entry error for UserID %d: %s", userID, errorMessage)
			http.Error(w, errorMessage, http.StatusConflict) // 409 Conflict
			return
		}

		logger.Errorf("USER", "Error executing update query for UserID %d: %v", userID, err)
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Errorf("USER", "Error getting rows affected for UserID %d: %v", userID, err)
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		logger.Warnf("USER", "No rows were updated for UserID %d. User may not exist or data was the same.", userID)
		// No es necesariamente un error, podría ser que los datos enviados eran los mismos.
		// Devolvemos un 200 OK pero con un mensaje.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "No changes detected or user not found."})
		return
	}

	logger.Successf("USER", "Profile updated successfully for UserID: %d", userID)

	// 5. Devolver una respuesta exitosa
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

// TODO: Implementar GetUserProfile (para ver perfiles de otros si es permitido)
// TODO: Implementar UpdateMyProfile (parcial o total, podría ser WS) - ¡HECHO!
