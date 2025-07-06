package models

import (
	"time"
)

// JobApplication representa una postulación a un evento/oferta.
type JobApplication struct {
	Id               int64     `json:"id"`
	CommunityEventId int64     `json:"community_event_id"`
	ApplicantId      int64     `json:"applicant_id"`
	Status           string    `json:"status"`
	AppliedAt        time.Time `json:"applied_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	CoverLetter      string    `json:"cover_letter,omitempty"`
}

// JobApplicationCreateRequest define los datos necesarios para crear una postulación.
type JobApplicationCreateRequest struct {
	CoverLetter *string `json:"cover_letter"`
}

// ApplicantProfile define la estructura de un postulante listado para una oferta.
type ApplicantProfile struct {
	ApplicantId       int64     `json:"applicant_id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Email             string    `json:"email"`
	AverageRating     float64   `json:"average_rating"`
	ReputationScore   int       `json:"reputation_score"`
	ApplicationStatus string    `json:"application_status"`
	AppliedAt         time.Time `json:"applied_at"`
}
