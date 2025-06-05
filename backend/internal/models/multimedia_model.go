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
	Type      string        `json:"type" db_field:"Type" sql_type:"VARCHAR(255)"` // Ej: "image", "video"
	Ratio     float32       `json:"ratio" db_field:"Ratio" sql_type:"FLOAT"`
	UserId    int64         `json:"user_id" db_field:"UserId" sql_type:"BIGINT"`
	FileName  string        `json:"file_name" db_field:"FileName" sql_type:"VARCHAR(255)"`             // Nombre del archivo en el almacenamiento en la nube
	CreateAt  time.Time     `json:"create_at" db_field:"CreateAt" sql_type:"DATE"`                     // Nota: schema.sql dice DATE, pero time.Time es más flexible. Se guardará como DATE.
	ContentId string        `json:"content_id,omitempty" db_field:"ContentId" sql_type:"VARCHAR(255)"` // Para agrupar diferentes versiones/resoluciones de un mismo contenido
	ChatId    string        `json:"chat_id,omitempty" db_field:"ChatId" sql_type:"VARCHAR(255)"`
	Size      sql.NullInt64 `json:"size,omitempty"` // Tamaño del archivo en bytes
}
