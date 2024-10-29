// internal/middleware/auth.go
package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
)

type TokenType string

const (
    AccessToken  TokenType = "access"
    RefreshToken TokenType = "refresh"
)

type Claims struct {
    UserID    int64     `json:"user_id"`
    TokenType TokenType `json:"token_type"`
    jwt.StandardClaims
}

type ContextKey string

const (
    UserIDKey       ContextKey = "userID"
    AccessExpiry              = 15 * time.Minute
    RefreshExpiry            = 7 * 24 * time.Hour
)

var (
    ErrInvalidToken  = errors.New("invalid token")
    ErrExpiredToken  = errors.New("token has expired")
    ErrInvalidType   = errors.New("invalid token type")
    ErrTokenRevoked  = errors.New("token has been revoked")
)

// TokenBlacklist for managing revoked tokens
type TokenBlacklist struct {
    tokens map[string]time.Time
    mu     sync.RWMutex
}

var blacklist = &TokenBlacklist{
    tokens: make(map[string]time.Time),
}

func (bl *TokenBlacklist) Add(token string, expiry time.Time) {
    bl.mu.Lock()
    defer bl.mu.Unlock()
    bl.tokens[token] = expiry
    bl.cleanup()
}

func (bl *TokenBlacklist) IsBlacklisted(token string) bool {
    bl.mu.RLock()
    defer bl.mu.RUnlock()
    _, exists := bl.tokens[token]
    return exists
}

func (bl *TokenBlacklist) cleanup() {
    now := time.Now()
    for token, expiry := range bl.tokens {
        if now.After(expiry) {
            delete(bl.tokens, token)
        }
    }
}

// AuthMiddleware creates a new middleware handler for JWT authentication
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            log.Printf("Auth Middleware: Processing request for path: %s", r.URL.Path)

            // Get token from Authorization header
            tokenString, err := TokenFromHeader(r)
            if err != nil {
                log.Printf("Auth Middleware: Token extraction failed: %v", err)
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            log.Printf("Auth Middleware: Token extracted successfully")

            // Parse and validate token
            claims := &Claims{}
            token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    log.Printf("Auth Middleware: Unexpected signing method: %v", token.Header["alg"])
                    return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
                }
                return []byte(jwtSecret), nil
            })

            if err != nil {
                log.Printf("Auth Middleware: Token parsing failed: %v", err)
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            if !token.Valid {
                log.Printf("Auth Middleware: Token is invalid")
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            // Add user ID to context
            ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
            log.Printf("Auth Middleware: Added user ID %d to context", claims.UserID)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID int64, jwtSecret string) (accessToken, refreshToken string, err error) {
    // Generate access token
    accessToken, err = generateToken(userID, jwtSecret, AccessToken, AccessExpiry)
    if err != nil {
        return "", "", err
    }

    // Generate refresh token
    refreshToken, err = generateToken(userID, jwtSecret, RefreshToken, RefreshExpiry)
    if err != nil {
        return "", "", err
    }

    return accessToken, refreshToken, nil
}

func generateToken(userID int64, jwtSecret string, tokenType TokenType, expiry time.Duration) (string, error) {
    claims := &Claims{
        UserID:    userID,
        TokenType: tokenType,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(expiry).Unix(),
            IssuedAt:  time.Now().Unix(),
            Subject:   fmt.Sprintf("%d", userID),
            Issuer:    "quantum-chat",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signedToken, err := token.SignedString([]byte(jwtSecret))
    if err != nil {
        return "", fmt.Errorf("error signing token: %v", err)
    }

    return signedToken, nil
}

// ValidateToken validates a token string and checks its type
func ValidateToken(tokenStr string, jwtSecret string, expectedType TokenType) (*Claims, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(jwtSecret), nil
    })

    if err != nil {
        if err == jwt.ErrSignatureInvalid {
            return nil, ErrInvalidToken
        }
        return nil, err
    }

    if !token.Valid {
        return nil, ErrInvalidToken
    }

    if time.Unix(claims.ExpiresAt, 0).Before(time.Now()) {
        return nil, ErrExpiredToken
    }

    if claims.TokenType != expectedType {
        return nil, ErrInvalidType
    }

    return claims, nil
}

// RefreshTokenPair creates new access and refresh tokens using a valid refresh token
func RefreshTokenPair(refreshToken string, jwtSecret string) (string, string, error) {
    // Validate refresh token
    claims, err := ValidateToken(refreshToken, jwtSecret, RefreshToken)
    if err != nil {
        return "", "", err
    }

    // Generate new token pair
    accessToken, newRefreshToken, err := GenerateTokenPair(claims.UserID, jwtSecret)
    if err != nil {
        return "", "", err
    }

    // Revoke old refresh token
    blacklist.Add(refreshToken, time.Unix(claims.ExpiresAt, 0))

    return accessToken, newRefreshToken, nil
}

// TokenFromHeader extracts token from Authorization header
func TokenFromHeader(r *http.Request) (string, error) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        return "", errors.New("authorization header is required")
    }

    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return "", errors.New("authorization header format must be Bearer {token}")
    }

    return parts[1], nil
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
    userID, ok := ctx.Value(UserIDKey).(int64)
    return userID, ok
}

// RevokeToken adds a token to the blacklist
func RevokeToken(token string, expiry time.Time) {
    blacklist.Add(token, expiry)
}

// CreateAuthenticatedContext creates a new context with user ID
func CreateAuthenticatedContext(ctx context.Context, userID int64) context.Context {
    return context.WithValue(ctx, UserIDKey, userID)
}