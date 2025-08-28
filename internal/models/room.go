package models

import "time"

type Role string

type Room struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Avatar    string    `json:"avatar" db:"avatar"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	OwnerId   string    `json:"ownerId" db:"owner_id"`
}
