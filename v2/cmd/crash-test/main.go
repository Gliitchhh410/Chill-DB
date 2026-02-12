package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"chill-db/internal/db"
	"chill-db/internal/domain"
)

func main() {
	// Check arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [crash|recover]")
		return
	}
	mode := os.Args[1]

	storageDir := "./data_crash"

	if mode == "crash" {
		fmt.Println("💥 STARTING CRASH MODE")
		// Clean start
		os.RemoveAll(storageDir)

		repo, _ := db.NewLSMRepository(storageDir)

		// 1. Insert Data
		row := domain.Row{"999", "SecretAgent", "topsecret@cia.gov"}
		err := repo.InsertRow("users", row)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("📝 Data written to WAL.")
		fmt.Println("⚠️  Simulating Power Failure in 3... 2... 1...")
		time.Sleep(1 * time.Second)

		// KILL THE PROCESS
		os.Exit(1)

	} else if mode == "recover" {
		fmt.Println("🚑 STARTING RECOVERY MODE")

		// 1. Initialize (Should trigger WAL Replay)
		repo, err := db.NewLSMRepository(storageDir)
		if err != nil {
			log.Fatalf("Failed to restart DB: %v", err)
		}

		// 2. Search for the lost key
		fmt.Println("🔍 Searching for 'SecretAgent'...")
		rows, _ := repo.Query("users", "999")

		if len(rows) > 0 {
			fmt.Printf("✅ FOUND: %v\n", rows[0])
			fmt.Println("🎉 RECOVERY SUCCESSFUL! The WAL saved the data.")
		} else {
			fmt.Println("❌ DATA LOST! WAL recovery failed.")
			os.Exit(1)
		}
	}
}
