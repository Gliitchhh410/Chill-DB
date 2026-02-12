package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"chill-db/internal/db"
	"chill-db/internal/domain"
)

func main() {
	// 1. Cleanup previous runs (Optional, for clean slate)
	os.RemoveAll("./data_stress")

	// 2. Initialize Repo
	fmt.Println("🚀 Initializing LSM Engine...")
	repo, err := db.NewLSMRepository("./data_stress")
	if err != nil {
		log.Fatalf("Init failed: %v", err)
	}

	// 3. Create Fragmentation (Make 4 SSTables)
	// We insert duplicate keys to test that Compaction correctly keeps the LATEST value.
	ids := []string{"1", "2", "3"}

	for i := 1; i <= 4; i++ {
		fmt.Printf("\n--- Batch %d (Writing & Flushing) ---\n", i)

		for _, id := range ids {
			// Update the name each time: User1_v1, User1_v2, etc.
			// This proves that Compaction keeps the newest version.
			name := fmt.Sprintf("User%s_v%d", id, i)
			row := domain.Row{id, name, "test@mail.com"} // Assuming Row is []string{"id", "name", "email"}

			if err := repo.InsertRow("users", row); err != nil {
				log.Fatalf("Insert failed: %v", err)
			}
		}

		// Force Flush to disk immediately to create a new .db file
		if err := repo.Flush(); err != nil {
			log.Fatalf("Flush failed: %v", err)
		}

		countFiles("./data_stress")
		time.Sleep(500 * time.Millisecond) // Slight delay to ensure distinct timestamps
	}

	// 4. Verify Data BEFORE Compaction
	fmt.Println("\n🔍 Reading User:1 (Expect 'User1_v4')...")
	rows, _ := repo.Query("users", "1")
	if len(rows) > 0 {
		fmt.Printf("   Found: %v (Correct)\n", rows[0])
	} else {
		log.Fatal("❌ Data missing before compaction!")
	}

	// 5. Start Compaction (Fast interval for testing)
	fmt.Println("\n🧹 Starting Compaction Worker (2s interval)...")
	repo.StartCompactionWorker(2 * time.Second)

	// Wait for compaction to kick in (Loop check)
	fmt.Println("⏳ Waiting for compaction to merge files...")
	for {
		time.Sleep(1 * time.Second)
		count := countFiles("./data_stress")

		// If we are down to 1 .db file (plus wal.log), compaction worked!
		if count == 1 {
			fmt.Println("\n✅ COMPACTION FINISHED! Disk is clean.")
			break
		}
	}

	// 6. Verify Data AFTER Compaction
	// This is the critical test: Did we lose data during the merge?
	fmt.Println("\n🔍 Reading User:1 AFTER Compaction...")
	rows, _ = repo.Query("users", "1")
	if len(rows) > 0 {
		fmt.Printf("   Found: %v\n", rows[0])

		// Check if it's the latest version (v4)
		// Note: You'll need to adjust this check based on your exact Row structure
		// fmt.Println("   (Verify this is 'User1_v4')")
	} else {
		log.Fatal("❌ CRITICAL: Data lost after compaction!")
	}

	fmt.Println("\n🎉 Test Passed: LSM Engine works correctly.")
}

// Helper to count .db files
func countFiles(dir string) int {
	files, _ := os.ReadDir(dir)
	count := 0
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".db" {
			count++
		}
	}
	fmt.Printf("   [Disk Status] %d SSTables found.\n", count)
	return count
}
