package models

import "time"

// Notification representa una notificación en el sistema
type Notification struct {
	NotificationId string     `json:"notificationId"`
	UserId         int64      `json:"userId"`
	Type           string     `json:"type"`
	Title          string     `json:"title"`
	Message        string     `json:"message"`
	IsRead         bool       `json:"isRead"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	OtherUserId    int64      `json:"otherUserId"`
	ActionRequired bool       `json:"actionRequired"`
	Status         string     `json:"status"`
	ActionTakenAt  *time.Time `json:"actionTakenAt,omitempty"`
}
