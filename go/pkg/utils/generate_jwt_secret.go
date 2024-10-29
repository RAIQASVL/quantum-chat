package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

func generateSecureSecret(length int) (string, error) {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}

func main() {
    secret, err := generateSecureSecret(32) // 32 bytes = 256 bits
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Generated JWT Secret: %s\n", secret)
}