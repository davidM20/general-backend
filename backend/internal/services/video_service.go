package services

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strconv"
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

// VideoUploadService encapsula la lógica para subir y procesar archivos de video.
type VideoUploadService struct {
	db  *sql.DB
	cfg *config.Config
}

// NewVideoUploadService crea una nueva instancia de VideoUploadService.
func NewVideoUploadService(db *sql.DB, cfg *config.Config) *VideoUploadService {
	return &VideoUploadService{db: db, cfg: cfg}
}

// UploadVideoDetails contiene la información del video subido para la respuesta inicial.
// En una implementación completa, esto podría incluir un ID de trabajo para seguir el progreso de transcodificación.
type UploadVideoDetails struct {
	ID        string `json:"id"`                // ID del contenido (ContentID del video original)
	FileName  string `json:"fileName"`          // Nombre del archivo original en GCS (ej: uuid.mp4)
	Extension string `json:"extension"`         // ej. "mp4", "mov"
	URL       string `json:"url"`               // URL GCS del archivo original
	Size      int64  `json:"size"`              // Tamaño del archivo original en bytes
	Message   string `json:"message,omitempty"` // Mensaje informativo (ej. "Procesamiento iniciado")
}

// MaxVideoSize define el tamaño máximo permitido para archivos de video (ej. 500MB).
const MaxVideoSize = 500 * 1024 * 1024 // 500 MB
const ProcessingStatusUploaded = "uploaded"
const ProcessingStatusProcessing = "processing"
const ProcessingStatusCompleted = "completed"
const ProcessingStatusFailed = "failed"

// ProcessAndUploadVideo procesa un archivo de video subido y lo guarda.
// Por ahora, solo sube el original. La transcodificación sería un paso asíncrono.
func (s *VideoUploadService) ProcessAndUploadVideo(ctx context.Context, userID int64, file multipart.File, fileHeader *multipart.FileHeader) (*UploadVideoDetails, error) {
	if fileHeader.Size > MaxVideoSize {
		logger.Warnf("ProcessAndUploadVideo", "Archivo de video excede el tamaño máximo permitido. Tamaño: %d bytes, Límite: %d bytes", fileHeader.Size, MaxVideoSize)
		return nil, fmt.Errorf("el archivo de video excede el tamaño máximo permitido de %d MB", MaxVideoSize/(1024*1024))
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Errorf("ProcessAndUploadVideo", "Error leyendo el archivo de video: %v", err)
		return nil, fmt.Errorf("error al leer el archivo de video: %w", err)
	}

	kind, _ := filetype.Match(fileBytes)
	if kind == types.Unknown {
		// Si filetype no puede determinarlo, intentamos por extensión (menos fiable)
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		supportedExtensions := map[string]string{
			".mp4": "video/mp4", ".mov": "video/quicktime", ".avi": "video/x-msvideo",
			".mkv": "video/x-matroska", ".webm": "video/webm", ".flv": "video/x-flv",
			".mpeg": "video/mpeg", ".mpg": "video/mpeg",
		}
		if mime, ok := supportedExtensions[ext]; ok {
			kind = types.Type{MIME: types.MIME{Value: mime}, Extension: strings.TrimPrefix(ext, ".")}
		} else {
			return nil, fmt.Errorf("tipo de archivo de video no soportado o inválido: extensión %s", ext)
		}
	}

	allowedMimeValues := map[string]bool{
		"video/mp4": true, "video/quicktime": true, "video/x-msvideo": true,
		"video/x-matroska": true, "video/webm": true, "video/x-flv": true, "video/mpeg": true,
	}
	if !allowedMimeValues[kind.MIME.Value] {
		logger.Warnf("ProcessAndUploadVideo", "El archivo no es un tipo de video permitido. Tipo detectado: %s (%s)", kind.MIME.Value, kind.Extension)
		return nil, fmt.Errorf("el tipo de archivo de video no está permitido: %s", kind.MIME.Value)
	}

	contentID := uuid.New().String()
	baseFileName := uuid.New().String()
	fileExtension := strings.ToLower(kind.Extension)
	switch kind.MIME.Value {
	case "video/quicktime":
		fileExtension = "mov"
	case "video/x-matroska":
		fileExtension = "mkv"
	}
	gcsOriginalFileName := baseFileName + "." + fileExtension

	mpFile := NewInMemoryMultipartFile(fileBytes, gcsOriginalFileName)

	err = cloudclient.UploadFile(ctx, mpFile, gcsOriginalFileName, kind.MIME.Value)
	if err != nil {
		logger.Errorf("ProcessAndUploadVideo", "Error subiendo video original a GCS: %v", err)
		return nil, fmt.Errorf("error subiendo video original a GCS: %w", err)
	}

	gcsOriginalURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.cfg.GCSBucketName, gcsOriginalFileName)

	multimediaRecord := &models.Multimedia{
		Id:               uuid.New().String(),
		Type:             "video",
		Ratio:            0.0, // Se calculará en transcodificación
		UserId:           userID,
		FileName:         gcsOriginalFileName,
		CreateAt:         time.Now(),
		ContentId:        contentID,
		Size:             sql.NullInt64{Int64: fileHeader.Size, Valid: true},
		ProcessingStatus: sql.NullString{String: ProcessingStatusUploaded, Valid: true},
		Duration:         sql.NullFloat64{Valid: false},
	}

	_, dbErr := queries.InsertMultimedia(s.db, multimediaRecord)
	if dbErr != nil {
		// TODO: Considerar borrar de GCS si falla la inserción en BD
		return nil, fmt.Errorf("error guardando registro de video en BD: %w", dbErr)
	}

	logger.Infof("ProcessAndUploadVideo", "Video original subido (ContentID: %s). Disparando transcodificación.", contentID)
	s.startAsyncTranscoding(contentID, gcsOriginalFileName, userID, fileBytes) // Pasar fileBytes para ffprobe simulado

	return &UploadVideoDetails{
		ID:        contentID,
		FileName:  gcsOriginalFileName,
		Extension: fileExtension,
		URL:       gcsOriginalURL,
		Size:      fileHeader.Size,
		Message:   "Video subido. El procesamiento para diferentes calidades ha comenzado.",
	}, nil
}

