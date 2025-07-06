package services

import (
	"database/sql"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

const jobApplicationServiceComponent = "JOB_APPLICATION_SERVICE"

// IJobApplication define la interfaz para el servicio de postulaciones.
type IJobApplication interface {
	ApplyToJob(eventID, applicantID int64, request models.JobApplicationCreateRequest) error
	ListApplicants(eventID int64) ([]models.ApplicantProfile, error)
	// TODO: Añadir métodos para cambiar estado, notificar, etc.
}

// JobApplicationService implementa la lógica de negocio para las postulaciones.
type JobApplicationService struct {
	db *sql.DB
}

// NewJobApplicationService crea una nueva instancia de JobApplicationService.
func NewJobApplicationService(db *sql.DB) *JobApplicationService {
	return &JobApplicationService{db: db}
}

// ApplyToJob permite a un usuario postularse a una oferta.
func (s *JobApplicationService) ApplyToJob(eventID, applicantID int64, request models.JobApplicationCreateRequest) error {
	_, err := s.db.Exec(queries.CreateJobApplication, eventID, applicantID, request.CoverLetter)
	if err != nil {
		logger.Errorf(jobApplicationServiceComponent, "Error al crear la postulación para el evento %d por el aplicante %d: %v", eventID, applicantID, err)
		return fmt.Errorf("no se pudo crear la postulación: %w", err)
	}
	// TODO: Aquí se podría disparar una notificación al creador de la oferta.
	logger.Successf(jobApplicationServiceComponent, "Postulación creada exitosamente para el evento %d por el aplicante %d", eventID, applicantID)
	return nil
}

// ListApplicants devuelve la lista de postulantes para una oferta, ordenada por reputación.
func (s *JobApplicationService) ListApplicants(eventID int64) ([]models.ApplicantProfile, error) {
	rows, err := s.db.Query(queries.ListApplicantsByEvent, eventID)
	if err != nil {
		logger.Errorf(jobApplicationServiceComponent, "Error al listar postulantes para el evento %d: %v", eventID, err)
		return nil, fmt.Errorf("error al consultar la base de datos: %w", err)
	}
	defer rows.Close()

	var applicants []models.ApplicantProfile
	for rows.Next() {
		var app models.ApplicantProfile
		if err := rows.Scan(
			&app.ApplicantId,
			&app.FirstName,
			&app.LastName,
			&app.Email,
			&app.AverageRating,
			&app.ReputationScore,
			&app.ApplicationStatus,
			&app.AppliedAt,
		); err != nil {
			logger.Errorf(jobApplicationServiceComponent, "Error al escanear el perfil del postulante: %v", err)
			return nil, fmt.Errorf("error al procesar los resultados: %w", err)
		}
		applicants = append(applicants, app)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf(jobApplicationServiceComponent, "Error durante la iteración de los postulantes: %v", err)
		return nil, fmt.Errorf("error al leer los resultados: %w", err)
	}

	return applicants, nil
}
