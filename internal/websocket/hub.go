package websocket

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
)

// BroadcastMessage sends to all clients in a tenant
type BroadcastMessage struct {
	TenantID uuid.UUID
	Message  *Message
}

// DirectMessage sends to a specific user (all their connections)
type DirectMessage struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Message  *Message
}

// Hub manages all WebSocket connections across tenants
type Hub struct {
	// Tenant rooms: tenantID -> set of clients
	tenantRooms map[uuid.UUID]map[*Client]struct{}

	// User channels: userID -> set of clients (user may have multiple tabs)
	userChannels map[uuid.UUID]map[*Client]struct{}

	// Channels for client registration/unregistration
	register   chan *Client
	unregister chan *Client

	// Broadcast channel for tenant-wide messages
	broadcast chan *BroadcastMessage

	// Direct message channel for user-specific messages
	direct chan *DirectMessage

	mu sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		tenantRooms:  make(map[uuid.UUID]map[*Client]struct{}),
		userChannels: make(map[uuid.UUID]map[*Client]struct{}),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		broadcast:    make(chan *BroadcastMessage, 256),
		direct:       make(chan *DirectMessage, 256),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case msg := <-h.broadcast:
			h.broadcastToTenant(msg)
		case msg := <-h.direct:
			h.sendToUser(msg)
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to tenant room
	if h.tenantRooms[client.tenantID] == nil {
		h.tenantRooms[client.tenantID] = make(map[*Client]struct{})
	}
	h.tenantRooms[client.tenantID][client] = struct{}{}

	// Add to user channels
	if h.userChannels[client.userID] == nil {
		h.userChannels[client.userID] = make(map[*Client]struct{})
	}
	h.userChannels[client.userID][client] = struct{}{}

	log.Printf("[WS] Client registered: user=%s tenant=%s", client.userID, client.tenantID)
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove from tenant room
	if clients, ok := h.tenantRooms[client.tenantID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.tenantRooms, client.tenantID)
		}
	}

	// Remove from user channels
	if clients, ok := h.userChannels[client.userID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.userChannels, client.userID)
		}
	}

	log.Printf("[WS] Client unregistered: user=%s tenant=%s", client.userID, client.tenantID)
}

// broadcastToTenant sends a message to all clients in a tenant
func (h *Hub) broadcastToTenant(msg *BroadcastMessage) {
	h.mu.RLock()
	clients := h.tenantRooms[msg.TenantID]
	h.mu.RUnlock()

	for client := range clients {
		client.Send(msg.Message)
	}
}

// sendToUser sends a message to all connections of a specific user
func (h *Hub) sendToUser(msg *DirectMessage) {
	h.mu.RLock()
	clients := h.userChannels[msg.UserID]
	h.mu.RUnlock()

	for client := range clients {
		// Only send if client is in the correct tenant
		if client.tenantID == msg.TenantID {
			client.Send(msg.Message)
		}
	}
}

// BroadcastToTenant is a public method to send messages to all clients in a tenant
func (h *Hub) BroadcastToTenant(tenantID uuid.UUID, msg *Message) {
	h.broadcast <- &BroadcastMessage{TenantID: tenantID, Message: msg}
}

// SendToUser is a public method to send messages to a specific user
func (h *Hub) SendToUser(tenantID, userID uuid.UUID, msg *Message) {
	h.direct <- &DirectMessage{TenantID: tenantID, UserID: userID, Message: msg}
}

// GetTenantClientCount returns the number of connected clients for a tenant
func (h *Hub) GetTenantClientCount(tenantID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.tenantRooms[tenantID])
}

// GetUserClientCount returns the number of connected clients for a user
func (h *Hub) GetUserClientCount(userID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userChannels[userID])
}

// Register adds a client to the hub (public method)
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub (public method)
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
