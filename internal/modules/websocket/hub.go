package websocket

import (
	"sync"

	"github.com/google/uuid"
)

// Client represents one authenticated WebSocket connection.
type Client struct {
	UserID uuid.UUID
	Send   chan []byte
}

// Hub tracks live connections per user.
type Hub struct {
	mu          sync.RWMutex
	connections map[uuid.UUID]map[*Client]struct{}
}

// NewHub creates a WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		connections: make(map[uuid.UUID]map[*Client]struct{}),
	}
}

// Register adds a client connection.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections[client.UserID] == nil {
		h.connections[client.UserID] = make(map[*Client]struct{})
	}
	h.connections[client.UserID][client] = struct{}{}
}

// Unregister removes a client connection.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userConnections := h.connections[client.UserID]
	if userConnections == nil {
		return
	}

	delete(userConnections, client)
	close(client.Send)
	if len(userConnections) == 0 {
		delete(h.connections, client.UserID)
	}
}

// SendToUser delivers a payload to all active connections for a user.
func (h *Hub) SendToUser(userID uuid.UUID, payload []byte) {
	h.mu.RLock()
	userConnections := h.connections[userID]
	clients := make([]*Client, 0, len(userConnections))
	for client := range userConnections {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- payload:
		default:
			go h.Unregister(client)
		}
	}
}

// ConnectionCount returns active connections for a user.
func (h *Hub) ConnectionCount(userID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections[userID])
}
