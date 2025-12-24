package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

const (
	writeTimeout = 10 * time.Second
	pingInterval = 30 * time.Second // Railway requires keep-alive
	pongWait     = 60 * time.Second
)

// Client represents a WebSocket connection
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	tenantID uuid.UUID
	userID   uuid.UUID
	send     chan *Message
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, tenantID, userID uuid.UUID) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		hub:      hub,
		conn:     conn,
		tenantID: tenantID,
		userID:   userID,
		send:     make(chan *Message, 256),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.hub.unregister <- c
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			ctx, cancel := context.WithTimeout(c.ctx, writeTimeout)
			data, err := json.Marshal(msg)
			if err != nil {
				cancel()
				log.Printf("[WS] Failed to marshal message: %v", err)
				continue
			}
			err = c.conn.Write(ctx, websocket.MessageText, data)
			cancel()
			if err != nil {
				log.Printf("[WS] Failed to write message: %v", err)
				return
			}
		case <-ticker.C:
			// Send ping for Railway keep-alive
			ctx, cancel := context.WithTimeout(c.ctx, writeTimeout)
			err := c.conn.Ping(ctx)
			cancel()
			if err != nil {
				log.Printf("[WS] Ping failed: %v", err)
				return
			}
		}
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.cancel()
		c.hub.unregister <- c
	}()

	for {
		_, data, err := c.conn.Read(c.ctx)
		if err != nil {
			if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
				log.Printf("[WS] Read error: %v", err)
			}
			return
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("[WS] Failed to unmarshal message: %v", err)
			continue
		}

		c.handleMessage(&msg)
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case MessageTypePing:
		// Respond with pong
		pong, _ := NewMessage(MessageTypePong, nil)
		c.send <- pong
	default:
		// Handle other message types as needed
		log.Printf("[WS] Received message type: %s", msg.Type)
	}
}

// Send sends a message to the client
func (c *Client) Send(msg *Message) {
	select {
	case c.send <- msg:
	default:
		// Buffer full, client is slow
		log.Printf("[WS] Client buffer full, dropping message")
	}
}

// Close closes the client connection
func (c *Client) Close() {
	c.cancel()
	close(c.send)
}
