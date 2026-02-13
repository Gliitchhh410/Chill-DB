package db

import (
	"fmt"
	"testing"
	"chill-db/internal/domain"
)
// to run D:\DS\Chill-DB\v2\internal\db> go test -v -bench='.' -run='^$'
func BenchmarkInsert(b *testing.B) {
	dir := b.TempDir()
	repo, _ := NewLSMRepository(dir)
    defer repo.Close() // <--- CRITICAL FIX

	baseRow := domain.Row{"key", "BenchUser", "bench@test.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		row := baseRow
		row[0] = key 
		_ = repo.InsertRow("users", row)
	}
}

func BenchmarkQuery(b *testing.B) {
	dir := b.TempDir()
	repo, _ := NewLSMRepository(dir)
    defer repo.Close() // <--- CRITICAL FIX

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("%d", i)
		_ = repo.InsertRow("users", domain.Row{key, "User", "email"})
	}
	repo.Flush() 

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("%d", i%1000)
		_, _ = repo.Query("users", key)
	}
}

func BenchmarkCompaction(b *testing.B) {
    // Note: b.N is the loop for the benchmark framework, 
    // but typically compaction is heavy so we might run it fewer times.
	for i := 0; i < b.N; i++ {
		b.StopTimer() // Pause setup time
		
		dir := b.TempDir()
		repo, _ := NewLSMRepository(dir)
		
		// Create fragmentation
		for f := 0; f < 5; f++ {
			for k := 0; k < 1000; k++ {
				key := fmt.Sprintf("key-%d", k)
				val := fmt.Sprintf("value-v%d", f)
				repo.InsertRow("users", domain.Row{key, val, "email"})
			}
			repo.Flush()
		}

		b.StartTimer() // Measure ONLY this part
		if err := repo.Compact(); err != nil {
			b.Fatal(err)
		}
        b.StopTimer() // Stop before cleanup

        repo.Close() // <--- CRITICAL FIX (Manual close inside loop)
	}
}