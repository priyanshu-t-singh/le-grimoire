package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
)

type contextKey string

const DeviceIDKey contextKey = "device_id"

type AuthStore interface {
	GetDeviceAuthHash(deviceID string) (string, error)
}

func DeviceAuth(repo AuthStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			deviceID := r.Header.Get("X-Device-Id")
			authHeader := r.Header.Get("Authorization")

			// Ignore /health endpoint for authentication
			if r.URL.Path == "/health" || r.URL.Path == "/simulator" {
				next.ServeHTTP(w, r)
				return
			}

			if deviceID == "" || authHeader == "" {
				http.Error(w, "Missing authentication headers", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			apiKey := parts[1]
			expectedHash, err := repo.GetDeviceAuthHash(deviceID)
			if err != nil {
				log.Printf("Auth failed: device %s not found or DB error: %v", deviceID, err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			hash := sha256.Sum256([]byte(apiKey))
			incomingHash := hex.EncodeToString(hash[:])

			if subtle.ConstantTimeCompare([]byte(incomingHash), []byte(expectedHash)) != 1 {
				log.Printf("Auth failed: invalid API key for device %s", deviceID)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := SetDeviceID(r.Context(), deviceID)
			UpdateLoggingContext(r, ctx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SetDeviceID(ctx context.Context, deviceID string) context.Context {
	return context.WithValue(ctx, DeviceIDKey, deviceID)
}

func GetDeviceID(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(DeviceIDKey).(string)
	return val, ok
}
