package routes

/*
 * ===================================================
 * GUÍA DE MANTENIMIENTO Y EXTENSIÓN DE RUTAS API
 * ===================================================
 *
 * ESTRUCTURA DEL ARCHIVO:
 * -----------------------
 * Este archivo, inspirado en la legibilidad y organización de frameworks como Gin,
 * define todas las rutas de la API REST del microservicio. La estructura sigue un
 * enfoque modular y agrupado por dominios para facilitar el mantenimiento.
 *
 * 1. Función principal `SetupApiRoutes` que orquesta la configuración de todos los grupos de rutas.
 * 2. Grupos de rutas claramente definidos:
 *    - Rutas públicas (`setupPublicRoutes`)
 *    - Rutas de streaming (`setupStreamingRoutes`)
 *    - Rutas protegidas con JWT (`setupProtectedRoutes`)
 *    - Rutas de administrador (`setupAdminRoutes`)
 * 3. Rutas definidas inline dentro de sub-routers para una máxima legibilidad (ej. `userRouter.HandleFunc("/me", ...)`).
 *
 * REGLAS DE NOMENCLATURA:
 * ----------------------
 * - Funciones de setup: `setup[Tipo][Dominio]Routes` (ej. `setupPublicAuthRoutes`, `setupUserProtectedRoutes`)
 * - Nombres de handlers: descriptivos del dominio que manejan.
 *
 * PROCESO PARA AÑADIR NUEVAS RUTAS:
 * --------------------------------
 * 1. Identifica el grupo al que pertenece la ruta (pública, protegida, admin).
 * 2. Navega a la función de configuración del grupo correspondiente (ej. `setupProtectedRoutes`).
 * 3. Añade la ruta en la función de dominio adecuada (ej. `setupUserProtectedRoutes`). Si el dominio
 *    es nuevo, crea una nueva función `setup[NuevoDominio]ProtectedRoutes` y llámala desde `setupProtectedRoutes`.
 * 4. Utiliza `router.HandleFunc("/mi-nueva-ruta", handler).Methods(http.MethodXXX)` para definir la ruta.
 *
 * PRINCIPIOS A SEGUIR:
 * ------------------
 * - Agrupación por dominio: Todas las rutas de un recurso (ej. "usuarios") deben estar juntas.
 * - Responsabilidad única: Cada función de setup tiene un propósito claro.
 * - Código limpio: Evitar duplicación y mantener la legibilidad.
 * - Documentación clara y concisa.
 */

import (
	"database/sql"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/config"     // Importar config
	"github.com/davidM20/micro-service-backend-go.git/internal/handlers"   // Crearemos este paquete
	"github.com/davidM20/micro-service-backend-go.git/internal/middleware" // Importar middleware
	"github.com/davidM20/micro-service-backend-go.git/internal/services"   // Necesario para inicializar ImageUploadService
	"github.com/gorilla/mux"
)

// Constantes para las rutas base
const (
	// Prefijos de rutas
	APIPrefix = "/api/v1"
)

// SetupApiRoutes configura todas las rutas para el microservicio de API REST
// siguiendo un enfoque modular inspirado en frameworks como Gin.
func SetupApiRoutes(r *mux.Router, db *sql.DB, cfg *config.Config) {
	// Crear instancias de los handlers
	handlers := initializeHandlers(db, cfg)

	// Crear subrouter para la API con prefijo /api/v1
	api := r.PathPrefix(APIPrefix).Subrouter()

	// Configurar grupos de rutas
	setupPublicRoutes(api, handlers)
	setupStreamingRoutes(api, handlers)
	setupProtectedRoutes(api, handlers, cfg)
	setupAdminRoutes(api, handlers.adminHandler, db, cfg)
}

// Estructura para agrupar todos los handlers y facilitar su paso a las funciones
type serviceHandlers struct {
	authHandler           *handlers.AuthHandler
	userHandler           *handlers.UserHandler
	enterpriseHandler     *handlers.EnterpriseHandler
	miscHandler           *handlers.MiscHandler
	mediaHandler          *handlers.MediaHandler
	categoryHandler       *handlers.CategoryHandler
	communityEventHandler *handlers.CommunityEventHandler
	imageHandler          *handlers.ImageHandler
	audioHandler          *handlers.AudioHandler
	pdfHandler            *handlers.PDFHandler
	videoHandler          *handlers.VideoHandler
	searchHandler         *handlers.SearchHandler
	adminHandler          *handlers.AdminHandler
	notificationHandler   *handlers.NotificationHandler
	jobApplicationHandler *handlers.JobApplicationHandler
	reputationHandler     *handlers.ReputationHandler
}

