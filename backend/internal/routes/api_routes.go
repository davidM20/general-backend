package routes

/*
* ===================================================
* GUÍA DE MANTENIMIENTO Y EXTENSIÓN DE RUTAS API
* ===================================================
*
* ESTRUCTURA DEL ARCHIVO:
* -----------------------
* Este archivo implementa todas las rutas de la API REST del microservicio siguiendo
* principios de código limpio y reutilizable:
*
* 1. Constantes para todas las rutas (evita strings repetidos)
* 2. Función principal SetupApiRoutes que configura todos los routers
* 3. Funciones específicas para diferentes dominios
* 4. Función genérica setupProtectedRoute para rutas protegidas
*
* REGLAS DE NOMENCLATURA:
* ----------------------
* - Constantes de rutas: NombreDominioPath (ej. UsersMePath, CategoriesPath)
* - Funciones de setup: setup[Público/Protected][Dominio]Routes
* - Nombres de handlers: siempre descriptivos del dominio que manejan
*
* PROCESO PARA AÑADIR NUEVAS RUTAS:
* --------------------------------
* 1. RUTAS PÚBLICAS:
*    a. Añadir la constante para la ruta en la sección de constantes
*    b. Añadir la ruta en la función setupPublic[Dominio]Routes correspondiente
*    c. Si es un nuevo dominio, crear una nueva función setupPublic[NuevoDominio]Routes
*
* 2. RUTAS PROTEGIDAS:
*    a. Añadir la constante para la ruta en la sección de constantes
*    b. Usar la función setupProtectedRoute en SetupApiRoutes, pasando:
*       - El router 'protected'
*       - La ruta (constante)
*       - El handler correspondiente
*       - El método HTTP (GET, POST, etc.)
*
* 3. NUEVOS DOMINIOS:
*    a. Añadir un nuevo campo en la estructura serviceHandlers
*    b. Inicializar el handler en la función initializeHandlers
*    c. Crear las funciones de setup necesarias para rutas públicas
*    d. Usar setupProtectedRoute para rutas protegidas
*
* PRINCIPIOS A SEGUIR:
* ------------------
* - Responsabilidad única: cada función hace una sola cosa
* - Evitar duplicación de código
* - Mantener la separación de responsabilidades
* - Documentar todas las funciones y tipos
* - Agrupar rutas por dominio funcional
 */

import (
	"database/sql"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/config"     // Importar config
	"github.com/davidM20/micro-service-backend-go.git/internal/handlers"   // Crearemos este paquete
	"github.com/davidM20/micro-service-backend-go.git/internal/middleware" // Importar middleware
	"github.com/gorilla/mux"
)

// Constantes para las rutas base
const (
	// Prefijos de rutas
	APIPrefix = "/api/v1"

	// Rutas de autenticación
	AuthPath         = "/login"
	RegisterBasePath = "/register"
	RegisterStep1    = RegisterBasePath + "/step1"
	RegisterStep2    = RegisterBasePath + "/step2"
	RegisterStep3    = RegisterBasePath + "/step3"

	// Rutas de recuperación de contraseña
	ResetPasswordPath     = "/reset-password"
	ResetPasswordRequest  = ResetPasswordPath + "/request"
	ResetPasswordVerify   = ResetPasswordPath + "/verify"
	ResetPasswordComplete = ResetPasswordPath + "/complete"

	// Rutas de usuarios
	UsersPath   = "/users"
	UsersMePath = UsersPath + "/me"

	// Rutas de empresas
	EnterprisesPath = "/enterprises"

	// Rutas de categorías
	CategoriesPath = "/categories"

	// Rutas de datos misceláneos
	NationalitiesPath = "/nationalities"
	UniversitiesPath  = "/universities"
	DegreesPath       = "/degrees/{universityID:[0-9]+}"

	// Rutas de multimedia
	MediaUploadPath = "/media/upload"

	// Rutas de sistema
	HealthPath = "/health"
)

// SetupApiRoutes configura todas las rutas para el microservicio de API REST
// siguiendo el principio de responsabilidad única y organización por dominio.
func SetupApiRoutes(r *mux.Router, db *sql.DB, cfg *config.Config) {
	// Crear instancias de los handlers
	handlers := initializeHandlers(db, cfg)

	// Crear subrouter para la API con prefijo /api/v1
	api := r.PathPrefix(APIPrefix).Subrouter()

	// Configurar rutas por dominio
	setupHealthRoutes(api)
	setupPublicAuthRoutes(api, handlers.authHandler)
	setupPublicEnterpriseRoutes(api, handlers.enterpriseHandler)
	setupPublicCategoryRoutes(api, handlers.categoryHandler)
	setupPublicMiscRoutes(api, handlers.miscHandler)

	// Configurar rutas protegidas (requieren autenticación JWT)
	protected := api.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(cfg))

	// Configurar todas las rutas protegidas usando la función genérica
	// Rutas de usuarios
	setupProtectedRoute(protected, UsersMePath, handlers.userHandler.GetMyProfile, http.MethodGet)

	// Rutas de registro protegidas (pasos 2 y 3)
	setupProtectedRoute(protected, RegisterStep2, handlers.authHandler.RegisterStep2, http.MethodPost)
	setupProtectedRoute(protected, RegisterStep3, handlers.authHandler.RegisterStep3, http.MethodPost)

	// Rutas de categorías
	setupProtectedRoute(protected, CategoriesPath, handlers.categoryHandler.AddCategory, http.MethodPost)

	// Rutas de multimedia
	setupProtectedRoute(protected, MediaUploadPath, handlers.mediaHandler.UploadMedia, http.MethodPost)

	// TODO: Implementar estas rutas cuando estén disponibles:
	// - GET /users/{userID:[0-9]+} - Ver perfil de otro usuario
	// - PUT /users/me - Actualizar perfil
	// - Gestión de empresas (listar, buscar, actualizar)
	// - Rutas para administradores (gestión de roles, categorías, etc.)
	// - Gestión de intereses
	// - Gestión de tipos de mensajes
}

