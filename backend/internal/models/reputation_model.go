package models

// CreateReviewRequest define la estructura para la solicitud de creación de una reseña.
// Contiene todos los datos necesarios para registrar una calificación en el sistema.
type CreateReviewRequest struct {
	// ID del usuario que está siendo calificado.
	RevieweeID int64 `json:"revieweeId"`

	// Calificación en formato de estrellas, de 0 a 5.
	// Se permite un decimal para calificaciones como 4.5.
	Rating float64 `json:"rating"`

	// Comentario o feedback cualitativo sobre la interacción.
	Comment string `json:"comment"`

	// Tipo de interacción que originó la reseña (ej. "ENTREVISTA").
	// Debe coincidir con los valores del ENUM en la base de datos.
	InteractionType string `json:"interactionType"`

	// Campo booleano para indicar si se debe aplicar el bono especial.
	// Corresponde a la condición de "3 estrellas extra".
	ApplyBonus bool `json:"applyBonus"`
}
