package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"

	"modernc.org/sqlite"
	_ "modernc.org/sqlite"
)

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func main() {
	deviceID := flag.String("id", "", "Unique Device ID (e.g. esp32-4in2-01)")
	rawKey := flag.String("key", "", "Raw API key to be hardcoded in firmware")
	dbPath := flag.String("db", ".db/app.db", "Path to SQLite database")
	flag.Parse()

	if *deviceID == "" || *rawKey == "" {
		log.Fatalf("Usage: go run cmd/register-device/main.go -id <device_id> -key <api_key>")
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	keyHash := hashKey(*rawKey)

	query := `INSERT INTO devices (device_id, api_key_hash, created_at) VALUES (?, ?, CURRENT_TIMESTAMP);`
	_, err = db.Exec(query, *deviceID, keyHash)
	if err != nil {
		// Check if the error is due to a unique constraint violation
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) {
			// 2067 = SQLITE_CONSTRAINT_UNIQUE
			// 1555 = SQLITE_CONSTRAINT_PRIMARYKEY
			if sqliteErr.Code() == 2067 || sqliteErr.Code() == 1555 {
				log.Fatalf("Device ID %s already exists (unique constraint violation)", *deviceID)
			}
		}
		log.Fatalf("Failed to insert device: %v", err)
	}

	fmt.Printf("Device successfully provisioned!\n")
	fmt.Printf("Device ID : %s\n", *deviceID)
	fmt.Printf("API Key   : %s\n", *rawKey)
}
