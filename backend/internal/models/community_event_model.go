package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// CommunityEvent representa la estructura de un evento en la base de datos.
type CommunityEvent struct {
	Id                   int64           `json:"id"`
	Title                string          `json:"title"`
	Description          NullString      `json:"description"`
	EventDate            time.Time       `json:"event_date"`
	Location             NullString      `json:"location"`
	Capacity             NullInt32       `json:"capacity"`
	Price                NullFloat64     `json:"price"`
	Tags                 json.RawMessage `json:"tags"`
	OrganizerCompanyName NullString      `json:"organizer_company_name"`
	OrganizerUserId      NullInt64       `json:"organizer_user_id"`
	OrganizerLogoUrl     NullString      `json:"organizer_logo_url"`
	ImageUrl             NullString      `json:"image_url"`
	CreatedByUserId      int64           `json:"created_by_user_id"`
	DmetaTitlePrimary    string          `json:"-"`
	DmetaTitleSecondary  string          `json:"-"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// CommunityEventCreateRequest representa los datos para crear un nuevo evento.
type CommunityEventCreateRequest struct {
	Title                string          `json:"title"`
	Description          string          `json:"description"`
	EventDate            string          `json:"event_date"`
	Location             string          `json:"location"`
	Capacity             *int            `json:"capacity"`
	Price                *float64        `json:"price"`
	Tags                 json.RawMessage `json:"tags"`
	OrganizerCompanyName string          `json:"organizer_company_name"`
	OrganizerUserId      *int64          `json:"organizer_user_id"`
	OrganizerLogoUrl     string          `json:"organizer_logo_url"`
	ImageUrl             string          `json:"image_url"`
}

// PaginatedCommunityEvents es la estructura para la respuesta paginada de eventos.
type PaginatedCommunityEvents struct {
	Data       []CommunityEvent  `json:"data"`
	Pagination PaginationDetails `json:"pagination"`
}

// ToNullString convierte un puntero a string en sql.NullString.
func ToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// ToNullInt64 convierte un puntero a int64 en sql.NullInt64.
func ToNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

// ToNullFloat64 convierte un puntero a float64 en sql.NullFloat64.
func ToNullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

// TagsToJSON convierte un slice de strings a una representación JSON (string).
// Devuelve sql.NullString porque el campo Tags puede ser nulo.
func TagsToJSON(tags []string) (sql.NullString, error) {
	if len(tags) == 0 {
		return sql.NullString{}, nil // Si no hay tags, se considera nulo en la DB
	}
	jsonData, err := json.Marshal(tags)
	if err != nil {
		return sql.NullString{}, err
	}
	return sql.NullString{String: string(jsonData), Valid: true}, nil
}

// TagsFromJSON convierte una representación JSON (string) a un slice de strings.
func TagsFromJSON(jsonData sql.NullString) ([]string, error) {
	if !jsonData.Valid || jsonData.String == "" {
		return []string{}, nil // Si es nulo o vacío en la DB, devuelve un slice vacío
	}
	var tags []string
	err := json.Unmarshal([]byte(jsonData.String), &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}
