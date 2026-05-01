package user

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	Username     string    `json:"username"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	Provider     string    `json:"provider"`
	PasswordHash *string   `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
