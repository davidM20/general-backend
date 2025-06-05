package queries

import (
	"database/sql"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

/*
 * ===================================================
 * CONSULTAS SQL PARA MULTIMEDIA
 * ===================================================
 *
 * Este archivo contiene las consultas SQL para gestionar registros
 * en la tabla Multimedia.
 */

// InsertMultimedia inserta un nuevo registro en la tabla Multimedia.
// Devuelve el ID del registro insertado o un error.
// TODO: Integrar el wrapper de métricas correctamente según el proyecto.
func InsertMultimedia(db *sql.DB, m *models.Multimedia) (string, error) {
	query := `
		INSERT INTO Multimedia (Id, Type, Ratio, UserId, FileName, CreateAt, ContentId, ChatId)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`
	stmt, err := db.Prepare(query)
	if err != nil {
		logger.Errorf("InsertMultimedia.Prepare", "Error preparando la consulta: %v", err)
		return "", fmt.Errorf("error preparando la consulta para insertar multimedia: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(m.Id, m.Type, m.Ratio, m.UserId, m.FileName, m.CreateAt, m.ContentId, m.ChatId)
	if err != nil {
		logger.Errorf("InsertMultimedia.Exec", "Error al ejecutar la inserción de multimedia: %v", err)
		return "", fmt.Errorf("error al insertar registro de multimedia: %w", err)
	}

	logger.Infof("InsertMultimedia", "Registro multimedia insertado con ID: %s, FileName: %s", m.Id, m.FileName)
	return m.Id, nil
}
