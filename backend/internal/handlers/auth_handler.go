package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"   // Para JWT y hash de contraseña
	"github.com/davidM20/micro-service-backend-go.git/internal/config" // Importar config
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/davidM20/micro-service-backend-go.git/pkg/phonetic"

	// Importa otros paquetes necesarios (ej. para validación, logging)

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/middleware"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mail.v2"
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

// Register maneja el primer paso del registro de usuario una vez que se ha registrado los pasos siguientes ocurren al hacer login
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
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

	// Verificar si el email o username ya existen usando la consulta centralizada
	exists, err := queries.CheckUserExists(h.DB, req.Email, req.UserName)
	if err != nil {
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
		logger.Errorf("REGISTER", "Error hashing password for %s: %v", req.Email, err)
		http.Error(w, "Error processing registration", http.StatusInternalServerError)
		return
	}

	// Generar claves fonéticas para el nombre de la persona
	fullName := req.FirstName + " " + req.LastName
	pKey, sKey, err := phonetic.GenerateKeysForPhrase(fullName)
	if err != nil {
		logger.Errorf("REGISTER", "Error al generar claves fonéticas para el usuario %s: %v", req.Email, err)
		// No detener el registro, pero loguear el error. Las claves se guardarán vacías.
		pKey = ""
		sKey = ""
	}

	// Insertar usuario inicial usando la consulta centralizada
	// Asignar rol por defecto (ej. invitado o pendiente) y estado pendiente
	// TODO: Definir IDs para rol 'pendiente' y estado 'pendiente verificación'
	defaultRoleId := 1   // Asumiendo 1 = /estudiante
	defaultStatusId := 1 // Asumiendo 1 como verificado, 2 como pendiente de verificación

	userID, err := queries.RegisterNewUser(h.DB, req, string(hashedPassword), defaultRoleId, defaultStatusId, pKey, sKey)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	// TODO: Generar un token de verificación/temporal y enviarlo (email?)
	// Devolver el ID del usuario para los siguientes pasos (o el token temporal)
	logger.Successf("REGISTER", "User %s completed step 1 registration with ID %d", req.Email, userID)

	// Crear un chat consigo mismo para notas/borradores usando UUID
	selfChatID := uuid.NewString()
	if err := queries.CreateContact(userID, userID, selfChatID, "accepted"); err != nil {
		// Loguear el error pero no interrumpir el registro
		logger.Errorf("REGISTER", "Failed to create self-contact for user %d: %v", userID, err)
	}

	// Enviar notificación de bienvenida
	welcomeNotif := models.Event{
		EventType:   "WELCOME_MESSAGE",
		EventTitle:  "¡Te damos la bienvenida!",
		Description: "Gracias por registrarte. ¡Explora la plataforma y conecta con otros!",
		UserId:      userID,
		OtherUserId: sql.NullInt64{Valid: false},
		ProyectId:   sql.NullInt64{Valid: false},
		GroupId:     sql.NullInt64{Valid: false},
	}
	if _, err := queries.CreateNotification(welcomeNotif); err != nil {
		logger.Errorf("REGISTER", "Failed to create welcome notification for user %d: %v", userID, err)
	}

	// Enviar notificación sobre el chat personal
	draftChatNotif := models.Event{
		EventType:   "SELF_CHAT_INFO",
		EventTitle:  "Tu espacio personal para notas",
		Description: "Hemos creado un chat contigo mismo. Puedes usarlo para guardar notas, enlaces o como borrador.",
		UserId:      userID,
		OtherUserId: sql.NullInt64{Valid: false},
		ProyectId:   sql.NullInt64{Valid: false},
		GroupId:     sql.NullInt64{Valid: false},
	}
	if _, err := queries.CreateNotification(draftChatNotif); err != nil {
		logger.Errorf("REGISTER", "Failed to create self-chat notification for user %d: %v", userID, err)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Step 1 complete", "userId": userID})
}