// initializeHandlers crea e inicializa todas las instancias de handlers necesarias
func initializeHandlers(db *sql.DB, cfg *config.Config) serviceHandlers {
	// Inicializar servicios primero si los handlers dependen de ellos
	imageUploadService := services.NewImageUploadService(db, cfg)
	audioUploadService := services.NewAudioUploadService(db, cfg)
	pdfUploadService := services.NewPDFUploadService(db, cfg)
	videoUploadService := services.NewVideoUploadService(db, cfg)
	searchService := services.NewSearchService(db)
	jobApplicationService := services.NewJobApplicationService(db)
	reputationService := services.NewReputationService(db)

	return serviceHandlers{
		authHandler:           handlers.NewAuthHandler(db, cfg),
		userHandler:           handlers.NewUserHandler(db),
		enterpriseHandler:     handlers.NewEnterpriseHandler(db),
		miscHandler:           handlers.NewMiscHandler(db),
		mediaHandler:          handlers.NewMediaHandler(db, cfg),
		categoryHandler:       handlers.NewCategoryHandler(),
		communityEventHandler: handlers.NewCommunityEventHandler(db, cfg),
		imageHandler:          handlers.NewImageHandler(imageUploadService, cfg),
		audioHandler:          handlers.NewAudioHandler(audioUploadService, cfg),
		pdfHandler:            handlers.NewPDFHandler(pdfUploadService, cfg),
		videoHandler:          handlers.NewVideoHandler(videoUploadService, db, cfg),
		searchHandler:         handlers.NewSearchHandler(searchService),
		adminHandler:          handlers.NewAdminHandler(db, cfg),
		notificationHandler:   handlers.NewNotificationHandler(db),
		jobApplicationHandler: handlers.NewJobApplicationHandler(jobApplicationService, db),
		reputationHandler:     handlers.NewReputationHandler(reputationService),
	}
}

// =================================================================================
// GRUPOS DE RUTAS
// =================================================================================

// ---------------------------------------------------------------------------------
// Rutas Públicas
// ---------------------------------------------------------------------------------

func setupPublicRoutes(api *mux.Router, h serviceHandlers) {
	setupHealthRoutes(api)
	setupPublicAuthRoutes(api, h.authHandler)
	setupPublicEnterpriseRoutes(api, h.enterpriseHandler)
	setupPublicCategoryRoutes(api, h.categoryHandler)
	setupPublicMiscRoutes(api, h.miscHandler)
}

// setupHealthRoutes configura las rutas de verificación de estado del sistema
func setupHealthRoutes(router *mux.Router) {
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API is healthy"))
	}).Methods(http.MethodGet)
}

// setupPublicAuthRoutes configura las rutas públicas de autenticación y registro
func setupPublicAuthRoutes(router *mux.Router, authHandler *handlers.AuthHandler) {
	// Grupo para registro
	registerRouter := router.PathPrefix("/register").Subrouter()
	{
		registerRouter.HandleFunc("", authHandler.Register).Methods(http.MethodPost)
		registerRouter.HandleFunc("/company", authHandler.RegisterCompany).Methods(http.MethodPost)
	}

	// Ruta de autenticación (Login)
	router.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)

	// Grupo para recuperación de contraseña
	resetPasswordRouter := router.PathPrefix("/reset-password").Subrouter()
	{
		resetPasswordRouter.HandleFunc("/request", authHandler.RequestPasswordReset).Methods(http.MethodPost)
		resetPasswordRouter.HandleFunc("/complete", authHandler.CompletePasswordReset).Methods(http.MethodPost)
	}
}

// setupPublicEnterpriseRoutes configura las rutas públicas para empresas
func setupPublicEnterpriseRoutes(router *mux.Router, enterpriseHandler *handlers.EnterpriseHandler) {
	router.HandleFunc("/enterprises", enterpriseHandler.RegisterEnterprise).Methods(http.MethodPost)
}

// setupPublicCategoryRoutes configura las rutas públicas para categorías
func setupPublicCategoryRoutes(router *mux.Router, categoryHandler *handlers.CategoryHandler) {
	router.HandleFunc("/categories", categoryHandler.ListCategories).Methods(http.MethodGet)
}

// setupPublicMiscRoutes configura las rutas públicas para datos misceláneos
func setupPublicMiscRoutes(router *mux.Router, miscHandler *handlers.MiscHandler) {
	router.HandleFunc("/nationalities", miscHandler.GetNationalities).Methods(http.MethodGet)
	router.HandleFunc("/universities", miscHandler.GetUniversities).Methods(http.MethodGet)
	router.HandleFunc("/degrees/{universityID:[0-9]+}", miscHandler.GetDegreesByUniversity).Methods(http.MethodGet)

	// TODO: Evaluar si estas rutas deberían requerir autenticación
	// Rutas comentadas pendientes de implementación:
	// - GET /languages - Falta implementar el handler
	// - GET /skills - Handler existe pero causa error
}

