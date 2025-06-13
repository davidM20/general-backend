package services

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"time"

	"github.com/chai2010/webp"
	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/cloudclient"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
)

const (
	lowResWidth    = 150 // Ancho en píxeles para baja resolución
	mediumResWidth = 600 // Ancho en píxeles para media resolución
	outputFormat   = "webp"
)

// ImageUploadService encapsula la lógica para subir y procesar imágenes.
type ImageUploadService struct {
	db  *sql.DB
	cfg *config.Config
	// No necesitamos el bucketHandle aquí si cloudclient usa variables globales
	// o si cloudclient.UploadFile es una función estática/paquete.
}

// NewImageUploadService crea una nueva instancia de ImageUploadService.
func NewImageUploadService(db *sql.DB, cfg *config.Config) *ImageUploadService {
	return &ImageUploadService{db: db, cfg: cfg}
}

// InMemoryMultipartFile es un adaptador para io.Reader a multipart.File para el cloudclient.
type InMemoryMultipartFile struct {
	Reader   *bytes.Reader // Cambiado a *bytes.Reader para asegurar que ReadAt está disponible
	FileName string
}

func NewInMemoryMultipartFile(data []byte, fileName string) *InMemoryMultipartFile {
	return &InMemoryMultipartFile{Reader: bytes.NewReader(data), FileName: fileName}
}

func (imf *InMemoryMultipartFile) Read(p []byte) (n int, err error) {
	return imf.Reader.Read(p)
}

func (imf *InMemoryMultipartFile) ReadAt(p []byte, off int64) (n int, err error) {
	return imf.Reader.ReadAt(p, off) // Delegar a bytes.Reader.ReadAt
}

func (imf *InMemoryMultipartFile) Seek(offset int64, whence int) (int64, error) {
	return imf.Reader.Seek(offset, whence) // Delegar a bytes.Reader.Seek
}

func (imf *InMemoryMultipartFile) Close() error {
	// No-op para un reader en memoria como bytes.Reader
	return nil
}

// UploadImageDetails contiene la información de la imagen subida para la respuesta.
type UploadImageDetails struct {
	ID        string  `json:"id"`        // ID del contenido principal (original)
	FileName  string  `json:"fileName"`  // Nombre del archivo original subido (ej: abc.webp)
	Extension string  `json:"extension"` // Siempre "webp"
	URL       string  `json:"url"`       // URL de la imagen original en GCS
	Ratio     float32 `json:"ratio"`
}

