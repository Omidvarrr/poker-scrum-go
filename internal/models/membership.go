package models

import "time"

const (
	RoleMember Role = "member"
	RoleAdmin  Role = "admin"
)

type RoomMember struct {
	RoomID   string    `db:"room_id"`
	UserID   int       `db:"user_id"`
	JoinedAt time.Time `db:"joined_at"`
	Role     Role      `db:"role"`
}
