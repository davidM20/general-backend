package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/davidM20/micro-service-backend-go.git/internal/db"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	// "github.com/davidM20/micro-service-backend-go.git/internal/utils" // Import comentado
)

// --- Funciones de Utilidad (Temporales) ---

// RespondWithError sends a JSON error response.
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// RespondWithJSON sends a JSON response.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error":"Internal Server Error: %v"}`, err)))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// --- Category Handler ---

type CategoryHandler struct {
	// Podría tener dependencias como el logger o DB si no se usaran globales
}

func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

// ListCategories handles GET requests to list all categories.
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := db.GetAllCategories()
	if err != nil {
		log.Printf("Error getting categories from DB: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve categories")
		return
	}

	// Si no hay categorías, devolver lista vacía en lugar de error
	if categories == nil {
		categories = []models.Category{}
	}

	respondWithJSON(w, http.StatusOK, categories)
}

// AddCategoryRequest defines the expected JSON body for adding a category.
type AddCategoryRequest struct {
	Name string `json:"name"`
}

// AddCategory handles POST requests to add a new category.
func (h *CategoryHandler) AddCategory(w http.ResponseWriter, r *http.Request) {
	var req AddCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	categoryName := strings.TrimSpace(req.Name)
	if categoryName == "" {
		respondWithError(w, http.StatusBadRequest, "Category name cannot be empty")
		return
	}

	// Verificar si ya existe
	exists, err := db.CheckCategoryExistsByName(categoryName)
	if err != nil {
		log.Printf("Error checking category existence for '%s': %v", categoryName, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to check category existence")
		return
	}
	if exists {
		respondWithError(w, http.StatusConflict, "Category with this name already exists")
		return
	}

	// Añadir la categoría
	newCategory, err := db.AddCategory(categoryName)
	if err != nil {
		log.Printf("Error adding category '%s': %v", categoryName, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to add category")
		return
	}

	respondWithJSON(w, http.StatusCreated, newCategory)
}
