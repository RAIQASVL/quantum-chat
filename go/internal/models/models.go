package models

import "encoding/json"

type User struct {
    ID        int64  `json:"id"`
    Username  string `json:"username"`
    Password  string `json:"-"` // Password hash, not exposed in JSON
    PublicKey []byte `json:"public_key"`
}

type Message struct {
    ID         int64  `json:"id"`
    SenderID   int64  `json:"sender_id"`
    ReceiverID int64  `json:"receiver_id"`
    Content    []byte `json:"content"`
    Timestamp  int64  `json:"timestamp"`
    Read       bool   `json:"read"`
}

type WSMessage struct {
    Type     string          `json:"type"`
    Content  json.RawMessage `json:"content"`
    Receiver int64          `json:"receiver_id"`
}

const (
    MessageTypeChat   = "chat"
    MessageTypeSystem = "system"
)

// SQL migrations
const (
    CreateTablesSQL = `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(255) UNIQUE NOT NULL,
        password VARCHAR(255) NOT NULL,
        public_key BYTEA NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS messages (
        id SERIAL PRIMARY KEY,
        sender_id INTEGER REFERENCES users(id),
        receiver_id INTEGER REFERENCES users(id),
        content BYTEA NOT NULL,
        timestamp BIGINT NOT NULL,
        read BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id);
    CREATE INDEX IF NOT EXISTS idx_messages_receiver ON messages(receiver_id);
    CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
    `
)