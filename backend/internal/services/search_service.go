package services

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"sync"

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
		userTextSearchConditions = append(userTextSearchConditions, "(u.FirstName LIKE ? OR u.LastName LIKE ? OR u.UserName LIKE ? OR u.RIF LIKE ? OR u.CompanyName LIKE ?)")
		userArgs = append(userArgs, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery)

		// Búsqueda en educación (solo para talentos)
		userTextSearchConditions = append(userTextSearchConditions, "EXISTS (SELECT 1 FROM Education e WHERE e.PersonId = u.Id AND e.Degree LIKE ?)")
		userArgs = append(userArgs, likeQuery)

		userConditions = append(userConditions, "("+strings.Join(userTextSearchConditions, " OR ")+")")
	}

	// Si se aplica un filtro que es exclusivo para talento, la búsqueda se centrará solo en usuarios.
	isTalentOnlySearch := params.Role != "" ||
		params.Career != "" ||
		params.University != "" ||
		params.GraduationYear != 0 ||
		params.IsCurrentlyStudying != nil ||
		params.IsCurrentlyWorking != nil ||
		len(params.Skills) > 0 ||
		len(params.Languages) > 0 ||
		params.YearsOfExperienceMin > 0 ||
		params.YearsOfExperienceMax > 0

	// === CONSTRUIR FILTROS DE USUARIO ===
	if params.Role != "" {
		var roleId int
		switch params.Role {
		case "student":
			roleId = 1
		case "graduate":
			roleId = 2
		}

		if roleId > 0 {
			userConditions = append(userConditions, "u.RoleId = ?")
			userArgs = append(userArgs, roleId)
		}
	}

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

	// === CONSTRUIR FILTROS DE EVENTOS (solo si no es una búsqueda de talento) ===
	if !isTalentOnlySearch {
		if params.Location != "" {
			eventConditions = append(eventConditions, "ce.Location LIKE ?")
			eventArgs = append(eventArgs, "%"+params.Location+"%")
		}
	} else {
		// Nos aseguramos de que los argumentos de eventos estén vacíos para no contaminar la consulta final.
		eventArgs = []interface{}{}
	}

	// === CONSTRUIR CONSULTAS FINALES ===
	userQuery := "SELECT 'user' as type, u.Id, u.CreatedAt, u.RoleId FROM User u"
	if len(userConditions) > 0 {
		userQuery += " WHERE " + strings.Join(userConditions, " AND ")
	}

	var eventQuery string
	if !isTalentOnlySearch {
		eventQuery = "SELECT 'event' as type, ce.Id, ce.CreatedAt, NULL as RoleId FROM CommunityEvent ce"
		if len(eventConditions) > 0 {
			eventQuery += " WHERE " + strings.Join(eventConditions, " AND ")
		}
	}

	// Si no hay filtros ni query, no devolver nada.
	if len(userConditions) == 0 && len(eventConditions) == 0 && !isTalentOnlySearch {
		return &models.UniversalSearchResponse{Pagination: models.PaginationDetails{CurrentPage: params.Page, PageSize: params.Limit}}, nil
	}
	// Si es una búsqueda de talento pero no hay filtros, tampoco hay nada que hacer.
	if isTalentOnlySearch && len(userConditions) == 0 {
		return &models.UniversalSearchResponse{Pagination: models.PaginationDetails{CurrentPage: params.Page, PageSize: params.Limit}}, nil
	}

	var wg sync.WaitGroup
	var errChan = make(chan error, 3)
	var response models.UniversalSearchResponse

	// Goroutine 1: Obtener resultados paginados
	wg.Add(1)
	go func() {
		defer wg.Done()
		paginatedUsers, paginatedCompanies, paginatedEvents, pagination, err := s.fetchPaginatedResults(ctx, userQuery, eventQuery, userArgs, eventArgs, params)
		if err != nil {
			errChan <- fmt.Errorf("error fetching paginated results: %w", err)
			return
		}
		response.Users = paginatedUsers
		response.Companies = paginatedCompanies
		response.Events = paginatedEvents
		response.Pagination = pagination
	}()

	// Goroutine 2: Obtener distribución de carreras
	wg.Add(1)
	go func() {
		defer wg.Done()
		careerDist, err := s.fetchCareerDistribution(ctx, userConditions, userArgs)
		if err != nil {
			errChan <- fmt.Errorf("error fetching career distribution: %w", err)
			return
		}
		response.CareerDistribution = careerDist
	}()

	// Goroutine 3: Obtener distribución de años de experiencia
	wg.Add(1)
	go func() {
		defer wg.Done()
		yearsDist, err := s.fetchYearsDistribution(ctx, userConditions, userArgs)
		if err != nil {
			errChan <- fmt.Errorf("error fetching years distribution: %w", err)
			return
		}
		response.YearsDistribution = yearsDist
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			// Devolver el primer error que ocurra
			return nil, err
		}
	}

	return &response, nil
}