// startAsyncTranscoding es un placeholder para la lógica que iniciaría la transcodificación.
func (s *VideoUploadService) startAsyncTranscoding(contentID string, originalFilename string, userID int64, originalFileBytes []byte) {
	logger.Infof("startAsyncTranscoding", "Iniciando goroutine para transcodificación de ContentID: %s", contentID)

	go func(db *sql.DB, cfg *config.Config, cID, origFName string, uID int64, fBytes []byte) {
		// Simular un retardo del procesamiento
		time.Sleep(5 * time.Second) // Simulación corta, en realidad puede tardar minutos/horas

		err := queries.UpdateMultimediaProcessingStatus(db, cID, ProcessingStatusProcessing)
		if err != nil {
			logger.Errorf("startAsyncTranscoding.UpdateStatus", "Error actualizando estado a '%s' para ContentID %s: %v", ProcessingStatusProcessing, cID, err)
			return // Salir si no podemos actualizar el estado
		}

		// Simular extracción de metadatos (ffprobe)
		// En un caso real, escribirías fBytes a un archivo temporal y ejecutarías ffprobe.
		dummyWidth, dummyHeight, dummyDuration, err := mockFFProbe(origFName)
		if err != nil {
			logger.Errorf("startAsyncTranscoding.mockFFProbe", "Error al extraer metadatos para ContentID %s: %v", cID, err)
			err = queries.UpdateMultimediaProcessingStatus(db, cID, ProcessingStatusFailed)
			if err != nil {
				logger.Errorf("startAsyncTranscoding.UpdateStatusFailed", "Error actualizando estado a '%s' para ContentID %s: %v", ProcessingStatusFailed, cID, err)
			}
			return
		}

		aspectRatio := 0.0
		if dummyHeight > 0 {
			aspectRatio = float64(dummyWidth) / float64(dummyHeight)
		}
		logger.Infof("transcodeAndSegmentVideo", "[SIMULADO] ContentID: %s - Metadatos: WxH: %dx%d, Duración: %.2fs, Ratio: %.2f", cID, dummyWidth, dummyHeight, dummyDuration, aspectRatio)

		resolutions := map[string]string{"1080p": "1920x1080", "720p": "1280x720", "480p": "854x480"}
		bitrates := map[string]string{"1080p": "5000k", "720p": "2500k", "480p": "1000k"}
		baseGCSManifestPath := fmt.Sprintf("videos/%s", cID) // Ej: videos/content-uuid/

		var p1080, p720, p480 string

		// Simulación de transcodificación para cada resolución
		for resName, resValue := range resolutions {
			// Verificar si la resolución del video es suficiente para esta calidad
			resParts := strings.Split(resValue, "x")
			targetWidth, _ := strconv.Atoi(resParts[0])
			// targetHeight, _ := strconv.Atoi(resParts[1]) // No se usa directamente aquí

			if dummyWidth < targetWidth && resName != "480p" { // Siempre generar 480p si es posible, o la más baja
				logger.Infof("transcodeAndSegmentVideo", "[SIMULADO] Omitiendo %s para ContentID %s, resolución original (%d) menor que objetivo (%d)", resName, cID, dummyWidth, targetWidth)
				continue
			}

			outputPath := filepath.Join(baseGCSManifestPath, resName)
			manifestName := "playlist.m3u8"
			segmentPrefix := "segment"
			segmentPath := filepath.Join(outputPath, segmentPrefix+"%03d.ts")
			manifestGCSPath := filepath.Join(outputPath, manifestName)

			ffmpegCmd := fmt.Sprintf("ffmpeg -i %s -vf scale=%s -c:v libx264 -preset medium -b:v %s -c:a aac -hls_time 10 -hls_playlist_type vod -hls_segment_filename %s %s",
				origFName, resValue, bitrates[resName], segmentPath, manifestGCSPath)
			logger.Infof("transcodeAndSegmentVideo", "[SIMULADO FFmpeg CMD para %s]: %s", resName, ffmpegCmd)

			// Simular ejecución de FFmpeg y subida a GCS
			logger.Infof("transcodeAndSegmentVideo", "[SIMULADO] Subiendo HLS para %s a GCS path: %s", resName, outputPath)

			switch resName {
			case "1080p":
				p1080 = manifestGCSPath
			case "720p":
				p720 = manifestGCSPath
			case "480p":
				p480 = manifestGCSPath
			}
			// Simular un pequeño retardo por cada resolución
			time.Sleep(2 * time.Second)
		}

		// Si no se generó ninguna variante (ej. video muy pequeño), marcar como fallido o manejar de otra forma.
		if p1080 == "" && p720 == "" && p480 == "" {
			logger.Warnf("transcodeAndSegmentVideo", "[SIMULADO] No se generaron variantes HLS para ContentID %s. Marcando como fallido.", cID)
			err = queries.UpdateMultimediaVariants(db, cID, aspectRatio, dummyDuration, "", "", "", "", ProcessingStatusFailed)
		} else {
			logger.Infof("transcodeAndSegmentVideo", "[SIMULADO] Transcodificación completada para ContentID %s. Actualizando BD.", cID)
			err = queries.UpdateMultimediaVariants(db, cID, aspectRatio, dummyDuration, baseGCSManifestPath, p1080, p720, p480, ProcessingStatusCompleted)
		}

		if err != nil {
			logger.Errorf("startAsyncTranscoding.UpdateVariants", "Error actualizando variantes en BD para ContentID %s: %v", cID, err)
			// Podríamos intentar marcar como fallido si la actualización de variantes falla después de un procesamiento exitoso (parcial)
			_ = queries.UpdateMultimediaProcessingStatus(db, cID, ProcessingStatusFailed) // Intentar marcar como fallido
		}
		logger.Infof("startAsyncTranscoding", "Goroutine de transcodificación finalizada para ContentID: %s", cID)
	}(s.db, s.cfg, contentID, originalFilename, userID, originalFileBytes) // Pasar s.db y s.cfg a la goroutine
}

// El placeholder transcodeAndSegmentVideo se ha integrado en la goroutine de startAsyncTranscoding
// para esta simulación. En una implementación real con workers, estaría en el código del worker.

// Mock ffprobe para simular la extracción de metadatos (solo para esta simulación)
// En un caso real, esto llamaría al ejecutable ffprobe.
func mockFFProbe(filePath string) (width, height int, duration float64, err error) {
	// Simulación: Estos valores serían extraídos por ffprobe
	// Para probar, podrías basar esto en el nombre del archivo o alguna heurística simple.
	if strings.Contains(filePath, "error") {
		return 0, 0, 0, fmt.Errorf("simulated ffprobe error")
	}
	// Valores por defecto simulados
	return 1920, 1080, 125.5, nil
}
