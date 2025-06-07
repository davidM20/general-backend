package services

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

type SearchService interface {
	SearchAll(searchTerm string, limit, offset int) ([]wsmodels.SearchResultItem, error)
}

type searchService struct {
	db *sql.DB
}

func NewSearchService(db *sql.DB) SearchService {
	return &searchService{db: db}
}

func (s *searchService) SearchAll(searchTerm string, limit, offset int) ([]wsmodels.SearchResultItem, error) {
	// 1. Llamar a la consulta de la base de datos
	users, err := queries.SearchAll(searchTerm, limit, offset)
	if err != nil {
		logger.Errorf("SEARCH_SERVICE", "Error al buscar 'all': %v", err)
		return nil, fmt.Errorf("error al realizar la búsqueda: %w", err)
	}

	// 2. Mapear los resultados de la base de datos a los modelos de WebSocket
	results := make([]wsmodels.SearchResultItem, 0, len(users))
	for _, user := range users {
		item := s.mapUserToSearchResult(user)
		results = append(results, item)
	}

	return results, nil
}

func (s *searchService) mapUserToSearchResult(user models.User) wsmodels.SearchResultItem {
	item := wsmodels.SearchResultItem{
		ID: strconv.FormatInt(user.Id, 10),
	}

	if user.RoleId == 3 { // Empresa
		item.Type = "company"
		item.Data = wsmodels.CompanySearchResultData{
			Name:     user.UserName,
			Logo:     user.Picture.String,
			Industry: "", // Campo no disponible en el modelo User simplificado
			Location: "", // Campo no disponible
			Headline: user.Summary.String,
		}
	} else { // Estudiante o Egresado
		item.Type = "student"
		if user.RoleId == 2 {
			item.Type = "graduate"
		}
		item.Data = wsmodels.UserSearchResultData{
			Name:       user.FirstName + " " + user.LastName,
			Avatar:     user.Picture.String,
			Career:     "", // Campo no disponible
			University: "", // Campo no disponible
			Headline:   user.Summary.String,
		}
	}

	return item
}
