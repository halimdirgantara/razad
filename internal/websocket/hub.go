// Package websocket provides a real-time event hub for streaming logs
// and metrics to connected browser clients.
package websocket

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	gws "github.com/gorilla/websocket"
)

var upgrader = gws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// TODO: restrict to configured allowed origins in production
	CheckOrigin: func(r *http.Request) bool { return true },
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		http.Error(w, reason.Error(), status)
	},
}

// Message is a generic WebSocket message.
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Client represents a single WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *gws.Conn
	send   chan []byte
	appID  string // empty = all apps
	done   chan struct{}
}

// Hub manages all active WebSocket connections and message routing.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]bool
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
	}
}

// HandleConnection upgrades an HTTP connection to WebSocket.
func (h *Hub) HandleConnection(w http.ResponseWriter, r *http.Request) {
	hasHijacker := false
	if _, ok := w.(http.Hijacker); ok {
		hasHijacker = true
	}
	slog.Debug("websocket: handler entry", "path", r.URL.Path, "writer", fmt.Sprintf("%T", w), "hijacker", hasHijacker)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("websocket: upgrade failed", "error", err)
		http.Error(w, "websocket upgrade failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	appID := r.URL.Query().Get("app")

	client := &Client{
		hub:   h,
		conn:  conn,
		send:  make(chan []byte, 256),
		appID: appID,
		done:  make(chan struct{}),
	}

	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()

	slog.Debug("websocket: client connected", "app", appID)

	go client.writePump()
	go client.readPump()
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			// Client's send buffer is full — skip
		}
	}
}

// BroadcastToApp sends a message to clients subscribed to a specific app.
func (h *Hub) BroadcastToApp(appID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.appID != "" && client.appID != appID {
			continue
		}
		select {
		case client.send <- data:
		default:
		}
	}
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// removeClient unregisters a client from the hub.
func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	delete(h.clients, client)
	h.mu.Unlock()
	close(client.done)
	slog.Debug("websocket: client disconnected")
}

// readPump reads messages from the client (keep-alive handling).
func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		c.hub.removeClient(c)
	}()

	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// writePump sends messages to the client connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second) // ping interval
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(gws.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(gws.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(gws.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}
