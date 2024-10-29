package repository

import (
	"database/sql"
	"fmt"
	"quantum-chat/internal/models"
	"time"

	_ "github.com/lib/pq"
)

type Database struct {
    db *sql.DB
}

func NewDatabase(url string) (*Database, error) {
    var db *sql.DB
    var err error
    
    maxRetries := 5
    for i := 0; i < maxRetries; i++ {
        db, err = sql.Open("postgres", url)
        if err == nil {
            if err = db.Ping(); err == nil {
                return &Database{db: db}, nil
            }
        }
        time.Sleep(time.Second * 2)
    }
    
    return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxRetries, err)
}

func (d *Database) Close() error {
    if d.db != nil {
        return d.db.Close()
    }
    return nil
}

// Add Ping method
func (d *Database) Ping() error {
    if d.db == nil {
        return fmt.Errorf("database connection is nil")
    }
    return d.db.Ping()
}

// User methods
func (d *Database) CreateUser(user *models.User) error {
    query := `
        INSERT INTO users (username, password, public_key)
        VALUES ($1, $2, $3)
        RETURNING id`
    
    return d.db.QueryRow(query, 
        user.Username, 
        user.Password, 
        user.PublicKey,
    ).Scan(&user.ID)
}

func (d *Database) GetUser(username string) (*models.User, error) {
    user := &models.User{}
    query := `
        SELECT id, username, password, public_key
        FROM users
        WHERE username = $1`
    
    err := d.db.QueryRow(query, username).Scan(
        &user.ID,
        &user.Username,
        &user.Password,
        &user.PublicKey,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return user, err
}

func (d *Database) GetUserByID(id int64) (*models.User, error) {
    user := &models.User{}
    query := `
        SELECT id, username, password, public_key
        FROM users
        WHERE id = $1`
    
    err := d.db.QueryRow(query, id).Scan(
        &user.ID,
        &user.Username,
        &user.Password,
        &user.PublicKey,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return user, err
}

// Message methods
func (d *Database) SaveMessage(msg *models.Message) error {
    query := `
        INSERT INTO messages (sender_id, receiver_id, content, timestamp, read)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`
    
    return d.db.QueryRow(query,
        msg.SenderID,
        msg.ReceiverID,
        msg.Content,
        msg.Timestamp,
        msg.Read,
    ).Scan(&msg.ID)
}

func (d *Database) GetMessages(userID int64, limit int) ([]*models.Message, error) {
    query := `
        SELECT id, sender_id, receiver_id, content, timestamp, read
        FROM messages
        WHERE sender_id = $1 OR receiver_id = $1
        ORDER BY timestamp DESC
        LIMIT $2`
    
    rows, err := d.db.Query(query, userID, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var messages []*models.Message
    for rows.Next() {
        msg := &models.Message{}
        err := rows.Scan(
            &msg.ID,
            &msg.SenderID,
            &msg.ReceiverID,
            &msg.Content,
            &msg.Timestamp,
            &msg.Read,
        )
        if err != nil {
            return nil, err
        }
        messages = append(messages, msg)
    }
    return messages, nil
}