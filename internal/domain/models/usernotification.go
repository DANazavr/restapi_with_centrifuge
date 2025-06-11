package models

type UserNotification struct {
	UID          int                    `json:"uid" db:"uid"`
	UserID       int                    `json:"user_id" db:"user_id"`
	CreatedAt    string                 `json:"created_at" db:"created_at"`
	SendAt       string                 `json:"send_at" db:"send_at"`
	ReadAt       string                 `json:"read_at" db:"read_at"`
	Notification map[string]interface{} `json:"notification" db:"notification"`
}
