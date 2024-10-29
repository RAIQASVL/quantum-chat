package handlers

import (
	"log"
	"sync"
)

type Hub struct {
    clients    map[int64]*Client
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
    mutex      sync.RWMutex
    handlers   *Handlers
}

func NewHub(handlers *Handlers) *Hub {
    return &Hub{
        clients:    make(map[int64]*Client),
        broadcast:  make(chan []byte),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        mutex:      sync.RWMutex{},
        handlers:   handlers,
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mutex.Lock()
            h.clients[client.UserID] = client
            h.mutex.Unlock()
            log.Printf("Hub: Client registered: %d", client.UserID)
            
        case client := <-h.unregister:
            h.mutex.Lock()
            if _, ok := h.clients[client.UserID]; ok {
                delete(h.clients, client.UserID)
                close(client.Send)
                log.Printf("Hub: Client unregistered: %d", client.UserID)
            }
            h.mutex.Unlock()
            
        case message := <-h.broadcast:
            h.mutex.RLock()
            for _, client := range h.clients {
                select {
                case client.Send <- message:
                default:
                    close(client.Send)
                    delete(h.clients, client.UserID)
                    log.Printf("Hub: Client removed due to blocked channel: %d", client.UserID)
                }
            }
            h.mutex.RUnlock()
        }
    }
}