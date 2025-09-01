package models

import "time"

type User struct {
	ID               int       `json:"id" db:"id"`
	Email            string    `json:"email" db:"email"`
	FirstName        string    `json:"first_name" db:"first_name"`
	LastName         string    `json:"last_name" db:"last_name"`
	Avatar           string    `json:"avatar" db:"avatar"`
	PasswordHash     string    `json:"-" db:"password_hash"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	ProfileCompleted bool      `json:"profile_completed" db:"profile_completed"`
}
