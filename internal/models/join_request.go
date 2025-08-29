package models

import "time"

type JoinRequestStatus string

const (
	JoinRequestStatusPending  JoinRequestStatus = "pending"
	JoinRequestStatusApproved JoinRequestStatus = "approved"
	JoinRequestStatusRejected JoinRequestStatus = "rejected"
)

type JoinRequest struct {
	ID          int               `json:"id" db:"id"`
	RoomID      string            `json:"room_id" db:"room_id"`
	UserID      int               `json:"user_id" db:"user_id"`
	Status      JoinRequestStatus `json:"status" db:"status"`
	RequestedAt time.Time         `json:"requested_at" db:"requested_at"`
	HandledAt   *time.Time        `json:"handled_at,omitempty" db:"handled_at"`
	HandledBy   *int              `json:"handled_by,omitempty" db:"handled_by"`
	User        *User             `json:"user,omitempty"`
	Room        *Room             `json:"room,omitempty"`
}
