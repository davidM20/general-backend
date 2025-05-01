package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"   // Para JWT y hash de contraseña
	"github.com/davidM20/micro-service-backend-go.git/internal/config" // Importar config
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/gorilla/mux"

	// Importa otros paquetes necesarios (ej. para validación, logging)
	"log"

	"golang.org/x/crypto/bcrypt"
)

// AuthHandler maneja las peticiones relacionadas con autenticación y registro
type AuthHandler struct {
	DB  *sql.DB
	Cfg *config.Config // Añadir configuración
}

// NewAuthHandler crea una nueva instancia de AuthHandler
func NewAuthHandler(db *sql.DB, cfg *config.Config) *AuthHandler { // Añadir cfg como parámetro
	return &AuthHandler{DB: db, Cfg: cfg} // Almacenar cfg
}

// RegisterStep1 maneja el primer paso del registro de usuario
func (h *AuthHandler) RegisterStep1(w http.ResponseWriter, r *http.Request) {
	var req models.RegistrationStep1
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Validar los datos de entrada (longitud, formato email, etc.)
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.UserName == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Verificar si el email o username ya existen
	var exists bool
	err := h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM User WHERE Email = ? OR UserName = ?)", req.Email, req.UserName).Scan(&exists)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Email or Username already exists", http.StatusConflict)
		return
	}

	// Hashear la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Error processing registration", http.StatusInternalServerError)
		return
	}

	// Insertar usuario inicial (con estado pendiente/incompleto)
	// Asignar rol por defecto (ej. invitado o pendiente) y estado pendiente
	// TODO: Definir IDs para rol 'pendiente' y estado 'pendiente verificación'
	defaultRoleId := 4   // Asumiendo 4 = invitado/pendiente registro
	defaultStatusId := 5 // Asumiendo 5 = Pending Verification

	result, err := h.DB.Exec(`
        INSERT INTO User (FirstName, LastName, UserName, Password, Email, Phone, RoleId, StatusAuthorizedId)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `, req.FirstName, req.LastName, req.UserName, string(hashedPassword), req.Email, req.Phone, defaultRoleId, defaultStatusId)
	if err != nil {
		log.Printf("Error inserting user step 1: %v", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		http.Error(w, "Error processing registration", http.StatusInternalServerError)
		return
	}

	// TODO: Generar un token de verificación/temporal y enviarlo (email?)
	// Devolver el ID del usuario para los siguientes pasos (o el token temporal)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Step 1 complete", "userId": userID})
}

// RegisterStep2 maneja el segundo paso del registro
func (h *AuthHandler) RegisterStep2(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userID"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req models.RegistrationStep2
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Validar DocId con el formato de NationalityId si es posible
	if req.DocId == "" || req.NationalityId == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Verificar si el DocId ya existe
	var exists bool
	err = h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM User WHERE DocId = ? AND Id != ?)", req.DocId, userID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking DocId existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Document ID already registered", http.StatusConflict)
		return
	}

	// Actualizar usuario
	_, err = h.DB.Exec("UPDATE User SET DocId = ?, NationalityId = ? WHERE Id = ?",
		req.DocId, req.NationalityId, userID)
	if err != nil {
		log.Printf("Error updating user step 2: %v", err)
		http.Error(w, "Failed to update registration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Step 2 complete"})
}

// RegisterStep3 maneja el tercer paso del registro
func (h *AuthHandler) RegisterStep3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userID"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req models.RegistrationStep3
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Validar Sex y Birthdate
	if req.Sex == "" || req.Birthdate.IsZero() {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Actualizar usuario y marcar como activo/verificado
	// TODO: Decidir el RoleId final (ej. estudiante-pregrado si aplica) y StatusAuthorizedId final (ej. Active)
	finalRoleId := 1   // Asumiendo 1 = estudiante-pregrado por defecto tras registro
	finalStatusId := 1 // Asumiendo 1 = Active

	_, err = h.DB.Exec("UPDATE User SET Sex = ?, Birthdate = ?, RoleId = ?, StatusAuthorizedId = ? WHERE Id = ?",
		req.Sex, req.Birthdate, finalRoleId, finalStatusId, userID)
	if err != nil {
		log.Printf("Error updating user step 3: %v", err)
		http.Error(w, "Failed to complete registration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Registration complete"})
}

// Login maneja el inicio de sesión del usuario
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Usar username O email para login?
	// Por ahora, la consulta SQL solo busca por Email.
	if req.Email == "" || req.Password == "" { // Ajustar validación si se permite username
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	var user models.User
	var hashedPassword string
	// Quitar CreateAt y UpdateAt del SELECT y Scan
	err := h.DB.QueryRow(`
        SELECT
            Id, FirstName, LastName, UserName, Password, Email, Phone, Sex, DocId,
            NationalityId, Birthdate, Picture, DegreeId, UniversityId,
            RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin
        FROM User WHERE Email = ?
    `, req.Email).Scan(
		&user.Id, &user.FirstName, &user.LastName, &user.UserName, &hashedPassword, &user.Email, &user.Phone, &user.Sex, &user.DocId,
		&user.NationalityId, &user.Birthdate, &user.Picture, &user.DegreeId, &user.UniversityId,
		&user.RoleId, &user.StatusAuthorizedId, &user.Summary, &user.Address, &user.Github, &user.Linkedin,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		// Loguear el error específico de Scan para depuración
		log.Printf("Error scanning user data: %v", err)
		http.Error(w, "Login failed due to server error", http.StatusInternalServerError)
		return
	}

	if user.StatusAuthorizedId != 1 { // Asumiendo 1 = Active
		log.Printf("Login attempt for inactive account: UserID %d, StatusID %d", user.Id, user.StatusAuthorizedId)
		http.Error(w, "Account is not active", http.StatusForbidden)
		return
	}

	println("hashedPassword: ", hashedPassword, "req.Password: ", req.Password)

	// if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
	// 	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	// 	return
	// }

	if (hashedPassword) != (req.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := 24 * time.Hour
	tokenString, err := auth.GenerateJWT(user.Id, int64(user.RoleId), []byte(h.Cfg.JwtSecret), expirationTime)
	if err != nil {
		log.Printf("Login Error: Failed generating JWT for UserID %d: %v", user.Id, err)
		http.Error(w, "Error generating session token", http.StatusInternalServerError)
		return
	}

	// Nota: La respuesta JSON ahora incluirá los tipos sql.Null*
	// El frontend podría necesitar manejarlos (ej. verificar campo .Valid antes de usar .String, .Int64, etc.)
	resp := models.LoginResponse{
		Token: tokenString,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
