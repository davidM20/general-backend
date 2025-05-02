package routes

import (
	"database/sql"
	"net/http"

	"github.com/davidM20/micro-service-backend-go.git/internal/config"     // Importar config
	"github.com/davidM20/micro-service-backend-go.git/internal/handlers"   // Crearemos este paquete
	"github.com/davidM20/micro-service-backend-go.git/internal/middleware" // Importar middleware
	"github.com/gorilla/mux"
)

// SetupApiRoutes configura todas las rutas para el microservicio de API REST
func SetupApiRoutes(r *mux.Router, db *sql.DB, cfg *config.Config) {

	// Crear instancias de los handlers pasando la conexión DB y config
	authHandler := handlers.NewAuthHandler(db, cfg)
	userHandler := handlers.NewUserHandler(db)
	enterpriseHandler := handlers.NewEnterpriseHandler(db)
	miscHandler := handlers.NewMiscHandler(db)
	mediaHandler := handlers.NewMediaHandler(db, cfg)
	categoryHandler := handlers.NewCategoryHandler()

	// Crear subrouter para la API, quizás con prefijo /api/v1
	api := r.PathPrefix("/api/v1").Subrouter()

	// --- Rutas Públicas (sin autenticación) ---
	api.HandleFunc("/categories", categoryHandler.ListCategories).Methods(http.MethodGet)

	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API is healthy"))
	}).Methods(http.MethodGet)

	// Registro (en pasos)
	api.HandleFunc("/register/step1", authHandler.RegisterStep1).Methods(http.MethodPost)
	// Para step 2 y 3, necesitaremos identificar al usuario parcialmente registrado.
	// Podríamos devolver un token temporal después del step 1.
	// Por ahora, asumiremos que el ID del usuario se pasa de alguna manera (ej. en el path o token temporal)
	// TODO: Implementar lógica de token temporal o similar para registro en pasos
	api.HandleFunc("/register/step2/{userID:[0-9]+}", authHandler.RegisterStep2).Methods(http.MethodPost)
	api.HandleFunc("/register/step3/{userID:[0-9]+}", authHandler.RegisterStep3).Methods(http.MethodPost)

	api.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)

	// Endpoints para obtener datos generales (requieren autenticación? Depende)
	// Podrían ser públicos o requerir autenticación básica/invitado
	// TODO: Decidir y aplicar middleware de autenticación si es necesario
	api.HandleFunc("/nationalities", miscHandler.GetNationalities).Methods(http.MethodGet)
	api.HandleFunc("/universities", miscHandler.GetUniversities).Methods(http.MethodGet)
	api.HandleFunc("/degrees/{universityID:[0-9]+}", miscHandler.GetDegreesByUniversity).Methods(http.MethodGet)
	// api.HandleFunc("/languages", miscHandler.GetLanguages).Methods(http.MethodGet) // Comentado - No existe?
	// api.HandleFunc("/skills", miscHandler.GetSkills).Methods(http.MethodGet) // Comentado - Causa error

	// --- Rutas Protegidas (requieren autenticación JWT) ---
	protected := api.PathPrefix("/").Subrouter() // Subrouter para aplicar middleware
	// Aplicar el middleware de autenticación JWT a este subrouter
	protected.Use(middleware.AuthMiddleware(cfg)) // Pasar la configuración

	// Todas las rutas definidas en 'protected' ahora requerirán un token JWT válido
	protected.HandleFunc("/categories", categoryHandler.AddCategory).Methods(http.MethodPost)
	protected.HandleFunc("/users/me", userHandler.GetMyProfile).Methods(http.MethodGet) // Ver mi propio perfil
	// protected.HandleFunc("/users/{userID:[0-9]+}", userHandler.GetUserProfile).Methods(http.MethodGet) // Ver perfil de otro (si permitido)
	// PUT /users/me para actualizar perfil (parcialmente, o todo?) - Podría ser WS

	// Empresas
	protected.HandleFunc("/enterprises", enterpriseHandler.RegisterEnterprise).Methods(http.MethodPost)
	// GET /enterprises (listar/buscar) - Podría ser WS
	// GET /enterprises/{enterpriseID:[0-9]+} (ver detalle) - Podría ser WS
	// PUT /enterprises/{enterpriseID:[0-9]+} (actualizar) - Podría ser WS

	// Subida de Multimedia (protegida)
	protected.HandleFunc("/media/upload", mediaHandler.UploadMedia).Methods(http.MethodPost)

	// TODO: Añadir rutas para:
	// - Gestión de Roles/Status (admin)
	// - Intereses
	// - Creación/Gestión de Categorías (admin?)
	// - Creación/Gestión de Tipos de Mensaje (admin?)
	// - Endpoints específicos que no vayan por WebSocket
}
