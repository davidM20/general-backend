package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

const reputationServiceComponent = "REPUTATION_SERVICE"

// IReputationService define la interfaz para el servicio de reputación.
type IReputationService interface {
	CreateReview(reviewerID int64, req models.CreateReviewRequest) error
}

// ReputationService implementa la lógica de negocio para el sistema de reputación.
type ReputationService struct {
	db *sql.DB
}

// NewReputationService crea una nueva instancia de ReputationService.
func NewReputationService(db *sql.DB) IReputationService {
	return &ReputationService{db: db}
}

// CreateReview gestiona la creación de una nueva reseña, calculando los RP
// y guardando el registro en la base de datos.
func (s *ReputationService) CreateReview(reviewerID int64, req models.CreateReviewRequest) error {
	if req.Rating < 0 || req.Rating > 5 {
		return errors.New("la calificación debe estar entre 0 y 5")
	}

	pointsRP := s.convertStarsToRP(req.Rating)

	// Aplicar el bono si la condición se cumple.
	// El bono de "3 estrellas extra" equivale a 25 RP.
	if req.ApplyBonus {
		pointsRP += 25
	}

	// TODO: Idealmente, esto usaría sqlc para una mayor seguridad de tipos.
	// Por ahora, se usa una consulta directa.
	query := `
        INSERT INTO ReputationReview (ReviewerId, RevieweeId, PointsRP, Rating, Comment, InteractionType)
        VALUES (?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, reviewerID, req.RevieweeID, pointsRP, req.Rating, req.Comment, req.InteractionType)
	if err != nil {
		logger.Errorf(reputationServiceComponent, "Error al insertar la reseña en la base de datos: %v", err)
		return fmt.Errorf("error interno al guardar la reseña: %w", err)
	}

	logger.Infof(reputationServiceComponent, "Reseña creada exitosamente por %d para %d con %d RP", reviewerID, req.RevieweeID, pointsRP)

	// TODO: Aquí se podría disparar un evento para recalcular el Nivel de Reputación
	// del usuario 'RevieweeId' de forma asíncrona.

	return nil
}

// convertStarsToRP convierte una calificación de 0-5 estrellas a Puntos de Reputación (RP).
func (s *ReputationService) convertStarsToRP(rating float64) int {
	switch {
	case rating >= 5:
		return 100
	case rating >= 4:
		return 60
	case rating >= 3:
		return 25
	case rating >= 2:
		return 10
	case rating >= 1:
		return 1
	default:
		return 0
	}
}
