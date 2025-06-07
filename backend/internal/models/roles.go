package models

// UserRole represents the role of a user in the system.
type UserRole int

const (
	RoleStudent  UserRole = 1
	RoleEgresado UserRole = 2
	RoleBusiness UserRole = 3
	RoleGuest    UserRole = 4
	RoleAdmin    UserRole = 8
)
