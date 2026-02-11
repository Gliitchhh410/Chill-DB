package main

import (
	"fmt"
	"log"
	"os"

	"chill-db/internal/db"
	"chill-db/internal/domain"
)

func main() {
	fmt.Println("=== Starting LSM Tree Test ===")

	// 1. Initialize the Repo (creates ./data directory)
	repo, err := db.NewLSMRepository("./data_test")
	if err != nil {
		log.Fatalf("Failed to init repo: %v", err)
	}

	// 2. Insert Data (Writes to WAL + MemTable)
	fmt.Println(">> Inserting 5 rows...")
	for i := 0; i < 5; i++ {
		// Create a dummy row: [ID, Name, Email]
		id := fmt.Sprintf("%d", i)
		name := fmt.Sprintf("User%d", i)
		email := fmt.Sprintf("user%d@chill-db.com", i)

		row := domain.Row{id, name, email}

		// Insert
		if err := repo.InsertRow("users", row); err != nil {
			log.Fatalf("Insert failed: %v", err)
		}
		fmt.Printf("   Inserted Key: users:%s\n", id)
	}

	// 3. Verify WAL exists
	if _, err := os.Stat("./data_test/wal.log"); err == nil {
		fmt.Println("✅ WAL file created successfully.")
	} else {
		log.Fatalf("❌ WAL file missing!")
	}

	// 4. Force Flush (Writes to SSTable)
	fmt.Println(">> Flushing MemTable to SSTable...")
	if err := repo.Flush(); err != nil {
		log.Fatalf("Flush failed: %v", err)
	}

	// 5. Verify SSTable exists
	files, _ := os.ReadDir("./data_test")
	sstCount := 0
	for _, f := range files {
		if f.Name() != "wal.log" {
			fmt.Printf("✅ Found SSTable: %s\n", f.Name())
			sstCount++
		}
	}

	if sstCount > 0 {
		fmt.Println("=== SUCCESS: Data persisted to disk! ===")
	} else {
		log.Fatalf("❌ No SSTable found after flush!")
	}
}
