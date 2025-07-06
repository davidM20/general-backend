package services

import (
	"database/sql"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// NotificationService maneja la lógica de negocio para las notificaciones (Eventos).
type NotificationService struct {
	DB *sql.DB
}

// NewNotificationService crea una nueva instancia de NotificationService.
func NewNotificationService(db *sql.DB) *NotificationService {
	return &NotificationService{DB: db}
}

// MarkAsRead marca una notificación como leída para un usuario específico.
// Solo el propietario de la notificación puede marcarla como leída.
func (s *NotificationService) MarkAsRead(notificationID int64, userID int64) error {
	query := `UPDATE Event SET IsRead = true WHERE Id = ? AND UserId = ?`

	stmt, err := s.DB.Prepare(query)
	if err != nil {
		logger.Errorf("NOTIFICATION_SERVICE", "Error preparing statement for MarkAsRead: %v", err)
		return fmt.Errorf("error interno del servidor al preparar la consulta")
	}
	defer stmt.Close()

	result, err := stmt.Exec(notificationID, userID)
	if err != nil {
		logger.Errorf("NOTIFICATION_SERVICE", "Error executing MarkAsRead for notification %d by user %d: %v", notificationID, userID, err)
		return fmt.Errorf("error interno del servidor al actualizar la notificación")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Errorf("NOTIFICATION_SERVICE", "Error getting rows affected for MarkAsRead (notification %d, user %d): %v", notificationID, userID, err)
		return fmt.Errorf("error interno del servidor al verificar la actualización")
	}

	if rowsAffected == 0 {
		logger.Warnf("NOTIFICATION_SERVICE", "MarkAsRead attempt on notification %d by user %d resulted in 0 rows affected. Either notification doesn't exist or user is not the owner.", notificationID, userID)
		// Devolvemos un error específico para que el handler pueda devolver un 404 Not Found.
		return fmt.Errorf("notificación no encontrada o no tienes permiso para modificarla")
	}

	logger.Successf("NOTIFICATION_SERVICE", "Notification %d successfully marked as read for user %d", notificationID, userID)
	return nil
}
