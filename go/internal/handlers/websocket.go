package handlers

import (
	"log"
	"net/http"
	"quantum-chat/internal/middleware"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // For development - make more restrictive in production
    },
    EnableCompression: true,
}

func (h *Handlers) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Get user ID from context (set by auth middleware)
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        log.Printf("WebSocket: No user ID in context")
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    log.Printf("WebSocket: Attempting connection for user %d", userID)

    // Log request headers for debugging
    log.Printf("WebSocket: Request headers:")
    for name, values := range r.Header {
        log.Printf("  %s: %v", name, values)
    }

    // Upgrade HTTP connection to WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket: Error upgrading connection: %v", err)
        return
    }

    log.Printf("WebSocket: Connection upgraded successfully for user %d", userID)

    // Create and register new client
    client := newClient(userID, conn, h.hub)
    h.hub.register <- client

    log.Printf("WebSocket: Client %d registered with hub", userID)

    // Start client message pumps in separate goroutines
    go client.writePump()
    go client.readPump()
}