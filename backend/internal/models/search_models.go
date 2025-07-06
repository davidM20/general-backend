package models

// UniversalSearchParams contiene todos los parámetros posibles para la búsqueda,
// combinando la búsqueda fonética por texto con filtros estructurados.
type UniversalSearchParams struct {
	Query                string
	Career               string
	University           string
	GraduationYear       int
	IsCurrentlyStudying  *bool
	IsCurrentlyWorking   *bool
	Location             string
	Skills               []string
	Languages            []string
	YearsOfExperienceMin int
	YearsOfExperienceMax int
	Page                 int
	Limit                int
}

// SearchResultProfile representa un perfil de usuario simplificado para los resultados de búsqueda.
type SearchResultProfile struct {
	ID                int64    `json:"id"`
	FirstName         string   `json:"first_name"`
	LastName          string   `json:"last_name"`
	Picture           *string  `json:"picture"`
	Summary           *string  `json:"summary"`
	Location          *string  `json:"location"`
	Career            *string  `json:"career,omitempty"`
	YearsOfExperience *float64 `json:"years_of_experience,omitempty"`
	TotalReputation   *int     `json:"total_reputation"`
	AverageRating     *float64 `json:"average_rating"`
	RoleId            int64    `json:"role_id"`
}

// UniversalSearchResponse es la estructura de respuesta que combina usuarios y eventos.
type UniversalSearchResponse struct {
	Users              []SearchResultProfile `json:"users"`
	Events             []CommunityEvent      `json:"events"`
	Pagination         PaginationDetails     `json:"pagination"`
	YearsDistribution  []YearsDistribution   `json:"years_distribution"`
	CareerDistribution []CareerDistribution  `json:"career_distribution"`
}

// YearsDistribution representa la cuenta de usuarios por años de experiencia.
type YearsDistribution struct {
	Years int `json:"years"`
	Count int `json:"count"`
}

// CareerDistribution representa la cuenta de usuarios por carrera.
type CareerDistribution struct {
	Career string `json:"career"`
	Count  int    `json:"count"`
}

// PaginatedTalentResponse es una respuesta paginada solo para usuarios.
// Mantenemos el nombre original por si se usa en otras partes, pero representa usuarios.
type PaginatedTalentResponse struct {
	Data       []SearchResultProfile `json:"data"`
	Pagination PaginationDetails     `json:"pagination"`
}

// PaginationDetails contiene la información de paginación.
type PaginationDetails struct {
	TotalItems  int `json:"total_items"`
	TotalPages  int `json:"total_pages"`
	CurrentPage int `json:"current_page"`
	PageSize    int `json:"page_size"`
}
