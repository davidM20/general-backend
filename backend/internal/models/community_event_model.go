package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// NullTime representa un time.Time que puede ser nulo, para compatibilidad con JSON.
type NullTime struct {
	sql.NullTime
}

// MarshalJSON para NullTime
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time)
}

// UnmarshalJSON para NullTime
func (nt *NullTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nt.Valid = false
		return nil
	}
	var t time.Time
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	nt.Time = t
	nt.Valid = true
	return nil
}

// CommunityEvent representa la estructura de una publicación en el feed.
// Puede ser un evento, noticia, artículo, etc., diferenciado por PostType.
type CommunityEvent struct {
	Id                     int64       `json:"id"`
	PostType               string      `json:"post_type"` // 'EVENTO', 'NOTICIA', 'ARTICULO', etc.
	Title                  string      `json:"title"`
	Description            NullString  `json:"description,omitempty"`
	ImageUrl               NullString  `json:"image_url,omitempty"`
	ContentUrl             NullString  `json:"content_url,omitempty"`
	LinkPreviewTitle       NullString  `json:"link_preview_title,omitempty"`
	LinkPreviewDescription NullString  `json:"link_preview_description,omitempty"`
	LinkPreviewImage       NullString  `json:"link_preview_image,omitempty"`
	EventDate              NullTime    `json:"event_date,omitempty"` // Modificado para ser nullable
	Location               NullString  `json:"location,omitempty"`
	Capacity               NullInt32   `json:"capacity,omitempty"`
	Price                  NullFloat64 `json:"price,omitempty"`

	// --- NUEVOS CAMPOS PARA DESAFÍOS ---
	ChallengeStartDate  NullTime   `json:"challenge_start_date,omitempty"`
	ChallengeEndDate    NullTime   `json:"challenge_end_date,omitempty"`
	ChallengeDifficulty NullString `json:"challenge_difficulty,omitempty"`
	ChallengePrize      NullString `json:"challenge_prize,omitempty"`
	ChallengeStatus     string     `json:"challenge_status"`

	// --- CAMPOS COMUNES ---
	Tags                 json.RawMessage `json:"tags,omitempty"`
	OrganizerCompanyName NullString      `json:"organizer_company_name,omitempty"`
	OrganizerUserId      NullInt64       `json:"organizer_user_id,omitempty"`
	OrganizerLogoUrl     NullString      `json:"organizer_logo_url,omitempty"`
	CreatedByUserId      int64           `json:"-"`
	DmetaTitlePrimary    string          `json:"-"`
	DmetaTitleSecondary  string          `json:"-"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// CommunityEventCreateRequest representa los datos para crear una nueva publicación en el feed.
type CommunityEventCreateRequest struct {
	PostType               string   `json:"post_type"` // 'EVENTO', 'NOTICIA', 'ARTICULO', 'ANUNCIO', 'MULTIMEDIA', 'DESAFIO' etc.. Requerido.
	Title                  string   `json:"title"`     // Requerido.
	Description            *string  `json:"description,omitempty"`
	ImageUrl               *string  `json:"image_url,omitempty"`
	ContentUrl             *string  `json:"content_url,omitempty"`
	LinkPreviewTitle       *string  `json:"link_preview_title,omitempty"`
	LinkPreviewDescription *string  `json:"link_preview_description,omitempty"`
	LinkPreviewImage       *string  `json:"link_preview_image,omitempty"`
	EventDate              *string  `json:"event_date,omitempty"` // Formato "YYYY-MM-DD HH:MM:SS"
	Location               *string  `json:"location,omitempty"`
	Capacity               *int32   `json:"capacity,omitempty"`
	Price                  *float64 `json:"price,omitempty"`

	// --- NUEVOS CAMPOS PARA DESAFÍOS ---
	ChallengeStartDate  *string `json:"challenge_start_date,omitempty"`
	ChallengeEndDate    *string `json:"challenge_end_date,omitempty"`
	ChallengeDifficulty *string `json:"challenge_difficulty,omitempty"`
	ChallengePrize      *string `json:"challenge_prize,omitempty"`

	// --- CAMPOS COMUNES ---
	Tags                 json.RawMessage `json:"tags,omitempty"`
	OrganizerCompanyName *string         `json:"organizer_company_name,omitempty"`
	OrganizerUserId      *int64          `json:"organizer_user_id,omitempty"`
	OrganizerLogoUrl     *string         `json:"organizer_logo_url,omitempty"`
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
