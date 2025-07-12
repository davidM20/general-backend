package models

import "time"

// ApplicantInfo representa la información detallada de un postulante.
type ApplicantInfo struct {
	ApplicantID       int64     `json:"applicantId"`
	FirstName         string    `json:"firstName"`
	LastName          string    `json:"lastName"`
	Email             string    `json:"email"`
	AverageRating     float64   `json:"averageRating"`
	ReputationScore   int       `json:"reputationScore"`
	ApplicationStatus string    `json:"applicationStatus"`
	AppliedAt         time.Time `json:"appliedAt"`
}

// JobApplicationCreateRequest define la estructura para crear una nueva postulación.
type JobApplicationCreateRequest struct {
	CoverLetter *string `json:"cover_letter"`
}

// UpdateApplicationStatusRequest define el cuerpo de la petición para cambiar el estado de una postulación.
type UpdateApplicationStatusRequest struct {
	Status string `json:"status"`
}
