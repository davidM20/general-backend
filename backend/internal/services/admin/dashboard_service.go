package admin

import (
	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

const dashboardServiceLogComponent = "SERVICE_DASHBOARD"

// GetDashboardData retrieves and assembles all data needed for the admin dashboard.
func GetDashboardData(activeUsers int) (*wsmodels.DashboardDataPayload, error) {
	counts, err := queries.GetDashboardCounts()
	if err != nil {
		logger.Errorf(dashboardServiceLogComponent, "Failed to get dashboard counts: %v", err)
		return nil, err
	}

	usersByCampus, err := queries.GetUsersByCampus()
	if err != nil {
		logger.Errorf(dashboardServiceLogComponent, "Failed to get users by campus: %v", err)
		return nil, err
	}

	monthlyActivity, err := queries.GetMonthlyActivity()
	if err != nil {
		logger.Errorf(dashboardServiceLogComponent, "Failed to get monthly activity: %v", err)
		return nil, err
	}

	// Transform data into wsmodels types
	wsUsersByCampus := make([]wsmodels.UserByCampus, len(usersByCampus))
	for i, v := range usersByCampus {
		wsUsersByCampus[i] = wsmodels.UserByCampus{
			Name:  v.Name,
			Users: v.Users,
		}
	}

	wsMonthlyActivityLabels := make([]string, len(monthlyActivity))
	wsMonthlyActivityData := make([]int64, len(monthlyActivity))
	for i, v := range monthlyActivity {
		wsMonthlyActivityLabels[i] = v.Month
		wsMonthlyActivityData[i] = v.Count
	}

	payload := &wsmodels.DashboardDataPayload{
		ActiveUsers:          int64(activeUsers), // Passed in from the handler
		TotalRegisteredUsers: counts.TotalRegisteredUsers,
		AdministrativeUsers:  counts.AdministrativeUsers,
		BusinessAccounts:     counts.BusinessAccounts,
		AlumniStudents:       counts.AlumniStudents,
		EgresadoUsers:        counts.EgresadoUsers,
		AverageUsageTime:     "N/A", // Placeholder as discussed
		UsersByCampus:        wsUsersByCampus,
		MonthlyActivity: wsmodels.MonthlyActivity{
			Labels: wsMonthlyActivityLabels,
			Data:   wsMonthlyActivityData,
		},
	}

	logger.Successf(dashboardServiceLogComponent, "Successfully retrieved dashboard data")
	return payload, nil
}