// RegisterStep2 maneja el segundo paso del registro
func (h *AuthHandler) RegisterStep2(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del usuario del contexto (establecido por el middleware de autenticación)
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok {
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
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

	// Verificar si el DocId ya existe usando la consulta centralizada
	exists, err := queries.CheckDocIdExists(h.DB, req.DocId, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Document ID already registered", http.StatusConflict)
		return
	}

	// Actualizar usuario usando la consulta centralizada
	err = queries.UpdateUserStep2(h.DB, userID, req.DocId, req.NationalityId)
	if err != nil {
		http.Error(w, "Failed to update registration", http.StatusInternalServerError)
		return
	}

	logger.Successf("REGISTER", "User ID %d completed step 2 registration", userID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Step 2 complete"})
}

// RegisterStep3 maneja el tercer paso del registro
func (h *AuthHandler) RegisterStep3(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del usuario del contexto (establecido por el middleware de autenticación)
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok {
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
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

	// Actualizar usuario y marcar como activo/verificado usando la consulta centralizada
	// TODO: Decidir el RoleId final (ej. estudiante-pregrado si aplica) y StatusAuthorizedId final (ej. Active)
	finalRoleId := 1   // Asumiendo 1 = estudiante-pregrado por defecto tras registro
	finalStatusId := 1 // Asumiendo 1 = Active

	err := queries.UpdateUserStep3(h.DB, userID, req.Sex, req.Birthdate, finalRoleId, finalStatusId)
	if err != nil {
		http.Error(w, "Failed to complete registration", http.StatusInternalServerError)
		return
	}

	logger.Successf("REGISTER", "User ID %d completed full registration", userID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Registration complete"})
}

// RegisterCompany maneja el registro de una nueva empresa
func (h *AuthHandler) RegisterCompany(w http.ResponseWriter, r *http.Request) {
	var req models.CompanyRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validar los datos de entrada
	if req.Email == "" || req.Password == "" || req.CompanyName == "" || req.RIF == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Verificar si el email o RIF ya existen
	exists, err := queries.CheckCompanyExists(req.Email, req.RIF)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Email or RIF already exists", http.StatusConflict)
		return
	}

	// Hashear la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("REGISTER_COMPANY", "Error hashing password for %s: %v", req.Email, err)
		http.Error(w, "Error processing registration", http.StatusInternalServerError)
		return
	}

	// Rol para empresa es 3, estado pendiente de verificación es 2
	companyRoleId := 3
	defaultStatusId := 1 // 2 = Activo/Verificado, 1 = Pendiente de Verificación

	userID, err := queries.RegisterNewCompany(h.DB, req, string(hashedPassword), companyRoleId, defaultStatusId)
	if err != nil {
		http.Error(w, "Failed to register company", http.StatusInternalServerError)
		return
	}

	logger.Successf("REGISTER_COMPANY", "Company %s completed registration with ID %d", req.CompanyName, userID)

	// Crear un chat consigo mismo para notas/borradores usando UUID
	selfChatID := uuid.NewString()
	if err := queries.CreateContact(userID, userID, selfChatID, "accepted"); err != nil {
		// Loguear el error pero no interrumpir el registro
		logger.Errorf("REGISTER_COMPANY", "Failed to create self-contact for company %d: %v", userID, err)
	}

	// Enviar notificación de bienvenida
	welcomeNotif := models.Event{
		EventType:   "WELCOME_MESSAGE",
		EventTitle:  fmt.Sprintf("¡Le damos la bienvenida a %s!", req.CompanyName),
		Description: "Gracias por registrar su empresa. ¡Explore la plataforma y conecte con el talento que busca!",
		UserId:      userID,
	}
	if _, err := queries.CreateNotification(welcomeNotif); err != nil {
		logger.Errorf("REGISTER_COMPANY", "Failed to create welcome notification for company %d: %v", userID, err)
	}

	// Enviar notificación sobre el chat personal
	draftChatNotif := models.Event{
		EventType:   "SELF_CHAT_INFO",
		EventTitle:  "Su espacio personal para notas",
		Description: "Hemos creado un chat para su empresa. Puede usarlo para guardar notas, borradores o información importante.",
		UserId:      userID,
	}
	if _, err := queries.CreateNotification(draftChatNotif); err != nil {
		logger.Errorf("REGISTER_COMPANY", "Failed to create self-chat notification for company %d: %v", userID, err)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Company registration complete", "userId": userID})
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

	// Obtener datos del usuario usando la consulta centralizada
	user, hashedPassword, err := queries.GetUserByEmail(h.DB, req.Email)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Login failed due to server error", http.StatusInternalServerError)
		return
	}

	// if user.StatusAuthorizedId != 1 { // Asumiendo 1 = Active
	// 	logger.Warnf("LOGIN", "Login attempt for inactive account: UserID %d, StatusID %d", user.Id, user.StatusAuthorizedId)
	// 	http.Error(w, "Account is not active", http.StatusForbidden)
	// 	return
	// }

	// Compara la contraseña ingresada con la contraseña hasheada almacenada
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generar el token JWT
	expirationTime := time.Hour * 24 * 360 // Token válido por 24 horas
	tokenString, tokenID, err := auth.GenerateJWT(user.Id, int64(user.RoleId), []byte(h.Cfg.JwtSecret), expirationTime)
	if err != nil {
		logger.Errorf("LOGIN", "Error generating JWT for user %s: %v", req.Email, err)
		http.Error(w, "Error generating session token", http.StatusInternalServerError)
		return
	}

	// Insertar el token en la tabla Session usando la consulta centralizada
	// clientIP := getClientIP(r)

	clientIP := "127.0.0.1"
	err = queries.RegisterUserSession(h.DB, user.Id, tokenString, clientIP, user.RoleId, tokenID)
	if err != nil {
		logger.Errorf("LOGIN", "Error creating session for user %s: %v", req.Email, err)
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	// Si el usuario es administrador, enviar notificación de seguridad en una goroutine
	if user.RoleId == int(models.RoleAdmin) {
		go h.handleAdminLoginNotification(user, clientIP)
	}

	// Preparar la respuesta
	resp := models.LoginResponse{
		Token: tokenString,
		User:  user.ToUserDTO(),
	}

	logger.Successf("LOGIN", "User %s (ID: %d) logged in successfully from IP %s", req.Email, user.Id, clientIP)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleAdminLoginNotification se encarga de enviar las notificaciones de inicio de sesión de admin.
func (h *AuthHandler) handleAdminLoginNotification(user models.User, ipAddress string) {
	// 1. Enviar correo electrónico
	err := sendAdminLoginNotification(user.Email, ipAddress)
	if err != nil {
		// El error ya se registra dentro de la función sendAdminLoginNotification
		logger.Warnf("ADMIN_LOGIN_NOTIF", "Failed to send admin login email for user %s, but login process continued.", user.Email)
	}

	j, err := json.Marshal(map[string]interface{}{"ipAddress": ipAddress, "alertSecurity": true})
	if err != nil {
		logger.Errorf("ADMIN_LOGIN_NOTIF", "Failed to marshal metadata: %v", err)
	}

	// 2. Crear notificación en la app
	notif := models.Event{
		EventType:      "ADMIN_LOGIN",
		EventTitle:     "Alerta de Seguridad: Inicio de Sesión de Administrador",
		Description:    fmt.Sprintf("Se ha iniciado sesión en una cuenta de administrador desde la IP: %s a las %s.", ipAddress, time.Now().Format("2006-01-02 15:04:05")),
		UserId:         user.Id,
		ActionRequired: true,
		Metadata:       j,
	}
	if _, err := queries.CreateNotification(notif); err != nil {
		logger.Errorf("ADMIN_LOGIN_NOTIF", "Failed to create admin login app notification for user ID %d: %v", user.Id, err)
	}
}

// getClientIP obtiene la dirección IP real del cliente.
func getClientIP(r *http.Request) string {
	// Primero, intenta obtener la IP desde X-Forwarded-For, que puede contener una lista de IPs.
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// La IP del cliente suele ser la primera en la lista.
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	// Si no, prueba con X-Real-IP.
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	// Finalmente, usa RemoteAddr como fallback.
	return r.RemoteAddr
}

// RequestPasswordReset maneja la solicitud de restablecimiento de contraseña
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	// Decodificar el cuerpo de la solicitud
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Verificar si el email existe
	user, _, err := queries.GetUserByEmail(h.DB, req.Email)
	if err == sql.ErrNoRows {
		// Por razones de seguridad, no revelamos si el email existe o no
		// Respondemos como si se hubiera enviado el correo
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Si el email existe, recibirás instrucciones para restablecer tu contraseña"})
		return
	}

	if err != nil {
		logger.Errorf("RESET_PASSWORD", "Error checking email existence: %v", err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Generar un código numérico de 5 dígitos
	resetCode, err := generateResetToken()
	if err != nil {
		logger.Errorf("RESET_PASSWORD", "Error generating reset code: %v", err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Guardar el código en la base de datos con expiración (1 hora)
	expiration := time.Now().Add(1 * time.Hour)
	err = saveResetCode(h.DB, user.Id, resetCode, expiration)
	if err != nil {
		logger.Errorf("RESET_PASSWORD", "Error saving reset code: %v", err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Enviar el correo con el código
	err = sendPasswordResetEmail(resetCode, req.Email)
	if err != nil {
		logger.Errorf("RESET_PASSWORD", "Error sending email: %v", err)
		http.Error(w, "Error sending email", http.StatusInternalServerError)
		return
	}

	logger.Successf("RESET_PASSWORD", "Password reset code sent to user %s (ID: %d)", req.Email, user.Id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Código de verificación enviado a tu correo electrónico",
	})
}

// VerifyPasswordReset verifica el código de restablecimiento y muestra la página para establecer nueva contraseña
func (h *AuthHandler) VerifyPasswordReset(w http.ResponseWriter, r *http.Request) {
	// Obtener el código de la URL
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code is required", http.StatusBadRequest)
		return
	}

	// Verificar que el código sea válido y no haya expirado
	_, valid, err := verifyResetCode(h.DB, code)
	if err != nil {
		logger.Errorf("RESET_PASSWORD", "Error verifying code: %v", err)
		http.Error(w, "Error verifying code", http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "Invalid or expired code", http.StatusBadRequest)
		return
	}

	// Redirigir al frontend con el código para completar el proceso
	redirectURL := fmt.Sprintf("%s/reset-password/complete?code=%s",
		h.Cfg.FrontendURL, code)

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// CompletePasswordReset completa el proceso de restablecimiento con la nueva contraseña
func (h *AuthHandler) CompletePasswordReset(w http.ResponseWriter, r *http.Request) {
	// Decodificar el cuerpo de la solicitud
	var req struct {
		Code        string `json:"code"`
		NewPassword string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Code == "" || req.NewPassword == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Verificar que el código sea válido y no haya expirado
	userID, valid, err := verifyResetCode(h.DB, req.Code)
	if err != nil {
		logger.Errorf("RESET_PASSWORD", "Error verifying code: %v", err)
		http.Error(w, "Error verifying code", http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "Invalid or expired code", http.StatusBadRequest)
		return
	}

	// Validar que la nueva contraseña cumpla con requisitos mínimos
	if len(req.NewPassword) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	// Hashear la nueva contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("RESET_PASSWORD", "Error hashing password: %v", err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Actualizar la contraseña en la base de datos
	err = updateUserPassword(h.DB, userID, string(hashedPassword))
	if err != nil {
		logger.Errorf("RESET_PASSWORD", "Error updating password: %v", err)
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	// Invalidar todos los códigos de restablecimiento para este usuario
	err = invalidateResetCodes(h.DB, userID)
	if err != nil {
		logger.Errorf("RESET_PASSWORD", "Error invalidating codes: %v", err)
		// No devolvemos error al cliente porque la contraseña ya se cambió
	}

	logger.Successf("RESET_PASSWORD", "Password reset completed for user ID %d", userID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Contraseña actualizada con éxito"})
}

// generateResetToken genera un código numérico de 5 dígitos para el restablecimiento de contraseña
func generateResetToken() (string, error) {
	// Generar un número aleatorio entre 10000 y 99999 (5 dígitos)
	min := 10000
	max := 99999

	// Usar crypto/rand para mayor seguridad
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return "", err
	}

	// Convertir a un número de 5 dígitos
	code := min + int(n.Int64())

	return strconv.Itoa(code), nil
}

// saveResetCode guarda el código de restablecimiento en la base de datos
func saveResetCode(db *sql.DB, userID int64, code string, expiration time.Time) error {
	// Esta función debería implementarse en el paquete queries
	query := `
		INSERT INTO PasswordReset (UserID, Code, ExpiresAt, Used)
		VALUES (?, ?, ?, 0)
	`

	_, err := db.Exec(query, userID, code, expiration)
	return err
}

// verifyResetCode verifica si un código es válido y no ha expirado
func verifyResetCode(db *sql.DB, code string) (int64, bool, error) {
	var userID int64
	var expiresAt time.Time
	var used bool

	query := `
		SELECT UserID, ExpiresAt, Used
		FROM PasswordReset
		WHERE Code = ?
	`

	err := db.QueryRow(query, code).Scan(&userID, &expiresAt, &used)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	// Verificar si el código ha sido usado o ha expirado
	if used || time.Now().After(expiresAt) {
		return 0, false, nil
	}

	return userID, true, nil
}

// updateUserPassword actualiza la contraseña de un usuario
func updateUserPassword(db *sql.DB, userID int64, hashedPassword string) error {
	query := "UPDATE User SET Password = ? WHERE Id = ?"
	_, err := db.Exec(query, hashedPassword, userID)
	return err
}

// invalidateResetCodes invalida todos los códigos de restablecimiento para un usuario
func invalidateResetCodes(db *sql.DB, userID int64) error {
	query := "UPDATE PasswordReset SET Used = 1 WHERE UserID = ?"
	_, err := db.Exec(query, userID)
	return err
}

// generatePasswordResetEmail genera el HTML para el correo de restablecimiento de contraseña
func generatePasswordResetEmail(code string) string {
	// Logo SVG profesional y moderno para Asendia con colores planos
	logo := `<svg width="180" height="60" viewBox="0 0 180 60" xmlns="http://www.w3.org/2000/svg">
		<!-- Forma principal -->
		<rect x="10" y="15" width="40" height="30" rx="2" fill="#003366" />
		<rect x="16" y="21" width="28" height="4" rx="1" fill="#ffffff" />
		<rect x="16" y="29" width="28" height="4" rx="1" fill="#ffffff" />
		<rect x="16" y="37" width="20" height="4" rx="1" fill="#ffffff" />
		
		<!-- Elemento distintivo -->
		<polygon points="55,15 65,15 65,45 55,45 60,30" fill="#0066cc" />
		
		<!-- Texto del logo -->
		<text x="70" y="38" font-family="Arial, sans-serif" font-size="22" font-weight="bold" fill="#003366">ASENDIA</text>
		
		<!-- Línea decorativa debajo del texto -->
		<rect x="70" y="42" width="80" height="2" rx="1" fill="#0066cc" />
	</svg>`

	// Simulación de la plantilla del correo
	return fmt.Sprintf(`
	<div style='background-color: #f7f9fc; padding: 30px; font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;'>
		<div style='background-color: white; border-radius: 12px; padding: 40px 30px; box-shadow: 0 8px 20px rgba(0,0,0,0.05);'>
			<div style='text-align: center; margin-bottom: 30px;'>
				%s
			</div>
			
			<h2 style='color: #003366; font-size: 24px; margin-bottom: 20px; text-align: center;'>
				Recuperación de Contraseña
			</h2>
			
			<p style='color: #333; font-size: 16px; line-height: 1.6; margin-bottom: 25px;'>
				Hemos recibido una solicitud para restablecer la contraseña de tu cuenta en Asendia.
				Si no realizaste esta solicitud, puedes ignorar este correo.
			</p>
			
			<p style='color: #333; font-size: 16px; line-height: 1.6; margin-bottom: 25px;'>
				Para crear una nueva contraseña, utiliza el siguiente código de verificación:
			</p>
			
			<div style='text-align: center; margin: 30px 0; background-color: #f2f5fa; padding: 20px; border-radius: 8px;'>
				<span style='font-size: 32px; font-weight: bold; letter-spacing: 5px; color: #003366;'>%s</span>
			</div>
			
			<p style='color: #666; font-size: 14px; line-height: 1.6;'>
				Este código expirará en 1 hora por razones de seguridad.
			</p>
			
			<hr style='border: none; border-top: 1px solid #eee; margin: 30px 0;'>
			
			<p style='color: #999; font-size: 14px; text-align: center;'>
				© %d Asendia. Todos los derechos reservados.
			</p>
		</div>
	</div>
	`, logo, code, time.Now().Year())
}

// sendPasswordResetEmail envía un correo con el código de restablecimiento
func sendPasswordResetEmail(code, email string) error {
	// Configurar el mensaje
	m := mail.NewMessage()
	m.SetHeader("From", "d18tarazona@gmail.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Código de recuperación de contraseña - Alumni USM")

	// Generar el contenido HTML del correo
	htmlContent := generatePasswordResetEmail(code)
	m.SetBody("text/html", htmlContent)

	// Configurar el servidor SMTP
	d := mail.NewDialer("smtp.gmail.com", 587, "d18tarazona@gmail.com", "hcyhtmyolvvdiauk")

	// Enviar el correo
	if err := d.DialAndSend(m); err != nil {
		logger.Errorf("RESET_PASSWORD", "Error sending email: %v", err)
		return err
	}

	logger.Successf("RESET_PASSWORD", "Password reset email sent to %s", email)
	return nil
}

// generateAdminLoginAlertEmail crea el contenido HTML para la alerta de inicio de sesión de administrador.
func generateAdminLoginAlertEmail(ipAddress string) string {
	logo := `<svg width="180" height="60" viewBox="0 0 180 60" xmlns="http://www.w3.org/2000/svg"><rect x="10" y="15" width="40" height="30" rx="2" fill="#B22222" /><polygon points="55,15 65,15 65,45 55,45 60,30" fill="#FF4500" /><text x="70" y="38" font-family="Arial, sans-serif" font-size="22" font-weight="bold" fill="#333">ALERTA</text><rect x="70" y="42" width="60" height="2" rx="1" fill="#B22222" /></svg>`
	now := time.Now().Format("02 Jan 2006 at 15:04:05 MST")

	return fmt.Sprintf(`
	<div style='background-color: #fdf2f2; padding: 30px; font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; border-left: 5px solid #B22222;'>
		<div style='background-color: white; border-radius: 8px; padding: 40px 30px; box-shadow: 0 4px 15px rgba(0,0,0,0.07);'>
			<div style='text-align: center; margin-bottom: 25px;'>
				%s
			</div>
			
			<h2 style='color: #B22222; font-size: 24px; margin-bottom: 20px; text-align: center;'>
				Alerta de Seguridad: Inicio de Sesión de Administrador
			</h2>
			
			<p style='color: #333; font-size: 16px; line-height: 1.6; margin-bottom: 15px;'>
				Este es un aviso de seguridad para informarle que se ha producido un inicio de sesión en su cuenta de administrador.
			</p>
			
			<div style='background-color: #fff8f8; border: 1px solid #fde2e2; border-radius: 8px; padding: 20px; margin: 25px 0;'>
				<p style='margin: 5px 0; font-size: 16px;'><strong style='color: #555;'>Dirección IP:</strong> <span style='font-family: monospace; color: #B22222;'>%s</span></p>
				<p style='margin: 5px 0; font-size: 16px;'><strong style='color: #555;'>Fecha y Hora:</strong> %s</p>
			</div>
			
			<p style='color: #333; font-size: 16px; line-height: 1.6;'>
				Si reconoce esta actividad, no necesita realizar ninguna acción. Si <strong>no</strong> ha sido usted, por favor, cambie su contraseña inmediatamente y contacte con el soporte técnico.
			</p>
			
			<hr style='border: none; border-top: 1px solid #eee; margin: 30px 0;'>
			
			<p style='color: #999; font-size: 14px; text-align: center;'>
				© %d Asendia Security. Este es un mensaje automático.
			</p>
		</div>
	</div>`, logo, ipAddress, now, time.Now().Year())
}

// sendAdminLoginNotification envía un correo de alerta de inicio de sesión a un administrador.
func sendAdminLoginNotification(email, ipAddress string) error {
	m := mail.NewMessage()
	m.SetHeader("From", "d18tarazona@gmail.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "⚠️ Alerta de Seguridad: Inicio de Sesión de Administrador Detectado")

	htmlContent := generateAdminLoginAlertEmail(ipAddress)
	m.SetBody("text/html", htmlContent)

	d := mail.NewDialer("smtp.gmail.com", 587, "d18tarazona@gmail.com", "hcyhtmyolvvdiauk")

	if err := d.DialAndSend(m); err != nil {
		logger.Errorf("ADMIN_LOGIN_NOTIF", "Error sending security alert email to %s: %v", email, err)
		return err
	}

	logger.Successf("ADMIN_LOGIN_NOTIF", "Security alert email sent successfully to %s", email)
	return nil
}
