package models

type User struct {
	ID                int    `json:"id"`
	Username          string `json:"username"`
	Password          string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
	Email             string `json:"email"`
	Role              string `json:"role"`
	CreatedAt         string `json:"created_at"`
}
