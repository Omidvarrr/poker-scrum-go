package repositories

import (
	"awesomeProject1/internal/models"
	"errors"
)

var (
	ErrNotFound = errors.New("record not found")
)

type UserRepository interface {
	GetUserById(id int) (models.User, error)
	GetUserByEmail(email string) (models.User, error)
	CreateUser(user models.User) (models.User, error)
	UpdateUserProfile(id int, name string, avatar string) error
	CompleteUserProfile(id int, name string, avatar string) error
}

type RoomRepository interface {
	CreateRoom(room models.Room) (models.Room, error)
	GetRoomById(id string) (models.Room, error)
	GetRoomsByOwner(ownerID int) ([]models.Room, error)
	GetAllRooms() ([]models.Room, error)
	GetRoomsByUser(userID int) ([]models.Room, error)
	UpdateRoom(id string, name string, avatar string) error
	DeleteRoom(id string) error
	GetRoomMembers(roomID string) ([]models.RoomMember, error)
	AddRoomMember(roomID string, userID int, role models.Role) error
	RemoveRoomMember(roomID string, userID int) error
	UpdateMemberRole(roomID string, userID int, role models.Role) error
	IsUserInRoom(roomID string, userID int) (bool, error)
	IsUserAdmin(roomID string, userID int) (bool, error)
	IsUserOwner(roomID string, userID int) (bool, error)
}

type JoinRequestRepository interface {
	CreateJoinRequest(request models.JoinRequest) (models.JoinRequest, error)
	GetJoinRequestsByRoom(roomID string) ([]models.JoinRequest, error)
	GetJoinRequestsByUser(userID int) ([]models.JoinRequest, error)
	GetJoinRequestById(id int) (models.JoinRequest, error)
	UpdateJoinRequestStatus(id int, status models.JoinRequestStatus, handledBy int) error
	HasPendingRequest(roomID string, userID int) (bool, error)
}

type VoteRepository interface {
	CreateVoteSession(session models.VoteSession) (models.VoteSession, error)
	GetVoteSession(sessionID string) (models.VoteSession, error)
	GetActiveVoteSession(roomID string) (models.VoteSession, error)
	RevealVoteSession(sessionID string) error
	CastVote(vote models.Vote) error
	GetVotesBySession(sessionID string) ([]models.Vote, error)
	DeleteVotesBySession(sessionID string) error
}