// ---------------------------------------------------------------------------------
// Rutas de Streaming y Visualización (autenticación por token en query param)
// ---------------------------------------------------------------------------------

func setupStreamingRoutes(api *mux.Router, h serviceHandlers) {
	// Grupo para visualización de imágenes
	imageRouter := api.PathPrefix("/images").Subrouter()
	{
		imageRouter.HandleFunc("/view/{filename}", h.imageHandler.ViewImage).Methods(http.MethodGet)
	}

	// Grupo para visualización de audio
	audioRouter := api.PathPrefix("/audios").Subrouter()
	{
		audioRouter.HandleFunc("/view/{filename}", h.audioHandler.ViewAudio).Methods(http.MethodGet)
	}

	// Grupo para visualización de PDFs
	pdfRouter := api.PathPrefix("/pdfs").Subrouter()
	{
		pdfRouter.HandleFunc("/view/{filename}", h.pdfHandler.ViewPDF).Methods(http.MethodGet)
	}

	// Grupo para streaming de video
	videoRouter := api.PathPrefix("/videos/stream").Subrouter()
	{
		videoRouter.HandleFunc("/{contentID}/master.m3u8", h.videoHandler.StreamVideoMasterPlaylist).Methods(http.MethodGet)
		videoRouter.HandleFunc("/{contentID}/{quality}/{fileName:.+}", h.videoHandler.StreamVideoVariant).Methods(http.MethodGet)
	}

	// Ruta para ver foto de perfil de usuario
	api.HandleFunc("/users/{userID:[0-9]+}/picture", h.imageHandler.ViewUserProfilePicture).Methods(http.MethodGet)
}

// ---------------------------------------------------------------------------------
// Rutas Protegidas (requieren token JWT)
// ---------------------------------------------------------------------------------

func setupProtectedRoutes(api *mux.Router, h serviceHandlers, cfg *config.Config) {
	protected := api.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(cfg))

	// Agrupar por dominio para mayor claridad
	setupAuthProtectedRoutes(protected, h.authHandler)
	setupUserProtectedRoutes(protected, h.userHandler, h.imageHandler)
	setupEnterpriseProtectedRoutes(protected, h.enterpriseHandler)
	setupCategoryProtectedRoutes(protected, h.categoryHandler)
	setupMediaProtectedRoutes(protected, h)
	setupCommunityEventsProtectedRoutes(protected, h.communityEventHandler)
	setupJobApplicationProtectedRoutes(protected, h.jobApplicationHandler)
	setupReputationProtectedRoutes(protected, h.reputationHandler)
	setupNotificationProtectedRoutes(protected, h.notificationHandler)
	setupSearchProtectedRoutes(protected, h.searchHandler)
}

// setupAuthProtectedRoutes configura las rutas protegidas de registro (pasos 2 y 3)
func setupAuthProtectedRoutes(router *mux.Router, authHandler *handlers.AuthHandler) {
	registerRouter := router.PathPrefix("/register").Subrouter()
	{
		registerRouter.HandleFunc("/step2", authHandler.RegisterStep2).Methods(http.MethodPost)
		registerRouter.HandleFunc("/step3", authHandler.RegisterStep3).Methods(http.MethodPost)
	}
}

// setupUserProtectedRoutes configura las rutas protegidas del perfil de usuario
func setupUserProtectedRoutes(router *mux.Router, userHandler *handlers.UserHandler, imageHandler *handlers.ImageHandler) {
	userRouter := router.PathPrefix("/users").Subrouter()
	{
		meRouter := userRouter.PathPrefix("/me").Subrouter()
		meRouter.HandleFunc("", userHandler.GetMyProfile).Methods(http.MethodGet)
		meRouter.HandleFunc("", userHandler.UpdateMyProfile).Methods(http.MethodPut)
		meRouter.HandleFunc("/picture", imageHandler.UpdateProfilePicture).Methods(http.MethodPost)
	}
}

// setupEnterpriseProtectedRoutes configura las rutas protegidas para empresas
func setupEnterpriseProtectedRoutes(router *mux.Router, enterpriseHandler *handlers.EnterpriseHandler) {
	enterpriseRouter := router.PathPrefix("/enterprises").Subrouter()
	{
		enterpriseRouter.HandleFunc("/me", enterpriseHandler.UpdateEnterpriseProfile).Methods(http.MethodPut)
	}
}

// setupCategoryProtectedRoutes configura las rutas protegidas para categorías
func setupCategoryProtectedRoutes(router *mux.Router, categoryHandler *handlers.CategoryHandler) {
	router.HandleFunc("/categories", categoryHandler.AddCategory).Methods(http.MethodPost)
}

