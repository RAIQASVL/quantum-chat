// cmd/server/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"quantum-chat/internal/config"
	"quantum-chat/internal/handlers"
	"quantum-chat/internal/repository"
)

type Server struct {
    config     *config.Config
    httpServer *http.Server
    db         *repository.Database
    handlers   *handlers.Handlers
}

func NewServer(cfg *config.Config) *Server {
    return &Server{
        config: cfg,
    }
}

func (s *Server) Initialize() error {
    log.Println("Initializing server...")

    // Initialize database
    log.Println("Connecting to database...")
    db, err := repository.NewDatabase(s.config.DatabaseURL)
    if err != nil {
        log.Printf("Database connection error: %v", err)
        return err
    }
    log.Println("Database connected successfully")
    s.db = db

    // Initialize handlers
    log.Println("Initializing handlers...")
    s.handlers = handlers.NewHandlers(s.db, s.config)

    // Setup HTTP server
    log.Println("Setting up HTTP server...")
    mux := http.NewServeMux()
    s.handlers.SetupRoutes(mux)

    s.httpServer = &http.Server{
        Addr:         s.config.Port,
        Handler:      mux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    log.Println("Server initialization complete")
    return nil
}

func (s *Server) Start() error {
    log.Printf("Server starting on %s...", s.config.Port)
    return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := s.httpServer.Shutdown(ctx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }

    if s.db != nil {
        if err := s.db.Close(); err != nil {
            log.Printf("Database closure error: %v", err)
        }
    }
}

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    
    cfg := config.LoadConfig()
    
    server := NewServer(cfg)
    if err := server.Initialize(); err != nil {
        log.Fatal("Failed to initialize server:", err)
    }

    // Graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        if err := server.Start(); err != http.ErrServerClosed {
            log.Fatal("Server failed:", err)
        }
    }()

    <-stop
    log.Println("Shutting down server...")
    server.Shutdown()
}