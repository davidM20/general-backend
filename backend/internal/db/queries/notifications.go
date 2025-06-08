package queries

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// CreateNotification inserta una nueva notificación (evento) en la base de datos.
// Utiliza un struct para un paso de parámetros claro y extensible.
func CreateNotification(notification models.Event) (int64, error) {
	var metadataJSON []byte
	var err error
	if notification.Metadata != nil {
		// Solo serializa si Metadata no es nulo
		metadataJSON, err = json.Marshal(notification.Metadata)
		if err != nil {
			logger.Errorf("QUERY", "Error al serializar metadatos de notificación: %v", err)
			return 0, fmt.Errorf("error al serializar metadatos: %w", err)
		}
	}

	query := `
        INSERT INTO Event (
            EventType, EventTitle, Description, UserId, OtherUserId, 
            ProyectId, CreateAt, IsRead, GroupId, Status, 
            ActionRequired, ActionTakenAt, Metadata
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Usar el tiempo actual para CreateAt, y false para IsRead y PENDING para Status
	// ActionTakenAt es nulo a menos que se especifique una acción ya tomada
	result, err := DB.Exec(query,
		notification.EventType,
		notification.EventTitle,
		notification.Description,
		notification.UserId,
		notification.OtherUserId,
		notification.ProyectId,
		time.Now().UTC(), // Usar UTC para consistencia
		false,
		notification.GroupId,
		"PENDING",
		notification.ActionRequired,
		notification.ActionTakenAt,
		metadataJSON, // Puede ser nil si no hay metadatos
	)

	if err != nil {
		logger.Errorf("QUERY", "Error al crear notificación para el usuario %d: %v", notification.UserId, err)
		return 0, fmt.Errorf("no se pudo crear la notificación: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Errorf("QUERY", "Error al obtener el ID de la última inserción para la notificación: %v", err)
		return 0, fmt.Errorf("no se pudo recuperar el ID de la notificación: %w", err)
	}

	logger.Successf("QUERY", "Notificación creada con éxito con ID %d para el usuario %d", id, notification.UserId)
	return id, nil
}
