package queries

const (
	// CreateJobApplication inserta una nueva postulación en la base de datos.
	CreateJobApplication = `
		INSERT INTO JobApplication (CommunityEventId, ApplicantId, CoverLetter, Status, AppliedAt, UpdatedAt)
		VALUES (?, ?, ?, 'ENVIADA', NOW(), NOW())
	`

	// ListApplicantsByEvent recupera la lista de postulantes para una oferta específica,
	// ordenados por su calificación y reputación.
	ListApplicantsByEvent = `
		WITH UserRatings AS (
			SELECT
				RevieweeId,
				AVG(Rating) AS AverageRating,
				COUNT(Id) AS TotalReviews
			FROM
				ReputationReview
			WHERE
				Rating IS NOT NULL
			GROUP BY
				RevieweeId
		),
		UserReputationScore AS (
			SELECT
				RevieweeId,
				SUM(PointsRP) AS TotalReputation
			FROM
				ReputationReview
			GROUP BY
				RevieweeId
		)
		SELECT
			ja.ApplicantId,
			u.FirstName,
			u.LastName,
			u.Email,
			COALESCE(ur.AverageRating, 0) AS AverageRating,
			COALESCE(urs.TotalReputation, 0) AS ReputationScore,
			ja.Status AS ApplicationStatus,
			ja.AppliedAt
		FROM
			JobApplication ja
		JOIN
			User u ON ja.ApplicantId = u.Id
		LEFT JOIN
			UserRatings ur ON ja.ApplicantId = ur.RevieweeId
		LEFT JOIN
			UserReputationScore urs ON ja.ApplicantId = urs.RevieweeId
		WHERE
			ja.CommunityEventId = ?
		ORDER BY
			AverageRating DESC,
			ReputationScore DESC;
	`

	// UpdateJobApplicationStatus actualiza el estado de una postulación específica.
	UpdateJobApplicationStatus = `
		UPDATE JobApplication
		SET Status = ?
		WHERE CommunityEventId = ? AND ApplicantId = ?
	`
	// TODO: Añadir más queries según se necesiten, como:
	// - GetJobApplicationByID: Para obtener los detalles de una postulación específica.
)
