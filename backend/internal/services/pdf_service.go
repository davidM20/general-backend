package services

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/cloudclient"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers" // Para application/pdf
	"github.com/h2non/filetype/types"
)

// PDFUploadService encapsula la lógica para subir y procesar archivos PDF.
type PDFUploadService struct {
	db  *sql.DB
	cfg *config.Config
}

// NewPDFUploadService crea una nueva instancia de PDFUploadService.
func NewPDFUploadService(db *sql.DB, cfg *config.Config) *PDFUploadService {
	return &PDFUploadService{db: db, cfg: cfg}
}

// UploadPDFDetails contiene la información del PDF subido para la respuesta.
type UploadPDFDetails struct {
	ID        string `json:"id"`        // ID del contenido (ContentID)
	FileName  string `json:"fileName"`  // Nombre del archivo en GCS (ej: uuid.pdf)
	Extension string `json:"extension"` // Siempre "pdf"
	URL       string `json:"url"`       // URL GCS del archivo
	Size      int64  `json:"size"`      // Tamaño del archivo en bytes
}

// MaxPDFSize define el tamaño máximo permitido para archivos PDF (ej. 10MB).
const MaxPDFSize = 10 * 1024 * 1024 // 10 MB

// ProcessAndUploadPDF procesa un archivo PDF subido y lo guarda.
// Incluye validaciones de tipo MIME y tamaño.
func (s *PDFUploadService) ProcessAndUploadPDF(ctx context.Context, userID int64, file multipart.File, fileHeader *multipart.FileHeader) (*UploadPDFDetails, error) {
	// Validar tamaño del archivo antes de leerlo completamente en memoria
	if fileHeader.Size > MaxPDFSize {
		logger.Warnf("ProcessAndUploadPDF", "Archivo PDF excede el tamaño máximo permitido. Tamaño: %d bytes, Límite: %d bytes", fileHeader.Size, MaxPDFSize)
		return nil, fmt.Errorf("el archivo PDF excede el tamaño máximo permitido de %d MB", MaxPDFSize/(1024*1024))
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Errorf("ProcessAndUploadPDF", "Error leyendo el archivo PDF: %v", err)
		return nil, fmt.Errorf("error al leer el archivo PDF: %w", err)
	}

	kind, err := filetype.Match(fileBytes)
	if err != nil {
		logger.Errorf("ProcessAndUploadPDF", "Error determinando el tipo de archivo: %v", err)
		return nil, fmt.Errorf("error al determinar el tipo de archivo: %w", err)
	}

	if kind == types.Unknown || kind.MIME.Value != matchers.TypePdf.MIME.Value {
		logger.Warnf("ProcessAndUploadPDF", "Tipo de archivo no es PDF. Detectado: %s (%s)", kind.MIME.Value, kind.Extension)
		return nil, fmt.Errorf("el archivo no es un PDF válido. Tipo detectado: %s", kind.MIME.Value)
	}

	// Aquí se podrían añadir validaciones más profundas del contenido del PDF si se dispone
	// de librerías especializadas para análisis de PDF (ej. buscar scripts maliciosos,
	// verificar estructura, etc.). Por ahora, nos centramos en tipo y tamaño.
	logger.Infof("ProcessAndUploadPDF", "Archivo validado como PDF. MIME: %s, Extension: %s", kind.MIME.Value, kind.Extension)

	contentID := uuid.New().String()
	baseFileName := uuid.New().String()
	fileExtension := "pdf" // Para PDFs, la extensión es fija.

	gcsFileName := baseFileName + "." + fileExtension

	// Asumimos que NewInMemoryMultipartFile está disponible en el paquete 'services'
	// (definido en audio_service.go o image_service.go)
	mpFile := NewInMemoryMultipartFile(fileBytes, gcsFileName)

	err = cloudclient.UploadFile(ctx, mpFile, gcsFileName, kind.MIME.Value)
	if err != nil {
		logger.Errorf("ProcessAndUploadPDF", "Error subiendo PDF a GCS: %v", err)
		return nil, fmt.Errorf("error subiendo PDF a GCS: %w", err)
	}

	gcsURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.cfg.GCSBucketName, gcsFileName)

	_, dbErr := queries.InsertMultimedia(s.db, &models.Multimedia{
		Id:        uuid.New().String(), // ID único para esta entrada de BD
		Type:      "pdf",
		Ratio:     0.0, // No aplica para PDF en este contexto
		UserId:    userID,
		FileName:  gcsFileName,
		CreateAt:  time.Now(),
		ContentId: contentID,
		Size:      sql.NullInt64{Int64: fileHeader.Size, Valid: true}, // Guardar tamaño del archivo
	})
	if dbErr != nil {
		logger.Errorf("ProcessAndUploadPDF", "Error guardando registro de PDF en BD: %v", dbErr)
		// Considerar borrar el archivo de GCS si la inserción en BD falla
		return nil, fmt.Errorf("error guardando registro de PDF en BD: %w", dbErr)
	}

	logger.Infof("ProcessAndUploadPDF", "PDF subido y registrado: UserID %d, FileName %s, GCS_URL %s, Size: %d", userID, gcsFileName, gcsURL, fileHeader.Size)

	return &UploadPDFDetails{
		ID:        contentID,
		FileName:  gcsFileName,
		Extension: fileExtension,
		URL:       gcsURL,
		Size:      fileHeader.Size,
	}, nil
}

// La estructura InMemoryMultipartFile y sus métodos asociados
// (NewInMemoryMultipartFile, Read, Seek, Close, ReadAt, Name, Size)
// se eliminan de aquí, ya que se asume que están disponibles en el paquete 'services'
// (por ejemplo, desde audio_service.go o image_service.go).
// Si esto causa problemas de compilación, significa que necesitan ser movidos
// a un paquete de utilidades compartido e importado adecuadamente.
