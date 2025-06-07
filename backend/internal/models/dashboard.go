package models

// DashboardCounts holds the basic counts for the dashboard.
type DashboardCounts struct {
	TotalRegisteredUsers int64 `json:"totalRegisteredUsers"`
	AdministrativeUsers  int64 `json:"administrativeUsers"`
	BusinessAccounts     int64 `json:"businessAccounts"`
	AlumniStudents       int64 `json:"alumniStudents"`
	EgresadoUsers        int64 `json:"egresadoUsers"`
}

// UserByCampus represents the number of users per campus.
type UserByCampus struct {
	Name  string `json:"name"`
	Users int64  `json:"users"`
}

// MonthlyActivity represents user registrations per month.
type MonthlyActivity struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}
