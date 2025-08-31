package handlers

import (
	"awesomeProject1/internal/dto"
	"awesomeProject1/internal/repositories"
	"awesomeProject1/internal/services"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type WebSocketHandler struct {
	roomRepo      repositories.RoomRepository
	voteService   *services.VoteService
	clients       map[string]map[*websocket.Conn]int // roomID -> conn -> userID
	globalClients map[*websocket.Conn]bool           // connections listening to global updates
	mutex         sync.RWMutex
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
		roomRepo:      roomRepo,
		clients:       make(map[string]map[*websocket.Conn]int),
		globalClients: make(map[*websocket.Conn]bool),
	}
}

func (h *WebSocketHandler) WithVoteService(voteService *services.VoteService) *WebSocketHandler {
	h.voteService = voteService
	return h
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
		case "join_global_updates":
			h.handleJoinGlobalUpdates(c, msg)
		case "cast_vote":
			h.handleCastVote(c, msg)
		case "reveal_votes":
			h.handleRevealVotes(c, msg)
		case "reset_votes":
			h.handleResetVotes(c, msg)
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

	if userID == 0 {
		conn.WriteJSON(
			map[string]string{
				"error": "Invalid user ID",
			},
		)
		return
	}

	h.addClient(roomID, conn, userID)

	h.broadcastOnlineUsers(roomID)

	// Notify global clients that someone joined a room
	h.broadcastToGlobalClients(WebSocketMessage{
		Type:   "user_joined_room",
		RoomID: roomID,
		Data: map[string]interface{}{
			"user_id": userID,
		},
	})
}

func (h *WebSocketHandler) addClient(roomID string, conn *websocket.Conn, userID int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.clients[roomID] == nil {
		h.clients[roomID] = make(map[*websocket.Conn]int)
	}
	h.clients[roomID][conn] = userID
}

func (h *WebSocketHandler) handleJoinGlobalUpdates(conn *websocket.Conn, msg WebSocketMessage) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.globalClients[conn] = true

	conn.WriteJSON(map[string]string{
		"type":    "global_updates_joined",
		"message": "Connected to global updates",
	})
}

func (h *WebSocketHandler) removeClient(conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Remove from room clients
	for roomID, roomClients := range h.clients {
		if _, exists := roomClients[conn]; exists {
			delete(roomClients, conn)
			if len(roomClients) == 0 {
				delete(h.clients, roomID)
			} else {
				h.mutex.Unlock()
				h.broadcastOnlineUsers(roomID)
				h.broadcastToGlobalClients(WebSocketMessage{
					Type:   "user_left_room",
					RoomID: roomID,
				})
				h.mutex.Lock()
			}
			break
		}
	}

	// Remove from global clients
	delete(h.globalClients, conn)
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

func (h *WebSocketHandler) broadcastToGlobalClients(message WebSocketMessage) {
	h.mutex.RLock()
	globalClients := make(map[*websocket.Conn]bool)
	for conn, active := range h.globalClients {
		globalClients[conn] = active
	}
	h.mutex.RUnlock()

	for conn := range globalClients {
		if err := conn.WriteJSON(message); err != nil {
			log.Println("WebSocket global write error:", err)
			// Remove dead connection
			h.mutex.Lock()
			delete(h.globalClients, conn)
			h.mutex.Unlock()
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

func (h *WebSocketHandler) handleCastVote(conn *websocket.Conn, msg WebSocketMessage) {
	if h.voteService == nil {
		conn.WriteJSON(map[string]string{"error": "Vote service not available"})
		return
	}

	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		conn.WriteJSON(map[string]string{"error": "Invalid vote data"})
		return
	}

	userID := h.getUserIDFromConnection(conn)
	if userID == 0 {
		conn.WriteJSON(map[string]string{"error": "User not authenticated"})
		return
	}

	voteValue, ok := data["vote_value"].(string)
	if !ok {
		conn.WriteJSON(map[string]string{"error": "Vote value required"})
		return
	}

	request := dto.VoteRequest{VoteValue: voteValue}
	err := h.voteService.CastVote(userID, msg.RoomID, request)
	if err != nil {
		conn.WriteJSON(map[string]string{"error": err.Error()})
		return
	}

	// Broadcast to all clients in the room
	h.BroadcastVoteCast(msg.RoomID, userID, voteValue)

	// Send success response to the voting user
	conn.WriteJSON(
		map[string]interface{}{
			"type":    "vote_success",
			"message": "Vote cast successfully",
		},
	)
}

func (h *WebSocketHandler) handleRevealVotes(conn *websocket.Conn, msg WebSocketMessage) {
	if h.voteService == nil {
		conn.WriteJSON(map[string]string{"error": "Vote service not available"})
		return
	}

	userID := h.getUserIDFromConnection(conn)
	if userID == 0 {
		conn.WriteJSON(map[string]string{"error": "User not authenticated"})
		return
	}

	err := h.voteService.RevealVotes(userID, msg.RoomID)
	if err != nil {
		conn.WriteJSON(map[string]string{"error": err.Error()})
		return
	}

	h.BroadcastVotesRevealed(msg.RoomID)
}

func (h *WebSocketHandler) handleResetVotes(conn *websocket.Conn, msg WebSocketMessage) {
	if h.voteService == nil {
		conn.WriteJSON(map[string]string{"error": "Vote service not available"})
		return
	}

	userID := h.getUserIDFromConnection(conn)
	if userID == 0 {
		conn.WriteJSON(map[string]string{"error": "User not authenticated"})
		return
	}

	err := h.voteService.ResetVotes(userID, msg.RoomID)
	if err != nil {
		conn.WriteJSON(map[string]string{"error": err.Error()})
		return
	}

	h.BroadcastVotesReset(msg.RoomID)
}

func (h *WebSocketHandler) getUserIDFromConnection(conn *websocket.Conn) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, roomClients := range h.clients {
		if userID, exists := roomClients[conn]; exists {
			return userID
		}
	}
	return 0
}
