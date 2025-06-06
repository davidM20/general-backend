package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// CommunityEvent representa un evento comunitario que se muestra en el feed.
type CommunityEvent struct {
	Id                   int64           `json:"id"`
	Title                string          `json:"title"`
	Description          sql.NullString  `json:"description"` // Puede ser nulo
	EventDate            time.Time       `json:"event_date"`
	Location             sql.NullString  `json:"location"`               // Puede ser nulo
	Capacity             sql.NullInt64   `json:"capacity"`               // Nuevo
	Price                sql.NullFloat64 `json:"price"`                  // Nuevo
	Tags                 []string        `json:"tags"`                   // Nuevo, se manejará como JSON en DB
	OrganizerCompanyName sql.NullString  `json:"organizer_company_name"` // Puede ser nulo
	OrganizerUserId      sql.NullInt64   `json:"organizer_user_id"`      // Puede ser nulo, FK a User(Id)
	OrganizerLogoUrl     sql.NullString  `json:"organizer_logo_url"`     // Puede ser nulo
	ImageUrl             sql.NullString  `json:"image_url"`              // Puede ser nulo
	CreatedByUserId      int64           `json:"created_by_user_id"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// CommunityEventCreateRequest es el DTO para la creación de un CommunityEvent.
// No incluimos Id, CreatedAt, UpdatedAt ya que son generados por la DB o el sistema.
// CreatedByUserId se obtendrá del token del usuario autenticado.
type CommunityEventCreateRequest struct {
	Title                string    `json:"title" validate:"required,min=3,max=255"`
	Description          *string   `json:"description"`
	EventDate            time.Time `json:"event_date" validate:"required"`
	Location             *string   `json:"location"`
	Capacity             *int64    `json:"capacity" validate:"omitempty,min=0"`         // Nuevo
	Price                *float64  `json:"price" validate:"omitempty,min=0"`            // Nuevo
	Tags                 []string  `json:"tags" validate:"omitempty,dive,min=1,max=50"` // Nuevo
	OrganizerCompanyName *string   `json:"organizer_company_name"`
	OrganizerUserId      *int64    `json:"organizer_user_id"` // Opcional
	ImageUrl             *string   `json:"image_url"`
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
