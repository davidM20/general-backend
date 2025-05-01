package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"log"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/gorilla/mux"
)

// MiscHandler maneja peticiones para obtener datos generales
type MiscHandler struct {
	DB *sql.DB
}

// NewMiscHandler crea una nueva instancia de MiscHandler
func NewMiscHandler(db *sql.DB) *MiscHandler {
	return &MiscHandler{DB: db}
}

// GetNationalities devuelve la lista de nacionalidades
func (h *MiscHandler) GetNationalities(w http.ResponseWriter, r *http.Request) {
	nationalities := models.GetDefaultNationalities() // Obtener desde los datos por defecto
	// Opcionalmente, podrías leerlos de la tabla Nationality si prefieres gestionarlos en BD
	/*
	   rows, err := h.DB.Query("SELECT Id, CountryName, IsoCode, DocIdFormat FROM Nationality ORDER BY CountryName")
	   if err != nil {
	       log.Printf("Error querying nationalities: %v", err)
	       http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
	       return
	   }
	   defer rows.Close()

	   nationalities := []models.Nationality{}
	   for rows.Next() {
	       var nat models.Nationality
	       if err := rows.Scan(&nat.Id, &nat.CountryName, &nat.IsoCode, &nat.DocIdFormat); err != nil {
	           log.Printf("Error scanning nationality row: %v", err)
	           continue // O manejar el error de otra forma
	       }
	       nationalities = append(nationalities, nat)
	   }
	   if err = rows.Err(); err != nil {
	        log.Printf("Error iterating nationality rows: %v", err)
	        http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
	        return
	   }
	*/

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nationalities)
}

// GetUniversities devuelve la lista de universidades
func (h *MiscHandler) GetUniversities(w http.ResponseWriter, r *http.Request) {
	// Leer desde la base de datos
	rows, err := h.DB.Query("SELECT Id, Name, Campus FROM University ORDER BY Name")
	if err != nil {
		log.Printf("Error querying universities: %v", err)
		http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	universities := []models.University{}
	for rows.Next() {
		var uni models.University
		if err := rows.Scan(&uni.Id, &uni.Name, &uni.Campus); err != nil {
			log.Printf("Error scanning university row: %v", err)
			continue
		}
		universities = append(universities, uni)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating university rows: %v", err)
		http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(universities)
}

// GetDegreesByUniversity devuelve la lista de carreras para una universidad específica
func (h *MiscHandler) GetDegreesByUniversity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	universityIDStr := vars["universityID"]
	universityID, err := strconv.ParseInt(universityIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid university ID", http.StatusBadRequest)
		return
	}

	rows, err := h.DB.Query("SELECT Id, DegreeName, Descriptions, Code FROM Degree WHERE UniversityId = ? ORDER BY DegreeName", universityID)
	if err != nil {
		log.Printf("Error querying degrees for university %d: %v", universityID, err)
		http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	degrees := []models.Degree{}
	for rows.Next() {
		var deg models.Degree
		// Omitimos UniversityId al escanear ya que lo tenemos del path
		if err := rows.Scan(&deg.Id, &deg.DegreeName, &deg.Descriptions, &deg.Code); err != nil {
			log.Printf("Error scanning degree row: %v", err)
			continue
		}
		deg.UniversityId = universityID // Asignar el ID conocido
		degrees = append(degrees, deg)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating degree rows: %v", err)
		http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(degrees)
}

// GetCategories devuelve la lista de categorías
func (h *MiscHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT CategoryId, Name, Description FROM Category ORDER BY Name")
	if err != nil {
		log.Printf("Error querying categories: %v", err)
		http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	categories := []models.Category{}
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.CategoryId, &cat.Name, &cat.Description); err != nil {
			log.Printf("Error scanning category row: %v", err)
			continue
		}
		categories = append(categories, cat)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating category rows: %v", err)
		http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(categories)
}
