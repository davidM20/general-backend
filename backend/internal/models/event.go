package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Event representa una notificación o evento en el sistema
type Event struct {
	Id             int64           `json:"id"`
	EventType      string          `json:"eventType"` // FRIEND_REQUEST, SYSTEM, EVENT, REQUEST_RESPONSE
	EventTitle     string          `json:"eventTitle"`
	Description    string          `json:"description"`
	UserId         int64           `json:"userId"`
	OtherUserId    sql.NullInt64   `json:"otherUserId"`
	ProyectId      sql.NullInt64   `json:"projectId"`
	CreateAt       time.Time       `json:"createAt"`
	IsRead         bool            `json:"isRead"`
	GroupId        sql.NullInt64   `json:"groupId"`
	Status         string          `json:"status"` // PENDING, ACCEPTED, REJECTED, CANCELLED
	ActionRequired bool            `json:"actionRequired"`
	ActionTakenAt  sql.NullTime    `json:"actionTakenAt"`
	Metadata       json.RawMessage `json:"metadata"`
}

// EventType constants
const (
	EventTypeFriendRequest   = "FRIEND_REQUEST"
	EventTypeSystem          = "SYSTEM"
	EventTypeEvent           = "EVENT"
	EventTypeRequestResponse = "REQUEST_RESPONSE"
)

// EventStatus constants
const (
	EventStatusPending   = "PENDING"
	EventStatusAccepted  = "ACCEPTED"
	EventStatusRejected  = "REJECTED"
	EventStatusCancelled = "CANCELLED"
)

// EventMetadata representa los datos adicionales específicos del tipo de evento
type EventMetadata struct {
	// Para solicitudes de amistad
	RequestMessage string `json:"requestMessage,omitempty"`
	ContactId      string `json:"contactId,omitempty"`

	// --- NUEVO CAMPO ---
	// ID del evento comunitario asociado a la notificación (ej. para una reseña).
	CommunityEventId int64 `json:"communityEventId,omitempty"`

	// IDs de los involucrados en una reseña, para facilitar el acceso en el cliente.
	// Quién emite la reseña (puede ser una empresa o un estudiante).
	ReviewerId int64 `json:"reviewerId,omitempty"`
	// Quién recibe la reseña.
	RevieweeId int64 `json:"revieweeId,omitempty"`

	// Para eventos del sistema
	SystemEventType string `json:"systemEventType,omitempty"`
	AdditionalData  any    `json:"additionalData,omitempty"`

	// Para eventos de proyecto
	EventDate     string `json:"eventDate,omitempty"`
	EventLocation string `json:"eventLocation,omitempty"`
}
