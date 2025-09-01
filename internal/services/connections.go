package services

import (
	"sync"
)

// RoomConnections manages the set of active user IDs for each room.
type RoomConnections struct {
	mu    sync.RWMutex
	rooms map[string]map[int]struct{} // roomID -> set of userIDs
}

func NewRoomConnections() *RoomConnections {
	return &RoomConnections{
		rooms: make(map[string]map[int]struct{}),
	}
}

// AddUser adds a user to a room.
func (rc *RoomConnections) AddUser(roomID string, userID int) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if _, ok := rc.rooms[roomID]; !ok {
		rc.rooms[roomID] = make(map[int]struct{})
	}
	rc.rooms[roomID][userID] = struct{}{}
}

// RemoveUser removes a user from a room.
func (rc *RoomConnections) RemoveUser(roomID string, userID int) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if users, ok := rc.rooms[roomID]; ok {
		delete(users, userID)
		if len(users) == 0 {
			delete(rc.rooms, roomID)
		}
	}
}

// GetActiveRoomIDs returns a list of room IDs that have at least one user.
func (rc *RoomConnections) GetActiveRoomIDs() []string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	activeRooms := make([]string, 0, len(rc.rooms))
	for roomID := range rc.rooms {
		activeRooms = append(activeRooms, roomID)
	}
	return activeRooms
}

// GetUsersInRoom returns a list of user IDs currently active in a specific room.
func (rc *RoomConnections) GetUsersInRoom(roomID string) []int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	if users, ok := rc.rooms[roomID]; ok {
		userIDs := make([]int, 0, len(users))
		for userID := range users {
			userIDs = append(userIDs, userID)
		}
		return userIDs
	}

	return []int{}
}

// Hub manages all WebSocket connections.
var ActiveConnections = NewRoomConnections()
