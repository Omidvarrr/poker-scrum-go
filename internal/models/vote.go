package models

import "time"

type Vote struct {
	ID        int       `json:"id" db:"id"`
	RoomID    string    `json:"room_id" db:"room_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	VoteValue string    `json:"vote_value" db:"vote_value"`
	SessionID string    `json:"session_id" db:"session_id"`
	VotedAt   time.Time `json:"voted_at" db:"voted_at"`
	User      *User     `json:"user,omitempty"`
}

type VoteSession struct {
	ID         string    `json:"id" db:"id"`
	RoomID     string    `json:"room_id" db:"room_id"`
	IsRevealed bool      `json:"is_revealed" db:"is_revealed"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	Votes      []Vote    `json:"votes,omitempty" gorm:"foreignKey:SessionID;references:ID"`
}
