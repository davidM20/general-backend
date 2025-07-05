package services

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/davidM20/micro-service-backend-go.git/pkg/phonetic"
)

type ISearchService interface {
	UniversalSearch(ctx context.Context, params models.UniversalSearchParams) (*models.UniversalSearchResponse, error)
}

type SearchService struct {
	db *sql.DB
}

func NewSearchService(db *sql.DB) ISearchService {
	return &SearchService{
		db: db,
	}
}

func (s *SearchService) UniversalSearch(ctx context.Context, params models.UniversalSearchParams) (*models.UniversalSearchResponse, error) {
	var userConditions, eventConditions []string
	var userArgs, eventArgs []interface{}
	likeQuery := "%" + params.Query + "%"

	// Aplicar búsqueda por texto si existe el parámetro 'q'
	if params.Query != "" {
		primaryKey, secondaryKey, err := phonetic.GenerateKeysForPhrase(params.Query)
		if err != nil {
			logger.Errorf("SEARCH_SERVICE", "Error generating phonetic keys for query '%s': %v", params.Query, err)
			return nil, fmt.Errorf("could not process search query")
		}

		var userTextSearchConditions []string
		if primaryKey != "" {
			// Búsqueda fonética
			userTextSearchConditions = append(userTextSearchConditions, "u.dmeta_person_primary LIKE ? OR u.dmeta_person_secondary LIKE ? OR u.dmeta_company_primary LIKE ? OR u.dmeta_company_secondary LIKE ?")
			userArgs = append(userArgs, primaryKey+"%", secondaryKey+"%", primaryKey+"%", secondaryKey+"%")

			eventConditions = append(eventConditions, "(ce.dmeta_title_primary LIKE ? OR ce.dmeta_title_secondary LIKE ?)")
			eventArgs = append(eventArgs, primaryKey+"%", secondaryKey+"%")
		}
		// Búsqueda LIKE tradicional
		userTextSearchConditions = append(userTextSearchConditions, "u.FirstName LIKE ? OR u.LastName LIKE ? OR u.UserName LIKE ? OR u.RIF LIKE ? OR EXISTS (SELECT 1 FROM Education e WHERE e.PersonId = u.Id AND e.Degree LIKE ?)")
		userArgs = append(userArgs, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery)

		userConditions = append(userConditions, "("+strings.Join(userTextSearchConditions, " OR ")+")")
	}

	// === CONSTRUIR FILTROS DE USUARIO ===
	if params.Career != "" {
		userConditions = append(userConditions, "EXISTS (SELECT 1 FROM Education e WHERE e.PersonId = u.Id AND e.Degree LIKE ?)")
		userArgs = append(userArgs, "%"+params.Career+"%")
	}
	if params.University != "" {
		userConditions = append(userConditions, "EXISTS (SELECT 1 FROM Education e WHERE e.PersonId = u.Id AND e.Institution LIKE ?)")
		userArgs = append(userArgs, "%"+params.University+"%")
	}
	if params.Location != "" {
		userConditions = append(userConditions, "u.Location LIKE ?")
		userArgs = append(userArgs, "%"+params.Location+"%")
	}
	// ... (Aquí se pueden añadir más filtros de usuario como skills, graduation_year, etc.)

	// === CONSTRUIR FILTROS DE EVENTOS ===
	if params.Location != "" {
		eventConditions = append(eventConditions, "ce.Location LIKE ?")
		eventArgs = append(eventArgs, "%"+params.Location+"%")
	}

	// === CONSTRUIR CONSULTAS FINALES ===
	userQuery := "SELECT 'user' as type, u.Id, u.CreatedAt FROM User u"
	if len(userConditions) > 0 {
		userQuery += " WHERE " + strings.Join(userConditions, " AND ")
	}

	eventQuery := "SELECT 'event' as type, ce.Id, ce.CreatedAt FROM CommunityEvent ce"
	if len(eventConditions) > 0 {
		eventQuery += " WHERE " + strings.Join(eventConditions, " AND ")
	}

	// Si no hay filtros ni query, no devolver nada.
	if len(userConditions) == 0 && len(eventConditions) == 0 {
		return &models.UniversalSearchResponse{Pagination: models.PaginationDetails{CurrentPage: params.Page, PageSize: params.Limit}}, nil
	}

	// Consulta de conteo
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ((%s) UNION ALL (%s)) as combined", userQuery, eventQuery)
	countArgs := append(userArgs, eventArgs...)
	var totalItems int
	if err := s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalItems); err != nil {
		logger.Errorf("SEARCH_SERVICE", "Error counting combined results: %v", err)
		return nil, fmt.Errorf("error counting results: %w", err)
	}

	if totalItems == 0 {
		return &models.UniversalSearchResponse{Pagination: models.PaginationDetails{TotalItems: 0, CurrentPage: params.Page, PageSize: params.Limit}}, nil
	}

	// Consulta de datos paginados
	fullQuery := fmt.Sprintf("(%s) UNION ALL (%s) ORDER BY CreatedAt DESC LIMIT ? OFFSET ?", userQuery, eventQuery)
	offset := (params.Page - 1) * params.Limit
	queryArgs := append(append(userArgs, eventArgs...), params.Limit, offset)

	rows, err := s.db.QueryContext(ctx, fullQuery, queryArgs...)
	if err != nil {
		logger.Errorf("SEARCH_SERVICE", "Error executing combined search: %v", err)
		return nil, fmt.Errorf("error executing search: %w", err)
	}
	defer rows.Close()

	var users []models.SearchResultProfile
	var events []models.CommunityEvent
	for rows.Next() {
		var resultType string
		var id int64
		var createdAt sql.NullTime
		if err := rows.Scan(&resultType, &id, &createdAt); err != nil {
			logger.Errorf("SEARCH_SERVICE", "Error scanning combined result row: %v", err)
			continue
		}

		if resultType == "user" {
			profile, err := queries.GetUserProfileByID(s.db, id)
			if err != nil {
				logger.Warnf("SEARCH_SERVICE", "Could not fetch full profile for user ID %d: %v", id, err)
				continue
			}
			users = append(users, *profile)
		} else if resultType == "event" {
			event, err := queries.GetCommunityEventByID(s.db, id)
			if err != nil {
				logger.Warnf("SEARCH_SERVICE", "Could not fetch full event for event ID %d: %v", id, err)
				continue
			}
			events = append(events, *event)
		}
	}

	return &models.UniversalSearchResponse{
		Users:  users,
		Events: events,
		Pagination: models.PaginationDetails{
			TotalItems:  totalItems,
			TotalPages:  int(math.Ceil(float64(totalItems) / float64(params.Limit))),
			CurrentPage: params.Page,
			PageSize:    params.Limit,
		},
	}, nil
}
