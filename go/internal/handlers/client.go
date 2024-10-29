package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"quantum-chat/internal/models"
	"time"

	"github.com/gorilla/websocket"
)

// newClient creates a new client instance
func newClient(userID int64, conn *websocket.Conn, hub *Hub) *Client {
    return &Client{
        UserID: userID,
        Conn:   conn,
        Send:   make(chan []byte, 256),
        hub:    hub,
    }
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.Conn.Close()
        log.Printf("Client readPump: Client %d disconnected", c.UserID)
    }()

    c.Conn.SetReadLimit(maxMessageSize)
    c.Conn.SetReadDeadline(time.Now().Add(pongWait))
    c.Conn.SetPongHandler(func(string) error {
        c.Conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, message, err := c.Conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("Client readPump: Error reading message: %v", err)
            }
            break
        }

        var wsMsg WSMessage
        if err := json.Unmarshal(message, &wsMsg); err != nil {
            log.Printf("Client readPump: Error parsing message: %v", err)
            c.sendError("Invalid message format")
            continue
        }

        // Set sender ID and timestamp
        wsMsg.SenderID = c.UserID
        wsMsg.Timestamp = time.Now().Unix()

        switch wsMsg.Type {
        case MessageTypeChat:
            if err := c.handleChatMessage(&wsMsg); err != nil {
                log.Printf("Client readPump: Error handling chat message: %v", err)
                c.sendError("Failed to process message")
            }
        default:
            c.sendError("Unknown message type")
        }
    }
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.Conn.Close()
        log.Printf("Client writePump: Connection closed for client %d", c.UserID)
    }()

    for {
        select {
        case message, ok := <-c.Send:
            c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                // The hub closed the channel.
                c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := c.Conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)

            // Add queued chat messages to the current websocket message.
            n := len(c.Send)
            for i := 0; i < n; i++ {
                w.Write(newline)
                w.Write(<-c.Send)
            }

            if err := w.Close(); err != nil {
                return
            }

        case <-ticker.C:
            c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

// handleChatMessage processes incoming chat messages
func (c *Client) handleChatMessage(wsMsg *WSMessage) error {
    msg := &models.Message{
        SenderID:   c.UserID,
        ReceiverID: wsMsg.ReceiverID,
        Content:    wsMsg.Content,
        Timestamp:  wsMsg.Timestamp,
        Read:       false,
    }

    // Save message to database
    if err := c.hub.handlers.db.SaveMessage(msg); err != nil {
        return fmt.Errorf("failed to save message: %v", err)
    }

    // Send acknowledgment to sender
    c.sendAck(msg.ID)

    // Forward message to recipient if online
    if recipient, ok := c.hub.clients[wsMsg.ReceiverID]; ok {
        wsMsg.MessageID = msg.ID
        messageJSON, _ := json.Marshal(wsMsg)
        recipient.Send <- messageJSON
    }

    return nil
}

// sendError sends an error message to the client
func (c *Client) sendError(message string) {
    errorMsg := WSMessage{
        Type:      MessageTypeError,
        Content:   json.RawMessage(fmt.Sprintf(`{"error":"%s"}`, message)),
        SenderID:  c.UserID,
        Timestamp: time.Now().Unix(),
    }
    messageJSON, _ := json.Marshal(errorMsg)
    c.Send <- messageJSON
}

// sendAck sends a message acknowledgment to the client
func (c *Client) sendAck(messageID int64) {
    ack := WSMessage{
        Type:      MessageTypeAck,
        Content:   json.RawMessage(fmt.Sprintf(`{"status":"delivered","message_id":%d}`, messageID)),
        SenderID:  c.UserID,
        Timestamp: time.Now().Unix(),
        MessageID: messageID,
    }
    messageJSON, _ := json.Marshal(ack)
    c.Send <- messageJSON
}