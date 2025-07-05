package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/services"
)

type SearchHandler struct {
	service services.ISearchService
}

func NewSearchHandler(s services.ISearchService) *SearchHandler {
	return &SearchHandler{
		service: s,
	}
}

// SearchTalent ahora realiza una búsqueda fonética universal en usuarios y eventos,
// con la capacidad de ser refinada por filtros estructurados.
func (h *SearchHandler) SearchTalent(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()

	page, err := strconv.Atoi(queryValues.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(queryValues.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 10 // Límite por defecto
	}

	params := models.UniversalSearchParams{
		Query:      queryValues.Get("q"),
		Career:     queryValues.Get("career"),
		University: queryValues.Get("university"),
		Location:   queryValues.Get("location"),
		Page:       page,
		Limit:      limit,
	}

	if years, err := strconv.Atoi(queryValues.Get("years_of_experience_min")); err == nil {
		params.YearsOfExperienceMin = years
	}
	if years, err := strconv.Atoi(queryValues.Get("years_of_experience_max")); err == nil {
		params.YearsOfExperienceMax = years
	}
	if graduationYear, err := strconv.Atoi(queryValues.Get("graduation_year")); err == nil {
		params.GraduationYear = graduationYear
	}
	if isStudying, err := strconv.ParseBool(queryValues.Get("is_currently_studying")); err == nil {
		params.IsCurrentlyStudying = &isStudying
	}
	if isWorking, err := strconv.ParseBool(queryValues.Get("is_currently_working")); err == nil {
		params.IsCurrentlyWorking = &isWorking
	}
	if skills := queryValues.Get("skills"); skills != "" {
		params.Skills = strings.Split(skills, ",")
	}
	if languages := queryValues.Get("languages"); languages != "" {
		params.Languages = strings.Split(languages, ",")
	}

	results, err := h.service.UniversalSearch(r.Context(), params)
	if err != nil {
		http.Error(w, "Error searching: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Error encoding results: "+err.Error(), http.StatusInternalServerError)
	}
}