func (s *SearchService) fetchPaginatedResults(ctx context.Context, userQuery, eventQuery string, userArgs, eventArgs []interface{}, params models.UniversalSearchParams) ([]models.SearchResultProfile, []models.SearchResultProfile, []models.CommunityEvent, models.PaginationDetails, error) {
	// Implementación de conteo y obtención de resultados paginados (lógica que ya teníamos)
	var countQuery, fullQuery string
	var countArgs, queryArgs []interface{}

	if eventQuery != "" {
		countQuery = fmt.Sprintf("SELECT COUNT(*) FROM ((%s) UNION ALL (%s)) as combined", userQuery, eventQuery)
		countArgs = append(userArgs, eventArgs...)
		fullQuery = fmt.Sprintf("(%s) UNION ALL (%s) ORDER BY CreatedAt DESC LIMIT ? OFFSET ?", userQuery, eventQuery)
	} else {
		countQuery = fmt.Sprintf("SELECT COUNT(*) FROM (%s) as combined", userQuery)
		countArgs = userArgs
		fullQuery = fmt.Sprintf("%s ORDER BY CreatedAt DESC LIMIT ? OFFSET ?", userQuery)
	}

	var totalItems int
	if err := s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalItems); err != nil {
		logger.Errorf("SEARCH_SERVICE", "Error counting combined results: %v", err)
		return nil, nil, nil, models.PaginationDetails{}, fmt.Errorf("error counting results: %w", err)
	}

	pagination := models.PaginationDetails{
		TotalItems:  totalItems,
		TotalPages:  int(math.Ceil(float64(totalItems) / float64(params.Limit))),
		CurrentPage: params.Page,
		PageSize:    params.Limit,
	}

	if totalItems == 0 {
		return []models.SearchResultProfile{}, []models.SearchResultProfile{}, []models.CommunityEvent{}, pagination, nil
	}

	offset := (params.Page - 1) * params.Limit
	if eventQuery != "" {
		queryArgs = append(append(userArgs, eventArgs...), params.Limit, offset)
	} else {
		queryArgs = append(userArgs, params.Limit, offset)
	}

	rows, err := s.db.QueryContext(ctx, fullQuery, queryArgs...)
	if err != nil {
		logger.Errorf("SEARCH_SERVICE", "Error executing combined search: %v", err)
		return nil, nil, nil, pagination, fmt.Errorf("error executing search: %w", err)
	}
	defer rows.Close()

	var users []models.SearchResultProfile
	var companies []models.SearchResultProfile
	var events []models.CommunityEvent
	for rows.Next() {
		var resultType string
		var id int64
		var createdAt sql.NullTime
		var roleId sql.NullInt64
		if err := rows.Scan(&resultType, &id, &createdAt, &roleId); err != nil {
			logger.Errorf("SEARCH_SERVICE", "Error scanning combined result row: %v", err)
			continue
		}

		if resultType == "user" {
			profile, err := queries.GetUserProfileByID(s.db, id)
			if err != nil {
				logger.Warnf("SEARCH_SERVICE", "Could not fetch full profile for user ID %d: %v", id, err)
				continue
			}
			if roleId.Valid && roleId.Int64 == 3 {
				companies = append(companies, *profile)
			} else {
				users = append(users, *profile)
			}
		} else if resultType == "event" {
			event, err := queries.GetCommunityEventByID(s.db, id)
			if err != nil {
				logger.Warnf("SEARCH_SERVICE", "Could not fetch full event for event ID %d: %v", id, err)
				continue
			}
			events = append(events, *event)
		}
	}
	return users, companies, events, pagination, nil
}

func (s *SearchService) fetchCareerDistribution(ctx context.Context, userConditions []string, userArgs []interface{}) ([]models.CareerDistribution, error) {
	query := `
		SELECT e.Degree, COUNT(DISTINCT u.Id) as count
		FROM User u
		JOIN Education e ON u.Id = e.PersonId
	`
	if len(userConditions) > 0 {
		query += " WHERE " + strings.Join(userConditions, " AND ")
	}
	query += " GROUP BY e.Degree ORDER BY count DESC LIMIT 10" // Limitar a los 10 más comunes

	rows, err := s.db.QueryContext(ctx, query, userArgs...)
	if err != nil {
		logger.Errorf("SEARCH_SERVICE", "Error fetching career distribution: %v", err)
		return nil, err
	}
	defer rows.Close()

	var distribution []models.CareerDistribution
	for rows.Next() {
		var item models.CareerDistribution
		if err := rows.Scan(&item.Career, &item.Count); err != nil {
			logger.Errorf("SEARCH_SERVICE", "Error scanning career distribution row: %v", err)
			continue
		}
		distribution = append(distribution, item)
	}
	return distribution, nil
}

func (s *SearchService) fetchYearsDistribution(ctx context.Context, userConditions []string, userArgs []interface{}) ([]models.YearsDistribution, error) {
	query := `
		SELECT FLOOR(TotalExperienceYears) as year, COUNT(*) as count
		FROM (
			SELECT u.Id, SUM(DATEDIFF(IF(we.IsCurrentJob, CURDATE(), we.EndDate), we.StartDate)) / 365.25 AS TotalExperienceYears
			FROM User u
			JOIN WorkExperience we ON u.Id = we.PersonId
	`
	if len(userConditions) > 0 {
		query += " WHERE " + strings.Join(userConditions, " AND ")
	}
	query += " GROUP BY u.Id) AS exp_by_user WHERE TotalExperienceYears IS NOT NULL GROUP BY year ORDER BY year ASC"

	rows, err := s.db.QueryContext(ctx, query, userArgs...)
	if err != nil {
		logger.Errorf("SEARCH_SERVICE", "Error fetching years distribution: %v", err)
		return nil, err
	}
	defer rows.Close()

	var distribution []models.YearsDistribution
	for rows.Next() {
		var item models.YearsDistribution
		if err := rows.Scan(&item.Years, &item.Count); err != nil {
			logger.Errorf("SEARCH_SERVICE", "Error scanning years distribution row: %v", err)
			continue
		}
		distribution = append(distribution, item)
	}
	return distribution, nil
}
