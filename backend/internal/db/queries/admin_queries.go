package queries

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

const adminQueriesLogComponent = "QUERIES_ADMIN"

// GetDashboardCounts retrieves the main counters for the admin dashboard.
func GetDashboardCounts() (*models.DashboardCounts, error) {
	var counts models.DashboardCounts

	query := `
		SELECT
			(SELECT COUNT(*) FROM User) AS total_users,
			(SELECT COUNT(*) FROM User WHERE RoleId = ?) AS admin_users,
			(SELECT COUNT(*) FROM User WHERE RoleId = ?) AS business_users,
			(SELECT COUNT(*) FROM User WHERE RoleId = ?) AS alumni_students_users,
			(SELECT COUNT(*) FROM User WHERE RoleId = ?) AS egresado_users
	`

	err := DB.QueryRow(query, models.RoleAdmin, models.RoleBusiness, models.RoleStudent, models.RoleEgresado).Scan(
		&counts.TotalRegisteredUsers,
		&counts.AdministrativeUsers,
		&counts.BusinessAccounts,
		&counts.AlumniStudents,
		&counts.EgresadoUsers,
	)

	if err != nil {
		logger.Errorf(adminQueriesLogComponent, "Error querying dashboard counts: %v", err)
		return nil, fmt.Errorf("error querying dashboard counts: %w", err)
	}

	return &counts, nil
}

// GetUsersByCampus retrieves the count of users for each university campus.
func GetUsersByCampus() ([]models.UserByCampus, error) {
	query := `
		SELECT 
			COALESCE(u.Campus, 'Desconocido') as campus, 
			COUNT(usr.Id) as user_count
		FROM User usr
		LEFT JOIN University u ON usr.UniversityId = u.Id
		GROUP BY campus
		ORDER BY user_count DESC;
	`

	rows, err := DB.Query(query)
	if err != nil {
		logger.Errorf(adminQueriesLogComponent, "Error querying users by campus: %v", err)
		return nil, fmt.Errorf("error querying users by campus: %w", err)
	}
	defer rows.Close()

	var results []models.UserByCampus
	for rows.Next() {
		var result models.UserByCampus
		if err := rows.Scan(&result.Name, &result.Users); err != nil {
			logger.Errorf(adminQueriesLogComponent, "Error scanning user by campus row: %v", err)
			return nil, fmt.Errorf("error scanning user by campus row: %w", err)
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf(adminQueriesLogComponent, "Error after iterating users by campus rows: %v", err)
		return nil, fmt.Errorf("error after iterating users by campus rows: %w", err)
	}

	return results, nil
}

// GetMonthlyActivity retrieves the number of new user registrations over the last 12 months.
func GetMonthlyActivity() ([]models.MonthlyActivity, error) {
	// Generar los últimos 12 meses para el eje X
	months := []string{}
	for i := 11; i >= 0; i-- {
		months = append(months, time.Now().AddDate(0, -i, 0).Format("Jan"))
	}

	query := `
		SELECT
			DATE_FORMAT(CreatedAt, '%Y-%m') AS month,
			COUNT(Id) AS count
		FROM User
		WHERE CreatedAt >= ?
		GROUP BY month
		ORDER BY month ASC;
	`

	twelveMonthsAgo := time.Now().AddDate(0, -11, 0)
	// Nos aseguramos de tomar desde el inicio de ese mes
	startOfMonth := time.Date(twelveMonthsAgo.Year(), twelveMonthsAgo.Month(), 1, 0, 0, 0, 0, twelveMonthsAgo.Location())

	rows, err := DB.Query(query, startOfMonth)
	if err != nil {
		logger.Errorf(adminQueriesLogComponent, "Error querying monthly activity: %v", err)
		return nil, fmt.Errorf("error querying monthly activity: %w", err)
	}
	defer rows.Close()

	activityMap := make(map[string]int64)
	for rows.Next() {
		var month string
		var count int64
		if err := rows.Scan(&month, &count); err != nil {
			logger.Errorf(adminQueriesLogComponent, "Error scanning monthly activity row: %v", err)
			return nil, fmt.Errorf("error scanning monthly activity row: %w", err)
		}
		activityMap[month] = count
	}

	if err = rows.Err(); err != nil {
		logger.Errorf(adminQueriesLogComponent, "Error after iterating monthly activity rows: %v", err)
		return nil, fmt.Errorf("error after iterating monthly activity rows: %w", err)
	}

	var results []models.MonthlyActivity
	now := time.Now()
	for i := 11; i >= 0; i-- {
		date := now.AddDate(0, -i, 0)
		monthKey := date.Format("2006-01") // Formato YYYY-MM que coincide con la query
		monthName := date.Format("Jan")    // Formato "Ene", "Feb", etc. para el label
		count, ok := activityMap[monthKey]
		if !ok {
			count = 0
		}
		results = append(results, models.MonthlyActivity{Month: monthName, Count: count})
	}

	return results, nil
}

// CountTotalUsers cuenta el número total de usuarios registrados.
func CountTotalUsers() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM User"
	err := DB.QueryRow(query).Scan(&count)
	if err != nil {
		logger.Errorf(adminQueriesLogComponent, "Error counting total users: %v", err)
		return 0, fmt.Errorf("error counting total users: %w", err)
	}
	return count, nil
}

