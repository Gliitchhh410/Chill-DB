package db

import (
	"fmt"
	"testing"
	"time"
	"chill-db/internal/domain"
)

// Run with: go test -v -bench=. -benchmem

func BenchmarkInsert(b *testing.B) {
	dir := b.TempDir()
	repo, _ := NewLSMRepository(dir)
	defer repo.Close()

	baseRow := domain.Row{"key", "BenchUser", "bench@test.com"}
	
    // Approx size of one insert for throughput calc (key + user + email)
	rowSize := int64(len("key-100000") + len("BenchUser") + len("bench@test.com"))

	b.ResetTimer()
	start := time.Now()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		row := baseRow
		row[0] = key
		if err := repo.InsertRow("users", row); err != nil {
			b.Fatal(err)
		}
	}

	elapsed := time.Since(start)
	
	// 1. Report Throughput (MB/s)
	b.SetBytes(rowSize) 
	
	// 2. Report custom "Ops per second" for clarity
	opsPerSec := float64(b.N) / elapsed.Seconds()
	b.ReportMetric(opsPerSec, "ops/sec")
}

func BenchmarkQuery(b *testing.B) {
	dir := b.TempDir()
	repo, _ := NewLSMRepository(dir)
	defer repo.Close()

	// Seed data (1000 items)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("%d", i)
		_ = repo.InsertRow("users", domain.Row{key, "User", "email"})
	}
	repo.Flush()

	b.ResetTimer()
	start := time.Now()

	for i := 0; i < b.N; i++ {
		// Use a smaller key range to ensure hits, or larger to force misses
		key := fmt.Sprintf("%d", i%1000) 
		_, _ = repo.Query("users", key)
	}

	elapsed := time.Since(start)
	opsPerSec := float64(b.N) / elapsed.Seconds()
	b.ReportMetric(opsPerSec, "reads/sec")
}

func BenchmarkCompaction(b *testing.B) {
	// Compact is slow/heavy. We don't want to run it millions of times.
	// We reset the timer inside the loop, so we measure "One Compaction" perfectly.

	for i := 0; i < b.N; i++ {
		b.StopTimer() // --- PAUSE ---

		// 1. SETUP: Create a fresh DB for every iteration
		dir := b.TempDir()
		repo, _ := NewLSMRepository(dir)

		// 2. FRAGMENTATION: Create 5 files with 1000 keys each
		const numFiles = 5
		const keysPerFile = 1000
		
		for f := 0; f < numFiles; f++ {
			for k := 0; k < keysPerFile; k++ {
				key := fmt.Sprintf("key-%d", k)
				// Create versioned values to ensure they differ
				val := fmt.Sprintf("value-v%d", f) 
				repo.InsertRow("users", domain.Row{key, val, "email"})
			}
			repo.Flush() // Force new SSTable
		}

		b.StartTimer() // --- RESUME ---
		
		// 3. ACTION: Measure Compact()
		if err := repo.Compact(); err != nil {
			b.Fatal(err)
		}
		
		b.StopTimer() // --- PAUSE ---

		// 4. CLEANUP
		repo.Close()
	}
	
	// Calculate total items compacted per op
	totalItems := 5 * 1000
	b.ReportMetric(float64(totalItems), "items/op")
}