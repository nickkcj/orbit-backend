package websocket

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"nhooyr.io/websocket"
)

// Handler handles WebSocket connections
type Handler struct {
	hub           *Hub
	authenticator *Authenticator
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, authenticator *Authenticator) *Handler {
	return &Handler{
		hub:           hub,
		authenticator: authenticator,
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket
// Endpoint: GET /ws?token=xxx&tenant=slug
func (h *Handler) HandleWebSocket(c echo.Context) error {
	// Extract auth credentials
	token := c.QueryParam("token")
	if token == "" {
		// Try cookie as fallback
		if cookie, err := c.Cookie("auth_token"); err == nil {
			token = cookie.Value
		}
	}

	tenantSlug := c.QueryParam("tenant")
	if tenantSlug == "" {
		tenantSlug = c.Request().Header.Get("X-Tenant-Slug")
	}

	// Authenticate
	authResult, err := h.authenticator.Authenticate(c.Request().Context(), token, tenantSlug)
	if err != nil {
		log.Printf("[WS] Authentication failed: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	// Upgrade to WebSocket
	conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{
		InsecureSkipVerify: false,
		OriginPatterns:     []string{"*"}, // Configure based on environment
	})
	if err != nil {
		log.Printf("[WS] Failed to accept connection: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to upgrade connection"})
	}

	// Create client and register
	client := NewClient(h.hub, conn, authResult.TenantID, authResult.UserID)
	h.hub.Register(client)

	// Send connection confirmation
	connMsg, _ := NewMessage(MessageTypeConnected, ConnectedPayload{
		Status:   "connected",
		UserID:   authResult.UserID.String(),
		TenantID: authResult.TenantID.String(),
	})
	client.Send(connMsg)

	// Start read/write pumps
	go client.WritePump()
	client.ReadPump() // Blocks until disconnect

	return nil
}
