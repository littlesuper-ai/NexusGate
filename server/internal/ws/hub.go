package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub maintains connected WebSocket clients and broadcasts messages.
type Hub struct {
	mu        sync.RWMutex
	clients   map[*websocket.Conn]struct{}
	jwtSecret string
}

func NewHub(jwtSecret string) *Hub {
	return &Hub{
		clients:   make(map[*websocket.Conn]struct{}),
		jwtSecret: jwtSecret,
	}
}

// HandleWS upgrades HTTP to WebSocket and registers the client.
// Auth is required via query param: /ws?token=<jwt>
func (h *Hub) HandleWS(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token query param"})
		return
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()

	// Keep connection alive with ping/pong
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Ping ticker
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			h.mu.RLock()
			_, exists := h.clients[conn]
			h.mu.RUnlock()
			if !exists {
				return
			}
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}()

	// Read loop (discards incoming messages, detects disconnect)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	conn.Close()
}

// Broadcast sends a JSON message to all connected clients.
func (h *Hub) Broadcast(msgType string, data any) {
	payload, err := json.Marshal(map[string]any{
		"type":      msgType,
		"data":      data,
		"timestamp": time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			conn.Close()
			go func(c *websocket.Conn) {
				h.mu.Lock()
				delete(h.clients, c)
				h.mu.Unlock()
			}(conn)
		}
	}
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
