package services

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/cloudclient"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

// AudioUploadService encapsula la lógica para subir y procesar archivos de audio.
type AudioUploadService struct {
	db  *sql.DB
	cfg *config.Config
}

// NewAudioUploadService crea una nueva instancia de AudioUploadService.
func NewAudioUploadService(db *sql.DB, cfg *config.Config) *AudioUploadService {
	return &AudioUploadService{db: db, cfg: cfg}
}

// UploadAudioDetails contiene la información del audio subido para la respuesta.
type UploadAudioDetails struct {
	ID        string `json:"id"`        // ID del contenido (ContentID)
	FileName  string `json:"fileName"`  // Nombre del archivo en GCS (ej: uuid.mp3)
	Extension string `json:"extension"` // ej. "mp3", "wav"
	URL       string `json:"url"`       // URL GCS del archivo
	// Podríamos añadir Duration float32 `json:"duration,omitempty"` si se implementa extracción de duración
}

// ProcessAndUploadAudio procesa un archivo de audio subido y lo guarda.
func (s *AudioUploadService) ProcessAndUploadAudio(ctx context.Context, userID int64, file multipart.File, fileHeader *multipart.FileHeader) (*UploadAudioDetails, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Errorf("ProcessAndUploadAudio", "Error leyendo el archivo de audio: %v", err)
		return nil, fmt.Errorf("error al leer el archivo de audio: %w", err)
	}

	kind, err := filetype.Match(fileBytes)
	if err != nil || kind == types.Unknown {
		logger.Warnf("ProcessAndUploadAudio", "Tipo de archivo de audio desconocido o no soportado. Error: %v, Kind: %s", err, kind.Extension)
		return nil, fmt.Errorf("tipo de archivo de audio no soportado o inválido")
	}

	// Validar que es un tipo de audio soportado
	// Referencia: https://github.com/h2non/filetype/blob/master/types/audio.go
	// Podemos ser más explícitos con los tipos que aceptamos.
	allowedMimeTypes := map[string]bool{
		"audio/mpeg":  true, // MP3
		"audio/wav":   true, // WAV
		"audio/x-wav": true, // WAV alternativo
		"audio/aac":   true, // AAC
		"audio/ogg":   true, // OGG Vorbis
		"audio/opus":  true, // Opus
		"audio/mp4":   true, // M4A (audio MP4)
		"audio/flac":  true, // FLAC
		"audio/webm":  true, // WebM audio (a menudo Opus o Vorbis)
	}

	if !allowedMimeTypes[kind.MIME.Value] {
		logger.Warnf("ProcessAndUploadAudio", "El archivo no es un tipo de audio permitido. Tipo detectado: %s (%s)", kind.MIME.Value, kind.Extension)
		return nil, fmt.Errorf("el tipo de archivo de audio no está permitido: %s", kind.MIME.Value)
	}

	contentID := uuid.New().String()
	baseFileName := uuid.New().String()
	fileExtension := strings.ToLower(kind.Extension) // Usar la extensión detectada

	// Para algunos tipos, la extensión puede no ser la más común (ej. mpeg para mp3)
	switch kind.MIME.Value {
	case "audio/mpeg":
		fileExtension = "mp3"
	case "audio/mp4":
		fileExtension = "m4a"
	case "audio/x-wav":
		fileExtension = "wav"
	}

	gcsFileName := baseFileName + "." + fileExtension

	// Subir el archivo original tal cual (sin conversión por ahora)
	// Usamos InMemoryMultipartFile porque cloudclient.UploadFile espera multipart.File
	// y ya tenemos fileBytes. Si cloudclient.UploadFile pudiera tomar io.Reader directamente,
	// podríamos usar bytes.NewReader(fileBytes).
	mpFile := NewInMemoryMultipartFile(fileBytes, gcsFileName) // Asumimos que InMemoryMultipartFile está en este paquete o importado

	err = cloudclient.UploadFile(ctx, mpFile, gcsFileName, kind.MIME.Value)
	if err != nil {
		logger.Errorf("ProcessAndUploadAudio", "Error subiendo audio a GCS: %v", err)
		return nil, fmt.Errorf("error subiendo audio a GCS: %w", err)
	}

	gcsURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.cfg.GCSBucketName, gcsFileName)

	_, dbErr := queries.InsertMultimedia(s.db, &models.Multimedia{
		Id:        uuid.New().String(), // ID único para esta entrada de BD
		Type:      "audio",
		Ratio:     0.0, // Duración no calculada en esta versión
		UserId:    userID,
		FileName:  gcsFileName,
		CreateAt:  time.Now(),
		ContentId: contentID, // ID para agrupar (aunque aquí solo hay un archivo por subida)
	})
	if dbErr != nil {
		logger.Errorf("ProcessAndUploadAudio", "Error guardando registro de audio en BD: %v", dbErr)
		// TODO: Considerar borrar el archivo de GCS si la inserción en BD falla (compensación)
		return nil, fmt.Errorf("error guardando registro de audio en BD: %w", dbErr)
	}

	logger.Infof("ProcessAndUploadAudio", "Audio subido y registrado: UserID %d, FileName %s, GCS_URL %s", userID, gcsFileName, gcsURL)

	return &UploadAudioDetails{
		ID:        contentID,
		FileName:  gcsFileName,
		Extension: fileExtension,
		URL:       gcsURL,
	}, nil
}
