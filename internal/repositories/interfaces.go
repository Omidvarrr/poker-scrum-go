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
	UpdateUserProfile(id int, firstName string, lastName string, avatar string) error
	CompleteUserProfile(id int, firstName string, lastName string, avatar string) error
}

type RoomRepository interface {
	CreateRoom(room models.Room) (models.Room, error)
	GetRoomById(id string) (models.Room, error)
	GetRoomsByOwner(ownerID int) ([]models.Room, error)
	GetAllRooms() ([]models.Room, error)
	GetRoomsByIDs(ids []string) ([]models.Room, error)
	UpdateRoom(id string, name string, avatar string) error
	DeleteRoom(id string) error
	IsUserOwner(roomID string, userID int) (bool, error)
}

type VoteRepository interface {
	CreateVoteSession(session models.VoteSession) (models.VoteSession, error)
	GetActiveVoteSession(roomID string) (models.VoteSession, error)
	RevealVoteSession(sessionID string) error
	CastVote(vote models.Vote) error
	RemoveVote(userID int, roomID string, sessionID string) error
	DeleteVotesBySession(sessionID string) error
}
