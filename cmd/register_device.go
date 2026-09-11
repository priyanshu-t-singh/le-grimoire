package cmd

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"modernc.org/sqlite"
)

var registerDeviceCmd = &cobra.Command{
	Use:   "register-device",
	Short: "Register a new e-ink device",
	Long:  `Provision a new e-ink device by inserting its ID and hashed API key into the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		deviceID, _ := cmd.Flags().GetString("id")
		rawKey, _ := cmd.Flags().GetString("key")
		dbPath := filepath.Join(AppFlags.DataDir, "le-grimoire.db")

		if deviceID == "" || rawKey == "" {
			log.Fatalf("Usage: le-grimoire register-device --id <device_id> --key <api_key> ")
		}

		dbPath, err := expandPath(dbPath)
		if err != nil {
			log.Fatalf("Failed to expand db path: %v", err)
		}

		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			log.Fatalf("Failed to create db directory: %v", err)
		}

		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			log.Fatalf("Failed to open DB: %v", err)
		}
		defer db.Close()

		keyHash := hashDeviceKey(rawKey)

		query := `INSERT INTO devices (device_id, api_key_hash, created_at) VALUES (?, ?, CURRENT_TIMESTAMP);`
		_, err = db.Exec(query, deviceID, keyHash)
		if err != nil {
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) {
				// 2067 = SQLITE_CONSTRAINT_UNIQUE
				// 1555 = SQLITE_CONSTRAINT_PRIMARYKEY
				if sqliteErr.Code() == 2067 || sqliteErr.Code() == 1555 {
					log.Fatalf("Device ID %s already exists (unique constraint violation)", deviceID)
				}
			}
			log.Fatalf("Failed to insert device: %v", err)
		}

		fmt.Printf("Device successfully provisioned!\n")
		fmt.Printf("Device ID : %s\n", deviceID)
		fmt.Printf("API Key   : %s\n", rawKey)
	},
}

func hashDeviceKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func expandPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("datadir path should not be empty")
	}

	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Join the home directory with the rest of the path
	return filepath.Join(homeDir, path[1:]), nil
}

func init() {
	registerDeviceCmd.Flags().String("id", "", "Unique Device ID (e.g. esp32-4in2-01)")
	registerDeviceCmd.Flags().String("key", "", "Raw API key to be hardcoded in firmware")
	rootCmd.AddCommand(registerDeviceCmd)
}
