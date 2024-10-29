package handlers

import (
	"log"
	"net/http"
	"quantum-chat/internal/config"
	"quantum-chat/internal/middleware"
	"quantum-chat/internal/repository"
)

type Handlers struct {
    db     *repository.Database
    config *config.Config
    hub    *Hub
}

func NewHandlers(db *repository.Database, config *config.Config) *Handlers {
    h := &Handlers{
        db:     db,
        config: config,
    }
    h.hub = NewHub(h)
    go h.hub.Run()
    return h
}

func (h *Handlers) SetupRoutes(mux *http.ServeMux) {
    // Logging middleware
    withLogging := func(handler http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
            handler(w, r)
        }
    }

    // Combined auth and logging middleware
    withAuthAndLogging := func(handler http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
            
            // Apply auth middleware
            authMiddleware := middleware.AuthMiddleware(h.config.JWTSecret)
            authMiddleware(http.HandlerFunc(handler)).ServeHTTP(w, r)
        }
    }

    // Public routes (no auth required)
    mux.HandleFunc("/health", withLogging(h.handleHealth))
    mux.HandleFunc("/api/auth/register", withLogging(h.handleRegister))
    mux.HandleFunc("/api/auth/login", withLogging(h.handleLogin))
    mux.HandleFunc("/api/auth/refresh", withLogging(h.handleRefreshToken))

    // Protected routes (auth required)
    mux.HandleFunc("/api/auth/logout", withAuthAndLogging(h.handleLogout))
    
    // WebSocket route (auth required)
    mux.HandleFunc("/ws", withAuthAndLogging(h.handleWebSocket))
}

func (h *Handlers) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}