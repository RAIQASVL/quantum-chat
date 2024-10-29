package handlers

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket message types
const (
    MessageTypeChat  = "chat"
    MessageTypeAck   = "ack"
    MessageTypeError = "error"
)

// WebSocket timeouts and limits
const (
    // Time allowed to write a message to the peer
    writeWait = 10 * time.Second

    // Time allowed to read the next pong message from the peer
    pongWait = 60 * time.Second

    // Send pings to peer with this period
    pingPeriod = (pongWait * 9) / 10

    // Maximum message size allowed from peer
    maxMessageSize = 512 * 1024
)

// WSMessage represents a WebSocket message
type WSMessage struct {
    Type       string          `json:"type"`
    Content    json.RawMessage `json:"content"`
    ReceiverID int64          `json:"receiver_id,omitempty"`
    SenderID   int64          `json:"sender_id,omitempty"`
    Timestamp  int64          `json:"timestamp,omitempty"`
    MessageID  int64          `json:"message_id,omitempty"`
}

// Client represents a connected WebSocket client
type Client struct {
    UserID int64
    Conn   *websocket.Conn
    Send   chan []byte
    hub    *Hub
}

var newline = []byte{'\n'}