package saveimage

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	// TODO: Importar cliente de GCS y librerías de imagen (ej. "image/jpeg", "nfnt/resize")
	// "cloud.google.com/go/storage"
	// "context"
	// "image"
	// "image/gif"
	// "image/jpeg"
	// "image/png"
	// "golang.org/x/image/draw"
)

// Config contiene la configuración necesaria para guardar imágenes.
type Config struct {
	StorageType string // "local" o "gcs"
	UploadDir   string // Directorio local si StorageType es "local"
	// GCSBucketName string // Nombre del bucket si StorageType es "gcs"
	// GCSClient *storage.Client // Cliente GCS preconfigurado
}

// SaveImageResult contiene el ID y la URL de la imagen guardada.
type SaveImageResult struct {
	MediaID string // ID único generado para la imagen (ej. UUID)
	FileURL string // URL pública o identificador para acceder a la imagen
}

// SaveImage procesa y guarda una imagen subida.
// Recibe el archivo, la configuración y las dimensiones deseadas (0 para no redimensionar).
func SaveImage(fileHeader *multipart.FileHeader, cfg Config, maxWidth, maxHeight uint) (*SaveImageResult, error) {
	log.Printf("SaveImage: Processing file '%s' (Size: %d)", fileHeader.Filename, fileHeader.Size)

	// 1. Abrir el archivo subido
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening uploaded file: %w", err)
	}
	defer file.Close()

	// 2. Validar tipo de archivo (opcional pero recomendado)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		return nil, fmt.Errorf("invalid file type: %s. Only jpg, png, gif allowed", ext)
	}

	// 3. Generar ID único y nombre de archivo
	mediaID := uuid.New().String()
	newFilename := mediaID + ext

	// --- Simulación de guardado (reemplazar con lógica real) ---
	var fileURL string
	switch cfg.StorageType {
	case "local":
		// Guardar localmente (asegurarse de que el directorio exista)
		if cfg.UploadDir == "" {
			cfg.UploadDir = "./uploads"
		}
		if err := os.MkdirAll(cfg.UploadDir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("error creating upload directory: %w", err)
		}
		dstPath := filepath.Join(cfg.UploadDir, newFilename)

		// Crear archivo destino
		dst, err := os.Create(dstPath)
		if err != nil {
			return nil, fmt.Errorf("error creating destination file: %w", err)
		}
		defer dst.Close()

		// Copiar contenido
		if _, err = io.Copy(dst, file); err != nil {
			return nil, fmt.Errorf("error saving file locally: %w", err)
		}

		fileURL = "/uploads/" + newFilename // URL relativa para servir localmente (ejemplo)
		log.Printf("SaveImage: Successfully saved file locally to %s", dstPath)

		// TODO: Implementar redimensionamiento si maxWidth/maxHeight > 0
		// - Decodificar la imagen (image.Decode)
		// - Redimensionar (ej. nfnt/resize)
		// - Guardar la imagen redimensionada (sobrescribir o nuevo archivo?)

	case "gcs":
		// TODO: Implementar lógica de subida a Google Cloud Storage
		// - Crear contexto (context.Background())
		// - Obtener el object handle (cfg.GCSClient.Bucket(cfg.GCSBucketName).Object(newFilename))
		// - Crear writer (obj.NewWriter(ctx))
		// - Establecer ACL (ej. writer.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}})
		// - Copiar el archivo (io.Copy(writer, file))
		// - Cerrar writer (writer.Close())
		// - Construir la URL pública (ej. fmt.Sprintf("https://storage.googleapis.com/%s/%s", cfg.GCSBucketName, newFilename))
		log.Printf("SaveImage: GCS storage not implemented. Skipping upload for %s", newFilename)
		fileURL = fmt.Sprintf("gcs://%s/%s", "SIMULATED_BUCKET", newFilename) // URL simulada

		// TODO: Implementar redimensionamiento si es necesario ANTES de subir

	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.StorageType)
	}

	// ----------------------------------------------------------

	result := &SaveImageResult{
		MediaID: mediaID,
		FileURL: fileURL,
	}

	return result, nil
}
