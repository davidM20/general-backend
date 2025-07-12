package queries

import (
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
)

// GetReputationStatsByUserID recupera las estadísticas de reputación para un usuario específico.
// Devuelve el número total de reseñas y la suma total de Puntos de Reputación (RP).
func GetReputationStatsByUserID(userID int64) (*models.ReputationStats, error) {
	query := `
        SELECT
            COUNT(*),
            COALESCE(SUM(PointsRP), 0)
        FROM ReputationReview
        WHERE RevieweeId = ?
    `
	var stats models.ReputationStats

	err := DB.QueryRow(query, userID).Scan(
		&stats.ReviewCount,
		&stats.TotalPointsRP,
	)

	if err != nil {
		return nil, fmt.Errorf("error al obtener estadísticas de reputación para el usuario %d: %w", userID, err)
	}

	return &stats, nil
}

// GetReputationReviewsByUserID recupera una lista de reseñas detalladas para un usuario.
// Solo incluye reseñas hechas por empresas (RoleId = 3).
func GetReputationReviewsByUserID(userID int64) ([]models.ReputationReviewInfo, error) {
	query := `
        SELECT
	    rr.Id,
            rr.Rating,
            rr.Comment,
            reviewer.CompanyName,
            reviewer.Picture
        FROM ReputationReview rr
        JOIN User reviewer ON rr.ReviewerId = reviewer.Id
        WHERE rr.RevieweeId = ? AND reviewer.RoleId = 3
        ORDER BY rr.CreatedAt DESC
    `

	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar reseñas para el usuario %d: %w", userID, err)
	}
	defer rows.Close()

	var reviews []models.ReputationReviewInfo
	for rows.Next() {
		var review models.ReputationReviewInfo
		if err := rows.Scan(
			&review.Id,
			&review.Rating,
			&review.Comment,
			&review.ReviewerCompanyName,
			&review.ReviewerPicture,
		); err != nil {
			return nil, fmt.Errorf("error al escanear reseña: %w", err)
		}
		reviews = append(reviews, review)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error al iterar sobre las filas de reseñas: %w", err)
	}

	return reviews, nil
}
