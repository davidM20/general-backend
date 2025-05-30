package queries

import (
	"database/sql"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
)

// MetricsRecorder define la interfaz para registrar métricas
type MetricsRecorder interface {
	RecordDatabaseQuery(duration time.Duration)
}

var metricsRecorder MetricsRecorder

// SetMetricsRecorder establece el recorder de métricas
func SetMetricsRecorder(recorder MetricsRecorder) {
	metricsRecorder = recorder
}

// MeasureQuery es un decorador que mide el tiempo de ejecución de una consulta
func MeasureQuery(queryFunc func() error) error {
	start := time.Now()
	err := queryFunc()
	duration := time.Since(start)

	if metricsRecorder != nil {
		metricsRecorder.RecordDatabaseQuery(duration)
	}

	return err
}

// MeasureQueryWithResult es un decorador para consultas que retornan un valor
func MeasureQueryWithResult[T any](queryFunc func() (T, error)) (T, error) {
	start := time.Now()
	result, err := queryFunc()
	duration := time.Since(start)

	if metricsRecorder != nil {
		metricsRecorder.RecordDatabaseQuery(duration)
	}

	return result, err
}

// Ejemplos de uso en funciones existentes:

// GetUserBySessionTokenWithMetrics es un ejemplo de cómo envolver una consulta existente
func GetUserBySessionTokenWithMetrics(db *sql.DB, token string) (*models.User, error) {
	return MeasureQueryWithResult(func() (*models.User, error) {
		return GetUserBySessionToken(db, token)
	})
}

// CreateMessageFromChatParamsWithMetrics crea un mensaje usando parámetros de chat con métricas
func CreateMessageFromChatParamsWithMetrics(db *sql.DB, fromUserID, toUserID int64, content string) (*models.Message, error) {
	// Aquí podrías agregar métricas como incrementar contadores, etc.
	return CreateMessageFromChatParams(db, fromUserID, toUserID, content)
}
