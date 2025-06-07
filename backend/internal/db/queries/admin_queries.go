package queries

import (
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