// Estructura para agrupar todos los handlers y facilitar su paso a las funciones
type serviceHandlers struct {
	authHandler       *handlers.AuthHandler
	userHandler       *handlers.UserHandler
	enterpriseHandler *handlers.EnterpriseHandler
	miscHandler       *handlers.MiscHandler
	mediaHandler      *handlers.MediaHandler
	categoryHandler   *handlers.CategoryHandler
}

// initializeHandlers crea e inicializa todas las instancias de handlers necesarias
func initializeHandlers(db *sql.DB, cfg *config.Config) serviceHandlers {
	return serviceHandlers{
		authHandler:       handlers.NewAuthHandler(db, cfg),
		userHandler:       handlers.NewUserHandler(db),
		enterpriseHandler: handlers.NewEnterpriseHandler(db),
		miscHandler:       handlers.NewMiscHandler(db),
		mediaHandler:      handlers.NewMediaHandler(db, cfg),
		categoryHandler:   handlers.NewCategoryHandler(),
	}
}

// setupHealthRoutes configura las rutas de verificación de estado del sistema
func setupHealthRoutes(router *mux.Router) {
	router.HandleFunc(HealthPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API is healthy"))
	}).Methods(http.MethodGet)
}

// setupPublicAuthRoutes configura las rutas públicas de autenticación y registro
func setupPublicAuthRoutes(router *mux.Router, authHandler *handlers.AuthHandler) {
	router.HandleFunc(RegisterStep1, authHandler.Register).Methods(http.MethodPost)
	router.HandleFunc(AuthPath, authHandler.Login).Methods(http.MethodPost)

	// Rutas para recuperación de contraseña
	router.HandleFunc(ResetPasswordRequest, authHandler.RequestPasswordReset).Methods(http.MethodPost)
	router.HandleFunc(ResetPasswordVerify, authHandler.VerifyPasswordReset).Methods(http.MethodGet)
	router.HandleFunc(ResetPasswordComplete, authHandler.CompletePasswordReset).Methods(http.MethodPost)

	// Nota: Los pasos 2 y 3 del registro ahora son rutas protegidas
	// que utilizan el token de autenticación para identificar al usuario
}

// setupPublicEnterpriseRoutes configura las rutas públicas para empresas
func setupPublicEnterpriseRoutes(router *mux.Router, enterpriseHandler *handlers.EnterpriseHandler) {
	router.HandleFunc(EnterprisesPath, enterpriseHandler.RegisterEnterprise).Methods(http.MethodPost)
}

// setupPublicCategoryRoutes configura las rutas públicas para categorías
func setupPublicCategoryRoutes(router *mux.Router, categoryHandler *handlers.CategoryHandler) {
	router.HandleFunc(CategoriesPath, categoryHandler.ListCategories).Methods(http.MethodGet)
}

// setupPublicMiscRoutes configura las rutas públicas para datos misceláneos
func setupPublicMiscRoutes(router *mux.Router, miscHandler *handlers.MiscHandler) {
	router.HandleFunc(NationalitiesPath, miscHandler.GetNationalities).Methods(http.MethodGet)
	router.HandleFunc(UniversitiesPath, miscHandler.GetUniversities).Methods(http.MethodGet)
	router.HandleFunc(DegreesPath, miscHandler.GetDegreesByUniversity).Methods(http.MethodGet)

	// TODO: Evaluar si estas rutas deberían requerir autenticación
	// Rutas comentadas pendientes de implementación:
	// - GET /languages - Falta implementar el handler
	// - GET /skills - Handler existe pero causa error
}

// setupProtectedRoute configura una ruta protegida individual recibiendo el handler y método
func setupProtectedRoute(router *mux.Router, path string, handler http.HandlerFunc, method string) {
	router.HandleFunc(path, handler).Methods(method)
}

// TODO: Añadir funciones para las siguientes rutas cuando se implementen:
// - setupProtectedEnterpriseRoutes: Gestión de empresas (listar, buscar, actualizar)
// - setupAdminRoutes: Rutas para administradores (gestión de roles, categorías, etc.)
// - setupInterestsRoutes: Gestión de intereses
// - setupMessagingRoutes: Gestión de tipos de mensajes