// ProcessAndUploadImage procesa una imagen subida, la convierte a webp, crea variantes y las sube.
func (s *ImageUploadService) ProcessAndUploadImage(ctx context.Context, userID int64, file multipart.File, fileHeader *multipart.FileHeader) (*UploadImageDetails, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Errorf("ProcessAndUploadImage", "Error leyendo el archivo: %v", err)
		return nil, fmt.Errorf("error al leer el archivo: %w", err)
	}

	kind, err := filetype.Match(fileBytes)
	if err != nil || kind == types.Unknown {
		logger.Warnf("ProcessAndUploadImage", "Tipo de archivo desconocido o no soportado. Error: %v, Kind: %s", err, kind.Extension)
		return nil, fmt.Errorf("tipo de archivo no soportado o inválido")
	}

	if !filetype.IsImage(fileBytes) {
		logger.Warnf("ProcessAndUploadImage", "El archivo no es una imagen soportada. Tipo detectado: %s", kind.MIME.Value)
		return nil, fmt.Errorf("el archivo no es una imagen soportada: %s", kind.MIME.Value)
	}

	img, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		logger.Errorf("ProcessAndUploadImage", "Error decodificando la imagen original (tipo: %s): %v", kind.MIME.Value, err)
		return nil, fmt.Errorf("error al decodificar la imagen: %w", err)
	}

	ratio := float32(0.0)
	if img.Bounds().Dy() != 0 {
		ratio = float32(img.Bounds().Dx()) / float32(img.Bounds().Dy())
	} else if img.Bounds().Dx() != 0 {
		ratio = 1.0 // O algún valor que indique que es muy ancha o un error
	} // si Dx y Dy son 0, ratio se queda en 0.0, lo que podría indicar un problema

	contentID := uuid.New().String()
	baseFileName := uuid.New().String()
	var originalGCSUrl string

	originalWebPBytes, err := s.convertToWebP(img)
	if err != nil {
		return nil, fmt.Errorf("error convirtiendo original a WebP: %w", err)
	}
	originalFileName := baseFileName + "." + outputFormat
	// Usar el constructor para InMemoryMultipartFile
	err = cloudclient.UploadFile(ctx, NewInMemoryMultipartFile(originalWebPBytes, originalFileName), originalFileName, "image/webp")
	if err != nil {
		return nil, fmt.Errorf("error subiendo original WebP a GCS: %w", err)
	}
	originalGCSUrl = fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.cfg.GCSBucketName, originalFileName)
	_, err = queries.InsertMultimedia(s.db, &models.Multimedia{
		Id:        uuid.New().String(),
		Type:      "image",
		Ratio:     ratio,
		UserId:    userID,
		FileName:  originalFileName,
		CreateAt:  time.Now(),
		ContentId: contentID,
	})
	if err != nil {
		return nil, fmt.Errorf("error guardando registro original en BD: %w", err)
	}

	// Procesar y subir variante de baja resolución
	lowResImg := s.resizeImage(img, lowResWidth)
	lowResWebPBytes, err := s.convertToWebP(lowResImg)
	if err != nil {
		logger.Warnf("ProcessAndUploadImage", "Error convirtiendo low-res a WebP: %v", err)
	} else {
		lowResFileName := "low-" + originalFileName
		err = cloudclient.UploadFile(ctx, NewInMemoryMultipartFile(lowResWebPBytes, lowResFileName), lowResFileName, "image/webp")
		if err != nil {
			logger.Warnf("ProcessAndUploadImage", "Error subiendo low-res WebP a GCS: %v", err)
		} else {
			_, errDb := queries.InsertMultimedia(s.db, &models.Multimedia{
				Id:        uuid.New().String(),
				Type:      "image_low_res",
				Ratio:     ratio,
				UserId:    userID,
				FileName:  lowResFileName,
				CreateAt:  time.Now(),
				ContentId: contentID,
			})
			if errDb != nil {
				logger.Warnf("ProcessAndUploadImage", "Error guardando registro low-res en BD: %v", errDb)
			}
		}
	}

	// Procesar y subir variante de media resolución
	mediumResImg := s.resizeImage(img, mediumResWidth)
	mediumResWebPBytes, err := s.convertToWebP(mediumResImg)
	if err != nil {
		logger.Warnf("ProcessAndUploadImage", "Error convirtiendo medium-res a WebP: %v", err)
	} else {
		mediumResFileName := "medium-" + originalFileName
		err = cloudclient.UploadFile(ctx, NewInMemoryMultipartFile(mediumResWebPBytes, mediumResFileName), mediumResFileName, "image/webp")
		if err != nil {
			logger.Warnf("ProcessAndUploadImage", "Error subiendo medium-res WebP a GCS: %v", err)
		} else {
			_, errDb := queries.InsertMultimedia(s.db, &models.Multimedia{
				Id:        uuid.New().String(),
				Type:      "image_medium_res",
				Ratio:     ratio,
				UserId:    userID,
				FileName:  mediumResFileName,
				CreateAt:  time.Now(),
				ContentId: contentID,
			})
			if errDb != nil {
				logger.Warnf("ProcessAndUploadImage", "Error guardando registro medium-res en BD: %v", errDb)
			}
		}
	}

	return &UploadImageDetails{
		ID:        contentID,
		FileName:  originalFileName,
		Extension: outputFormat,
		URL:       originalGCSUrl,
		Ratio:     ratio,
	}, nil
}

// UpdateUserProfilePicture actualiza el campo de la imagen de perfil en la base de datos.
func (s *ImageUploadService) UpdateUserProfilePicture(ctx context.Context, userID int64, pictureFileName string) error {
	err := queries.UpdateUserPicture(userID, pictureFileName)
	if err != nil {
		// El error ya se ha logueado en la capa de queries, aquí solo lo envolvemos para dar contexto de servicio.
		return fmt.Errorf("servicio falló al actualizar la foto de perfil para el usuario %d: %w", userID, err)
	}
	logger.Infof("UpdateUserProfilePicture", "Foto de perfil actualizada exitosamente para el usuario %d con el archivo %s", userID, pictureFileName)
	return nil
}

func (s *ImageUploadService) convertToWebP(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 80}); err != nil {
		logger.Errorf("convertToWebP", "Error codificando a WebP: %v", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *ImageUploadService) resizeImage(originalImage image.Image, width int) image.Image {
	origBounds := originalImage.Bounds()
	origWidth := origBounds.Dx()
	origHeight := origBounds.Dy()

	if origWidth == 0 || origHeight == 0 || width <= 0 {
		logger.Warnf("resizeImage", "Dimensiones originales inválidas (%dx%d) o ancho de destino inválido (%d).", origWidth, origHeight, width)
		// Devolver la imagen original sin modificar o una imagen mínima para evitar pánico.
		if origWidth == 0 || origHeight == 0 {
			return image.NewRGBA(image.Rect(0, 0, 1, 1)) // Imagen mínima
		}
		return originalImage // Devolver original si el nuevo ancho es inválido
	}

	height := (width * origHeight) / origWidth
	if height <= 0 {
		logger.Warnf("resizeImage", "Altura calculada inválida (%d) para ancho %d. Usando altura mínima de 1.", height, width)
		height = 1 // Asegurar que la altura no sea 0 o negativa
	}

	dstRect := image.Rect(0, 0, width, height)
	dstImage := image.NewRGBA(dstRect)

	draw.CatmullRom.Scale(dstImage, dstRect, originalImage, origBounds, draw.Over, nil)
	return dstImage
}

// getFileExtension no es necesaria aquí ya que siempre guardamos como .webp
// func getFileExtension(fileName string) string { ... }
