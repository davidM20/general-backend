package queries

import (
	"database/sql"
	"fmt"
	"time" // Necesario para UpdateMultimediaVariants si se actualiza CreateAt o similar

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

/*
Package queries proporciona un lugar centralizado para toda la lógica de acceso a la base de datos.
Este archivo contiene funciones para interactuar con las tablas de la base de datos.

NORMAS Y DIRECTRICES PARA ESTE ARCHIVO:

1. CONEXIÓN A LA BASE DE DATOS:
  - La variable global `DB *sql.DB` se inicializa en el arranque.
  - NO pasar el puntero de conexión a la base de datos como argumento a las funciones.
  - Todas las funciones de consulta dentro de este paquete deben usar la variable global `DB` directamente.

2. REUTILIZACIÓN Y RESPONSABILIDAD DEL CÓDIGO:
  - Antes de añadir una nueva función, comprueba si una existente puede ser reutilizada o generalizada.
  - Cada función debe tener una única responsabilidad, claramente definida (p. ej., obtener datos de usuario, insertar un mensaje).
  - Mantén las funciones concisas y enfocadas.

3. DOCUMENTACIÓN:
  - Documenta cada nueva función y tipo.
  - Los comentarios deben explicar el propósito de la función, sus parámetros y lo que devuelve.
  - Explica cualquier lógica compleja o comportamiento no obvio.

4. MANEJO DE ERRORES:
  - Comprueba siempre los errores devueltos por `DB.Query`, `DB.QueryRow`, `DB.Exec` y `rows.Scan`.
  - Utiliza `fmt.Errorf("contexto: %w", err)` para envolver los errores, proporcionando contexto sin perder el error original.
  - Maneja `sql.ErrNoRows` específicamente cuando se espera que una consulta a veces no devuelva resultados (p. ej., `GetUserBy...`).

5. CONVENCIONES DE NOMENCLATURA:
  - Sigue las convenciones de nomenclatura idiomáticas de Go (p. ej., `CamelCase` para identificadores exportados).
  - Usa nombres descriptivos para las funciones (p. ej., `GetUserBySessionToken`, `CreateMessage`).

6. CONSTANTES:
  - Para campos de estado o IDs de tipo (p. ej., estado del mensaje), define constantes en la parte superior del archivo.
  - Usa estas constantes en lugar de números mágicos para mejorar la legibilidad y el mantenimiento.

7. MANEJO DE COLUMNAS ANULABLES:
  - Usa `sql.NullString`, `sql.NullInt64`, `sql.NullTime`, etc., para columnas de la base de datos que pueden ser NULL.
  - Comprueba siempre el campo `Valid` antes de acceder al valor de un tipo anulable.

8. SEGURIDAD:
  - Para prevenir la inyección de SQL, SIEMPRE usa consultas parametrizadas con `?` como marcadores de posición.
  - NUNCA construyas consultas concatenando cadenas con entradas proporcionadas por el usuario.

9. AÑADIR NUEVAS CONSULTAS:
  - Agrupa las funciones relacionadas (p. ej., todas las consultas relacionadas con el usuario, todas las relacionadas con los mensajes).
  - Considera las implicaciones de rendimiento. Usa `JOIN`s con criterio y añade cláusulas `LIMIT` donde sea aplicable.
  - Asegúrate de que tu consulta devuelva solo las columnas necesarias.
*/

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
		INSERT INTO Multimedia (
			Id, Type, Ratio, UserId, FileName, CreateAt, ContentId, ChatId, Size, 
			ProcessingStatus, Duration, HLSManifestBaseURL, 
			HLSManifest1080p, HLSManifest720p, HLSManifest480p
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	stmt, err := db.Prepare(query)
	if err != nil {
		logger.Errorf("InsertMultimedia.Prepare", "Error preparando la consulta: %v", err)
		return "", fmt.Errorf("error preparando la consulta para insertar multimedia: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		m.Id, m.Type, m.Ratio, m.UserId, m.FileName, m.CreateAt, m.ContentId, m.ChatId, m.Size,
		m.ProcessingStatus, m.Duration, m.HLSManifestBaseURL,
		m.HLSManifest1080p, m.HLSManifest720p, m.HLSManifest480p,
	)
	if err != nil {
		logger.Errorf("InsertMultimedia.Exec", "Error al ejecutar la inserción de multimedia: %v", err)
		return "", fmt.Errorf("error al insertar registro de multimedia: %w", err)
	}

	logger.Infof("InsertMultimedia", "Registro multimedia insertado con ID: %s, ContentID: %s, FileName: %s", m.Id, m.ContentId, m.FileName)
	return m.Id, nil
}

// UpdateMultimediaProcessingStatus actualiza el estado de procesamiento de un video.
func UpdateMultimediaProcessingStatus(db *sql.DB, contentID string, status string) error {
	query := `UPDATE Multimedia SET ProcessingStatus = ? WHERE ContentId = ? AND Type = 'video';`
	stmt, err := db.Prepare(query)
	if err != nil {
		logger.Errorf("UpdateMultimediaProcessingStatus.Prepare", "Error preparando la consulta para actualizar estado: %v", err)
		return fmt.Errorf("error preparando la consulta para actualizar estado de video: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(status, contentID)
	if err != nil {
		logger.Errorf("UpdateMultimediaProcessingStatus.Exec", "Error actualizando estado de video para ContentID %s: %v", contentID, err)
		return fmt.Errorf("error actualizando estado de video: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Warnf("UpdateMultimediaProcessingStatus", "No se encontró ningún video con ContentID %s para actualizar estado.", contentID)
		return fmt.Errorf("no se encontró video con ContentID %s", contentID) // O podría no ser un error si es esperado
	}

	logger.Infof("UpdateMultimediaProcessingStatus", "Estado de video actualizado para ContentID %s a %s", contentID, status)
	return nil
}

// UpdateMultimediaVariants actualiza los detalles de las variantes de video procesadas.
func UpdateMultimediaVariants(db *sql.DB, contentID string, ratio float64, duration float64, baseURL, p1080, p720, p480, status string) error {
	query := `
		UPDATE Multimedia SET 
			Ratio = ?,
			Duration = ?,
			HLSManifestBaseURL = ?,
			HLSManifest1080p = ?,
			HLSManifest720p = ?,
			HLSManifest480p = ?,
			ProcessingStatus = ?,
			CreateAt = ?  -- Actualizar CreateAt para reflejar la última modificación del procesamiento
		WHERE ContentId = ? AND Type = 'video';
	`
	// Nota: Se actualiza CreateAt. Si se quiere mantener el tiempo de subida original,
	// se podría añadir un campo separate `ProcessedAt` o `UpdatedAt`.

	stmt, err := db.Prepare(query)
	if err != nil {
		logger.Errorf("UpdateMultimediaVariants.Prepare", "Error preparando la consulta para actualizar variantes: %v", err)
		return fmt.Errorf("error preparando la consulta para actualizar variantes de video: %w", err)
	}
	defer stmt.Close()

	// Usar sql.Null types si los parámetros pueden ser nulos
	nullBaseURL := sql.NullString{String: baseURL, Valid: baseURL != ""}
	nullP1080 := sql.NullString{String: p1080, Valid: p1080 != ""}
	nullP720 := sql.NullString{String: p720, Valid: p720 != ""}
	nullP480 := sql.NullString{String: p480, Valid: p480 != ""}

	result, err := stmt.Exec(
		ratio, duration, nullBaseURL, nullP1080, nullP720, nullP480, status, time.Now(), contentID,
	)
	if err != nil {
		logger.Errorf("UpdateMultimediaVariants.Exec", "Error actualizando variantes de video para ContentID %s: %v", contentID, err)
		return fmt.Errorf("error actualizando variantes de video: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Warnf("UpdateMultimediaVariants", "No se encontró ningún video con ContentID %s para actualizar variantes.", contentID)
		return fmt.Errorf("no se encontró video con ContentID %s para actualizar variantes", contentID)
	}

	logger.Infof("UpdateMultimediaVariants", "Variantes de video actualizadas para ContentID %s. Estado: %s", contentID, status)
	return nil
}

// GetMultimediaByContentID recupera un registro multimedia por su ContentId,
// específicamente para videos, incluyendo campos relevantes para HLS.
func GetMultimediaByContentID(db *sql.DB, contentID string) (*models.Multimedia, error) {
	query := `
		SELECT 
			Id, Type, Ratio, UserId, FileName, CreateAt, ContentId, ChatId, Size, 
			ProcessingStatus, Duration, HLSManifestBaseURL, 
			HLSManifest1080p, HLSManifest720p, HLSManifest480p
		FROM Multimedia 
		WHERE ContentId = ? AND Type = 'video';
	`
	stmt, err := db.Prepare(query)
	if err != nil {
		logger.Errorf("GetMultimediaByContentID.Prepare", "Error preparando la consulta: %v", err)
		return nil, fmt.Errorf("error preparando la consulta para obtener multimedia por ContentID: %w", err)
	}
	defer stmt.Close()

	m := &models.Multimedia{}
	err = stmt.QueryRow(contentID).Scan(
		&m.Id, &m.Type, &m.Ratio, &m.UserId, &m.FileName, &m.CreateAt, &m.ContentId, &m.ChatId, &m.Size,
		&m.ProcessingStatus, &m.Duration, &m.HLSManifestBaseURL,
		&m.HLSManifest1080p, &m.HLSManifest720p, &m.HLSManifest480p,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warnf("GetMultimediaByContentID.Scan", "No se encontró multimedia con ContentID %s", contentID)
			return nil, fmt.Errorf("multimedia no encontrada: %w", err) // Devolver err para que el servicio lo maneje como 404
		}
		logger.Errorf("GetMultimediaByContentID.Scan", "Error escaneando multimedia para ContentID %s: %v", contentID, err)
		return nil, fmt.Errorf("error escaneando multimedia: %w", err)
	}

	logger.Infof("GetMultimediaByContentID", "Multimedia recuperada para ContentID: %s", contentID)
	return m, nil
}
