package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"quantum-chat/internal/middleware"
	"quantum-chat/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
    Username string `json:"username" validate:"required"`
    Password string `json:"password" validate:"required"`
}

type loginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    UserID       int64  `json:"user_id"`
    Username     string `json:"username"`
    ExpiresIn    int64  `json:"expires_in"`
}

type registerRequest struct {
    Username  string `json:"username" validate:"required,min=3,max=50"`
    Password  string `json:"password" validate:"required,min=6"`
    PublicKey []byte `json:"public_key" validate:"required"`
}

func (h *Handlers) handleLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req loginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Get user from database
    user, err := h.db.GetUser(req.Username)
    if err != nil {
        log.Printf("Error getting user: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    if user == nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Generate token pair
    accessToken, refreshToken, err := middleware.GenerateTokenPair(user.ID, h.config.JWTSecret)
    if err != nil {
        log.Printf("Error generating tokens: %v", err)
        http.Error(w, "Error generating tokens", http.StatusInternalServerError)
        return
    }

    // Send response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(loginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        UserID:       user.ID,
        Username:     user.Username,
        ExpiresIn:    time.Now().Add(middleware.AccessExpiry).Unix(),
    })
}

func (h *Handlers) handleRegister(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req registerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := validateRegisterRequest(req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Check if username exists
    existingUser, err := h.db.GetUser(req.Username)
    if err != nil {
        log.Printf("Error checking existing user: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    if existingUser != nil {
        http.Error(w, "Username already exists", http.StatusConflict)
        return
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        log.Printf("Error hashing password: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Create user
    user := &models.User{
        Username:  req.Username,
        Password:  string(hashedPassword),
        PublicKey: req.PublicKey,
    }

    if err := h.db.CreateUser(user); err != nil {
        log.Printf("Error creating user: %v", err)
        http.Error(w, "Error creating user", http.StatusInternalServerError)
        return
    }

    // Generate tokens
    accessToken, refreshToken, err := middleware.GenerateTokenPair(user.ID, h.config.JWTSecret)
    if err != nil {
        log.Printf("Error generating tokens: %v", err)
        http.Error(w, "Error generating tokens", http.StatusInternalServerError)
        return
    }

    // Send response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(loginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        UserID:       user.ID,
        Username:     user.Username,
        ExpiresIn:    time.Now().Add(middleware.AccessExpiry).Unix(),
    })
}

func (h *Handlers) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    token, err := middleware.TokenFromHeader(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    // Generate new token pair
    newAccess, newRefresh, err := middleware.RefreshTokenPair(token, h.config.JWTSecret)
    if err != nil {
        switch err {
        case middleware.ErrInvalidToken, middleware.ErrExpiredToken, middleware.ErrInvalidType:
            http.Error(w, err.Error(), http.StatusUnauthorized)
        default:
            log.Printf("Error refreshing tokens: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    // Send response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(loginResponse{
        AccessToken:  newAccess,
        RefreshToken: newRefresh,
        ExpiresIn:    time.Now().Add(middleware.AccessExpiry).Unix(),
    })
}

func validateRegisterRequest(req registerRequest) error {
    if strings.TrimSpace(req.Username) == "" {
        return errors.New("username is required")
    }
    if len(req.Username) < 3 {
        return errors.New("username must be at least 3 characters")
    }
    if len(req.Username) > 50 {
        return errors.New("username must not exceed 50 characters")
    }
    if strings.TrimSpace(req.Password) == "" {
        return errors.New("password is required")
    }
    if len(req.Password) < 6 {
        return errors.New("password must be at least 6 characters")
    }
    if len(req.PublicKey) == 0 {
        return errors.New("public key is required")
    }
    return nil
}

func (h *Handlers) handleLogout(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Get token from header
    token, err := middleware.TokenFromHeader(r)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    // Get user ID from context
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Validate token
    claims, err := middleware.ValidateToken(token, h.config.JWTSecret, middleware.AccessToken)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    // Revoke the token
    middleware.RevokeToken(token, time.Unix(claims.ExpiresAt, 0))

    // Close any active WebSocket connections for this user
    if client, ok := h.hub.clients[userID]; ok {
        client.Conn.Close()
        delete(h.hub.clients, userID)
    }

    // Return success response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Successfully logged out",
    })
}