// setupMediaProtectedRoutes configura las rutas protegidas para subida de multimedia
func setupMediaProtectedRoutes(router *mux.Router, h serviceHandlers) {
	router.HandleFunc("/media/upload", h.mediaHandler.UploadMedia).Methods(http.MethodPost)
	router.HandleFunc("/images/upload", h.imageHandler.UploadImage).Methods(http.MethodPost)
	router.HandleFunc("/audios/upload", h.audioHandler.UploadAudio).Methods(http.MethodPost)
	router.HandleFunc("/pdfs/upload", h.pdfHandler.UploadPDF).Methods(http.MethodPost)
	router.HandleFunc("/videos/upload", h.videoHandler.UploadVideo).Methods(http.MethodPost)
}

// setupCommunityEventsProtectedRoutes configura las rutas protegidas para eventos comunitarios
func setupCommunityEventsProtectedRoutes(router *mux.Router, communityEventHandler *handlers.CommunityEventHandler) {
	communityEventsRouter := router.PathPrefix("/community-events").Subrouter()
	{
		communityEventsRouter.HandleFunc("", communityEventHandler.CreateCommunityEvent).Methods(http.MethodPost)
		communityEventsRouter.HandleFunc("/my-events", communityEventHandler.GetMyCommunityEvents).Methods(http.MethodGet)
	}
}

// setupJobApplicationProtectedRoutes configura las rutas protegidas para postulaciones
func setupJobApplicationProtectedRoutes(router *mux.Router, jobApplicationHandler *handlers.JobApplicationHandler) {
	// Grupo de rutas bajo /community-events/{eventID}/applicants
	applicantsRouter := router.PathPrefix("/community-events/{eventID:[0-9]+}").Subrouter()
	{
		applicantsRouter.HandleFunc("/apply", jobApplicationHandler.ApplyToJob).Methods(http.MethodPost)
		applicantsRouter.HandleFunc("/applicants", jobApplicationHandler.ListApplicants).Methods(http.MethodGet)
		applicantsRouter.HandleFunc("/applicants/{applicantID:[0-9]+}/status", jobApplicationHandler.UpdateApplicationStatus).Methods(http.MethodPatch)
	}
}

// setupReputationProtectedRoutes configura las rutas protegidas para reseñas y reputación
func setupReputationProtectedRoutes(router *mux.Router, reputationHandler *handlers.ReputationHandler) {
	reviewsRouter := router.PathPrefix("/reviews").Subrouter()
	{
		reviewsRouter.HandleFunc("", reputationHandler.CreateReview).Methods(http.MethodPost)
		reviewsRouter.HandleFunc("/student", reputationHandler.CreateReviewByStudent).Methods(http.MethodPost)
	}
}

// setupNotificationProtectedRoutes configura las rutas protegidas para notificaciones
func setupNotificationProtectedRoutes(router *mux.Router, notificationHandler *handlers.NotificationHandler) {
	notificationRouter := router.PathPrefix("/notifications").Subrouter()
	{
		notificationRouter.HandleFunc("/{notificationID:[0-9]+}/read", notificationHandler.MarkAsRead).Methods(http.MethodPut)
	}
}

// setupSearchProtectedRoutes configura las rutas protegidas para búsqueda
func setupSearchProtectedRoutes(router *mux.Router, searchHandler *handlers.SearchHandler) {
	searchRouter := router.PathPrefix("/search").Subrouter()
	{
		searchRouter.HandleFunc("/talent", searchHandler.SearchTalent).Methods(http.MethodGet)
	}
}

// ---------------------------------------------------------------------------------
// Rutas de Administrador
// ---------------------------------------------------------------------------------

// setupAdminRoutes configura las rutas que requieren privilegios de administrador.
// Aplica tanto el middleware de autenticación como el de verificación de rol de administrador.
func setupAdminRoutes(router *mux.Router, adminHandler *handlers.AdminHandler, db *sql.DB, cfg *config.Config) {
	adminRouter := router.PathPrefix("/admin").Subrouter()

	// Cadena de middlewares: primero autenticación, luego validación de rol y sesión de admin.
	adminRouter.Use(middleware.AuthMiddleware(cfg))
	adminRouter.Use(middleware.AdminMiddleware(db))

	adminRouter.HandleFunc("/dashboard", adminHandler.GetDashboard).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users", adminHandler.ListUsers).Methods(http.MethodGet)
	adminRouter.HandleFunc("/companies/unapproved", adminHandler.ListUnapprovedCompanies).Methods(http.MethodGet)

	// TODO: Implementar los siguientes handlers y rutas
	// adminRouter.HandleFunc("/users/{id}", adminHandler.ManageUser).Methods(http.MethodPut, http.MethodDelete)
	// adminRouter.HandleFunc("/categories", adminHandler.ManageCategories).Methods(http.MethodPost, http.MethodPut)
}
