package models

import "time"

// REGLAS A SEGUIR EN ESTE ARCHIVO
// 1. Todas las estructuras deben ser específicas para la vista de perfil de empresa.
// 2. Mantener consistencia con la nomenclatura de la base de datos y los DTOs del frontend.
// 3. Documentar cada struct para aclarar su propósito.

// CompanyProfile define la estructura de datos para el perfil de una empresa.
type CompanyProfile struct {
	Id                 int64     `json:"Id"`
	CompanyName        string    `json:"CompanyName"`
	Email              string    `json:"Email"`
	ContactEmail       string    `json:"ContactEmail,omitempty"`
	RIF                string    `json:"RIF"`
	Sector             string    `json:"Sector"`
	Location           string    `json:"Location"`
	Address            string    `json:"Address,omitempty"`
	FoundationYear     *int      `json:"FoundationYear,omitempty"`
	EmployeeCount      *int      `json:"EmployeeCount,omitempty"`
	Summary            string    `json:"Summary,omitempty"`
	Phone              string    `json:"Phone,omitempty"`
	Github             string    `json:"Github,omitempty"`
	Linkedin           string    `json:"Linkedin,omitempty"`
	Twitter            string    `json:"Twitter,omitempty"`
	Facebook           string    `json:"Facebook,omitempty"`
	Picture            string    `json:"Picture,omitempty"`
	RoleId             int       `json:"RoleId"`
	StatusAuthorizedId int       `json:"StatusAuthorizedId,omitempty"`
	CreatedAt          time.Time `json:"CreatedAt"`
	UpdatedAt          time.Time `json:"UpdatedAt"`
}

// CompanyEvent representa un evento creado por una empresa.
type CompanyEvent struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventDate   NullTime  `json:"event_date,omitempty"`
	Location    string    `json:"location"`
	ImageURL    string    `json:"image_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CompanyStats contiene estadísticas calculadas sobre la actividad de una empresa.
type CompanyStats struct {
	TotalEvents       int `json:"totalEvents"`
	UpcomingEvents    int `json:"upcomingEvents"`
	YearsOfExperience int `json:"yearsOfExperience"`
	EmployeeCount     int `json:"employeeCount"`
}

// CompleteCompanyProfile es la estructura completa que se envía al frontend.
type CompleteCompanyProfile struct {
	Company CompanyProfile `json:"company"`
	Events  []CompanyEvent `json:"events"`
	Stats   CompanyStats   `json:"stats"`
}

// EnterpriseProfileUpdate define los campos que una empresa puede actualizar.
// Se usan punteros para que los campos no proporcionados en el JSON no se actualicen.
type EnterpriseProfileUpdate struct {
	CompanyName    *string `json:"companyName,omitempty"`
	ContactEmail   *string `json:"contactEmail,omitempty"`
	Twitter        *string `json:"twitter,omitempty"`
	Facebook       *string `json:"facebook,omitempty"`
	Phone          *string `json:"phone,omitempty"`
	Picture        *string `json:"picture,omitempty"`
	Summary        *string `json:"summary,omitempty"`
	Address        *string `json:"address,omitempty"`
	Github         *string `json:"github,omitempty"`
	Linkedin       *string `json:"linkedin,omitempty"`
	Sector         *string `json:"sector,omitempty"`
	Location       *string `json:"location,omitempty"`
	FoundationYear *int    `json:"foundationYear,omitempty"`
	EmployeeCount  *int    `json:"employeeCount,omitempty"`
}
