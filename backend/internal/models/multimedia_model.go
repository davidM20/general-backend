package models

import (
	"database/sql"
	"time"
)

/*
 * ===================================================
 * MODELO DE DATOS PARA MULTIMEDIA
 * ===================================================
 *
 * Este archivo define la estructura de datos para la tabla Multimedia,
 * que almacena información sobre los archivos subidos (imágenes, videos, etc.).
 */

// Multimedia representa la estructura de la tabla Multimedia en la base de datos.
type Multimedia struct {
	Id        string        `json:"id" db_field:"Id" sql_type:"VARCHAR(255)"`
	Type      string        `json:"type" db_field:"Type" sql_type:"VARCHAR(255)"` // Ej: "image", "video", "audio", "pdf"
	Ratio     float32       `json:"ratio" db_field:"Ratio" sql_type:"FLOAT"`      // Aspect ratio para imágenes/videos
	UserId    int64         `json:"user_id" db_field:"UserId" sql_type:"BIGINT"`
	FileName  string        `json:"file_name" db_field:"FileName" sql_type:"VARCHAR(255)"` // Nombre del archivo original en GCS para tipo="video", o el único archivo para otros tipos
	CreateAt  time.Time     `json:"create_at" db_field:"CreateAt" sql_type:"TIMESTAMP"`
	ContentId string        `json:"content_id,omitempty" db_field:"ContentId" sql_type:"VARCHAR(255)"` // ID para agrupar diferentes versiones/resoluciones (especialmente videos)
	ChatId    string        `json:"chat_id,omitempty" db_field:"ChatId" sql_type:"VARCHAR(255)"`
	Size      sql.NullInt64 `json:"size,omitempty" db_field:"Size" sql_type:"BIGINT"` // Tamaño del archivo original en bytes

	// Campos específicos para videos y su procesamiento
	ProcessingStatus   sql.NullString  `json:"processing_status,omitempty" db_field:"ProcessingStatus" sql_type:"VARCHAR(50)"`        // Ej: uploaded, processing, completed, failed
	Duration           sql.NullFloat64 `json:"duration,omitempty" db_field:"Duration" sql_type:"FLOAT"`                               // Duración del video en segundos
	HLSManifestBaseURL sql.NullString  `json:"hls_manifest_base_url,omitempty" db_field:"HLSManifestBaseURL" sql_type:"VARCHAR(512)"` // URL base en GCS para manifiestos y segmentos HLS
	HLSManifest1080p   sql.NullString  `json:"hls_manifest_1080p,omitempty" db_field:"HLSManifest1080p" sql_type:"VARCHAR(255)"`      // Path relativo al BaseURL para 1080p (ej. 1080p/playlist.m3u8)
	HLSManifest720p    sql.NullString  `json:"hls_manifest_720p,omitempty" db_field:"HLSManifest720p" sql_type:"VARCHAR(255)"`        // Path relativo para 720p
	HLSManifest480p    sql.NullString  `json:"hls_manifest_480p,omitempty" db_field:"HLSManifest480p" sql_type:"VARCHAR(255)"`        // Path relativo para 480p
	// Podríamos añadir más campos para DASH si fuera necesario
}
