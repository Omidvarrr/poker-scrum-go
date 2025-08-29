package handlers

import (
	"awesomeProject1/internal/repositories"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type WebSocketHandler struct {
	roomRepo repositories.RoomRepository
	clients  map[string]map[*websocket.Conn]int // roomID -> conn -> userID
	mutex    sync.RWMutex
}

type WebSocketMessage struct {
	Type   string      `json:"type"`
	RoomID string      `json:"room_id,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

type OnlineUsersMessage struct {
	Type        string `json:"type"`
	OnlineUsers []int  `json:"online_users"`
}

func NewWebSocketHandler(roomRepo repositories.RoomRepository) *WebSocketHandler {
	return &WebSocketHandler{
		roomRepo: roomRepo,
		clients:  make(map[string]map[*websocket.Conn]int),
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *websocket.Conn) {
	defer func() {
		h.removeClient(c)
		c.Close()
	}()

	for {
		var msg WebSocketMessage
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		switch msg.Type {
		case "join_room":
			h.handleJoinRoom(c, msg)
		case "vote_cast":
			h.broadcastToRoom(msg.RoomID, msg)
		case "votes_revealed":
			h.broadcastToRoom(msg.RoomID, msg)
		case "votes_reset":
			h.broadcastToRoom(msg.RoomID, msg)
		}
	}
}

func (h *WebSocketHandler) handleJoinRoom(conn *websocket.Conn, msg WebSocketMessage) {
	roomID := msg.RoomID
	if roomID == "" {
		return
	}

	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}

	userIDFloat, ok := data["user_id"].(float64)
	if !ok {
		return
	}
	userID := int(userIDFloat)

	isMember, err := h.roomRepo.IsUserInRoom(roomID, userID)
	if err != nil || !isMember {
		conn.WriteJSON(map[string]string{
			"error": "User is not a member of this room",
		})
		return
	}

	h.addClient(roomID, conn, userID)

	h.broadcastOnlineUsers(roomID)
}

func (h *WebSocketHandler) addClient(roomID string, conn *websocket.Conn, userID int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.clients[roomID] == nil {
		h.clients[roomID] = make(map[*websocket.Conn]int)
	}
	h.clients[roomID][conn] = userID
}

func (h *WebSocketHandler) removeClient(conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for roomID, roomClients := range h.clients {
		if _, exists := roomClients[conn]; exists {
			delete(roomClients, conn)
			if len(roomClients) == 0 {
				delete(h.clients, roomID)
			} else {
				h.mutex.Unlock()
				h.broadcastOnlineUsers(roomID)
				h.mutex.Lock()
			}
			break
		}
	}
}

func (h *WebSocketHandler) broadcastToRoom(roomID string, message WebSocketMessage) {
	h.mutex.RLock()
	roomClients := h.clients[roomID]
	h.mutex.RUnlock()

	if roomClients == nil {
		return
	}

	for conn := range roomClients {
		if err := conn.WriteJSON(message); err != nil {
			log.Println("WebSocket write error:", err)
		}
	}
}

func (h *WebSocketHandler) broadcastOnlineUsers(roomID string) {
	h.mutex.RLock()
	roomClients := h.clients[roomID]
	h.mutex.RUnlock()

	if roomClients == nil {
		return
	}

	var onlineUsers []int
	for _, userID := range roomClients {
		onlineUsers = append(onlineUsers, userID)
	}

	message := OnlineUsersMessage{
		Type:        "online_users_update",
		OnlineUsers: onlineUsers,
	}

	for conn := range roomClients {
		if err := conn.WriteJSON(message); err != nil {
			log.Println("WebSocket write error:", err)
		}
	}
}

func (h *WebSocketHandler) BroadcastVoteCast(roomID string, userID int, vote string) {
	message := WebSocketMessage{
		Type:   "vote_cast",
		RoomID: roomID,
		Data: map[string]interface{}{
			"user_id": userID,
			"vote":    vote,
		},
	}
	h.broadcastToRoom(roomID, message)
}

func (h *WebSocketHandler) BroadcastVotesRevealed(roomID string) {
	message := WebSocketMessage{
		Type:   "votes_revealed",
		RoomID: roomID,
	}
	h.broadcastToRoom(roomID, message)
}

func (h *WebSocketHandler) BroadcastVotesReset(roomID string) {
	message := WebSocketMessage{
		Type:   "votes_reset",
		RoomID: roomID,
	}
	h.broadcastToRoom(roomID, message)
}