// GetUsersPaginated recupera una lista paginada de usuarios.
// Devuelve una lista de usuarios y un error.
func GetUsersPaginated(page, pageSize int) ([]models.UserDTO, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT
			u.Id, u.FirstName, u.LastName, u.UserName, u.Email, u.Phone,
			u.Picture, u.RoleId, r.Name as RoleName, u.StatusAuthorizedId, s.Name as StatusName,
			u.CreatedAt, u.UpdatedAt
		FROM User u
		LEFT JOIN Role r ON u.RoleId = r.Id
		LEFT JOIN StatusAuthorized s ON u.StatusAuthorizedId = s.Id
		ORDER BY u.Id ASC
		LIMIT ? OFFSET ?
	`

	rows, err := DB.Query(query, pageSize, offset)
	if err != nil {
		logger.Errorf(adminQueriesLogComponent, "Error querying paginated users: %v", err)
		return nil, fmt.Errorf("error querying paginated users: %w", err)
	}
	defer rows.Close()

	var users []models.UserDTO
	for rows.Next() {
		var user models.UserDTO
		var firstName, lastName, phone, picture, roleName, statusName sql.NullString

		if err := rows.Scan(
			&user.Id, &firstName, &lastName, &user.UserName, &user.Email, &phone,
			&picture, &user.RoleId, &roleName, &user.StatusAuthorizedId, &statusName,
			&user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			logger.Errorf(adminQueriesLogComponent, "Error scanning user row: %v", err)
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}

		// Asignar valores desde tipos Null a string
		user.FirstName = firstName.String
		user.LastName = lastName.String
		user.Phone = phone.String
		user.Picture = picture.String
		user.RoleName = roleName.String
		user.StatusName = statusName.String

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf(adminQueriesLogComponent, "Error after iterating user rows: %v", err)
		return nil, fmt.Errorf("error after iterating user rows: %w", err)
	}

	return users, nil
}

// CountUnapprovedCompanies cuenta el número total de empresas pendientes de aprobación.
func CountUnapprovedCompanies() (int, error) {
	var count int
	// StatusAuthorizedId = 1 significa 'No Autorizado' o 'Pendiente'
	query := "SELECT COUNT(*) FROM User WHERE RoleId = ? AND StatusAuthorizedId = 1"
	err := DB.QueryRow(query, models.RoleBusiness).Scan(&count)
	if err != nil {
		logger.Errorf(adminQueriesLogComponent, "Error counting unapproved companies: %v", err)
		return 0, fmt.Errorf("error counting unapproved companies: %w", err)
	}
	return count, nil
}

// GetUnapprovedCompaniesPaginated recupera una lista paginada de empresas pendientes de aprobación.
func GetUnapprovedCompaniesPaginated(page, pageSize int) ([]models.CompanyApprovalDTO, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT
			u.Id, u.CompanyName, u.RIF, u.Email, u.FirstName, u.Phone, s.Name as StatusName, u.CreatedAt
		FROM User u
		LEFT JOIN StatusAuthorized s ON u.StatusAuthorizedId = s.Id
		WHERE u.RoleId = ? AND u.StatusAuthorizedId = 1
		ORDER BY u.CreatedAt ASC
		LIMIT ? OFFSET ?
	`
	rows, err := DB.Query(query, models.RoleBusiness, pageSize, offset)
	if err != nil {
		logger.Errorf(adminQueriesLogComponent, "Error querying unapproved companies: %v", err)
		return nil, fmt.Errorf("error querying unapproved companies: %w", err)
	}
	defer rows.Close()

	var companies []models.CompanyApprovalDTO
	for rows.Next() {
		var company models.CompanyApprovalDTO
		// Usamos sql.NullString para campos que podrían ser NULL en la BD aunque el DTO los espere como string
		var companyName, rif, contactName, phone, statusName sql.NullString

		if err := rows.Scan(
			&company.Id, &companyName, &rif, &company.Email, &contactName, &phone, &statusName, &company.CreatedAt,
		); err != nil {
			logger.Errorf(adminQueriesLogComponent, "Error scanning unapproved company row: %v", err)
			return nil, fmt.Errorf("error scanning unapproved company row: %w", err)
		}

		company.CompanyName = companyName.String
		company.RIF = rif.String
		company.ContactName = contactName.String
		company.Phone = phone.String
		company.StatusName = statusName.String

		companies = append(companies, company)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf(adminQueriesLogComponent, "Error after iterating unapproved company rows: %v", err)
		return nil, fmt.Errorf("error after iterating unapproved company rows: %w", err)
	}

	return companies, nil
}
