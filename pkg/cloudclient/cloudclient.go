package cloudclient

import (
	"context"
	"fmt"
	"io"
	"log" // Usar log estándar en lugar de tools
	"mime/multipart"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var bucket *storage.BucketHandle
var gcsBucketName string // Variable global para el nombre del bucket

// UploadFile sube un archivo a GCS.
func UploadFile(ctx context.Context, file multipart.File, remotePath string, contentType string) error {
	// Obtiene un writer para escribir el archivo en GCS.
	if bucket == nil {
		log.Printf("ERROR: GCS bucket handle is not initialized. Call Open() first.")
		return fmt.Errorf("GCS bucket handle not initialized")
	}
	wc := bucket.Object(remotePath).NewWriter(ctx)

	// Establece el tipo de contenido del archivo.
	wc.ContentType = contentType
	// Hacer público el archivo
	wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}

	// Rebobina el archivo al principio (importante si se leyó antes).
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Printf("ERROR: Failed to seek to start of file: %v", err)
		return err
	}

	// Copia el archivo local a GCS.
	if _, err := io.Copy(wc, file); err != nil {
		log.Printf("ERROR: Failed to copy file to GCS: %v", err)
		return err
	}

	// Cierra el writer para finalizar la subida del archivo.
	if err := wc.Close(); err != nil {
		log.Printf("ERROR: Failed to close GCS writer: %v", err)
		return err
	}

	log.Printf("File uploaded to gs://%s/%s", gcsBucketName, remotePath)
	return nil
}

// Open inicializa la conexión con el bucket de GCS y asigna la variable global bucket.
// TODO: Considerar devolver el handle en lugar de usar variable global.
func Open(bucketNameInput string, credentialsFile string) error {
	if bucket != nil {
		log.Println("GCS client already initialized.")
		return nil // Ya inicializado
	}
	ctx := context.Background()

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Printf("ERROR: Failed to create GCS client: %v", err)
		return fmt.Errorf("storage.NewClient: %w", err)
	}

	// Asignar a la variable global
	bucket = client.Bucket(bucketNameInput)
	gcsBucketName = bucketNameInput // Guardar el nombre del bucket globalmente
	log.Printf("GCS client initialized for bucket: %s", gcsBucketName)
	return nil
}

// GetBucketHandle devuelve el handle del bucket (si está inicializado).
// Puede ser útil si no se quiere depender de la variable global directamente.
func GetBucketHandle() *storage.BucketHandle {
	return bucket
}

// DownloadFile descarga un archivo de GCS.
func DownloadFile(ctx context.Context, remotePath string) ([]byte, error) {
	if bucket == nil {
		log.Printf("ERROR: GCS bucket handle is not initialized. Call Open() first.")
		return nil, fmt.Errorf("GCS bucket handle not initialized")
	}
	// Obtiene un reader para leer el archivo de GCS.
	rc, err := bucket.Object(remotePath).NewReader(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to create reader for %s: %v", remotePath, err)
		return nil, err
	}
	defer rc.Close()

	// Lee todo el contenido del reader en un arreglo de bytes.
	data, err := io.ReadAll(rc)
	if err != nil {
		log.Printf("ERROR: Failed to read file from GCS (%s): %v", remotePath, err)
		return nil, err
	}

	return data, nil
}
