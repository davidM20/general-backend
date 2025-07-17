package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

const jobApplicationServiceComponent = "JOB_APPLICATION_SERVICE"

// IJobApplication define la interfaz para el servicio de postulaciones.
type IJobApplication interface {
	ApplyToJob(eventID, applicantID int64, request models.JobApplicationCreateRequest) error
	ListApplicants(eventID int64) ([]models.ApplicantInfo, error)
	UpdateApplicationStatus(eventID, applicantID int64, newStatus string) error
}

var validStatuses = map[string]struct{}{
	"ENVIADA":          {},
	"EN_REVISION":      {},
	"ENTREVISTA":       {},
	"PRUEBA_TECNICA":   {},
	"OFERTA_REALIZADA": {},
	"APROBADA":         {},
	"RECHAZADA":        {},
	"RETIRADA":         {},
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
func (s *JobApplicationService) ListApplicants(eventID int64) ([]models.ApplicantInfo, error) {
	rows, err := s.db.Query(queries.ListApplicantsByEvent, eventID)
	if err != nil {
		logger.Errorf(jobApplicationServiceComponent, "Error al listar postulantes para el evento %d: %v", eventID, err)
		return nil, fmt.Errorf("error al consultar la base de datos: %w", err)
	}
	defer rows.Close()

	var applicants []models.ApplicantInfo
	for rows.Next() {
		var nullableApp struct {
			ApplicantID       int64
			FirstName         sql.NullString
			LastName          sql.NullString
			Email             sql.NullString
			AverageRating     sql.NullFloat64
			ReputationScore   sql.NullInt64
			ApplicationStatus sql.NullString
			AppliedAt         sql.NullTime
		}
		if err := rows.Scan(
			&nullableApp.ApplicantID,
			&nullableApp.FirstName,
			&nullableApp.LastName,
			&nullableApp.Email,
			&nullableApp.AverageRating,
			&nullableApp.ReputationScore,
			&nullableApp.ApplicationStatus,
			&nullableApp.AppliedAt,
		); err != nil {
			logger.Errorf(jobApplicationServiceComponent, "Error al escanear el perfil del postulante: %v", err)
			return nil, fmt.Errorf("error al procesar los resultados: %w", err)
		}

		app := models.ApplicantInfo{
			ApplicantID:       nullableApp.ApplicantID,
			FirstName:         nullableApp.FirstName.String,
			LastName:          nullableApp.LastName.String,
			Email:             nullableApp.Email.String,
			AverageRating:     nullableApp.AverageRating.Float64,
			ReputationScore:   int(nullableApp.ReputationScore.Int64),
			ApplicationStatus: nullableApp.ApplicationStatus.String,
			AppliedAt:         nullableApp.AppliedAt.Time,
		}
		applicants = append(applicants, app)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf(jobApplicationServiceComponent, "Error durante la iteración de los postulantes: %v", err)
		return nil, fmt.Errorf("error al leer los resultados: %w", err)
	}

	return applicants, nil
}

// UpdateApplicationStatus actualiza el estado de una postulación.
func (s *JobApplicationService) UpdateApplicationStatus(eventID, applicantID int64, newStatus string) error {
	// Validar que el estado sea uno de los permitidos por el ENUM de la BD.
	if _, ok := validStatuses[newStatus]; !ok {
		return fmt.Errorf("estado de postulación no válido: %s", newStatus)
	}

	result, err := s.db.Exec(queries.UpdateJobApplicationStatus, newStatus, eventID, applicantID)
	if err != nil {
		logger.Errorf(jobApplicationServiceComponent, "Error al actualizar estado de postulación para evento %d y aplicante %d: %v", eventID, applicantID, err)
		return fmt.Errorf("no se pudo actualizar el estado: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Infof(jobApplicationServiceComponent, "Advertencia: No se pudo obtener el número de filas afectadas: %v", err)
		return nil // No es un error fatal, la operación probablemente tuvo éxito.
	}

	if rowsAffected == 0 {
		return errors.New("no se encontró la postulación para actualizar o el estado ya era el mismo")
	}

	// TODO: Disparar una notificación al aplicante sobre el cambio de estado.
	logger.Successf(jobApplicationServiceComponent, "Estado de postulación actualizado a '%s' para evento %d y aplicante %d", newStatus, eventID, applicantID)
	return nil
}
