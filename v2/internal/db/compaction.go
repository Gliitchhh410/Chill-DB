package db

import (
	"fmt"
	"os"
	"time"
)

func (r *LSMRepository) StartCompactionWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := r.Compact(); err != nil {
				fmt.Println("Compaction error:", err)
			}
		}
	}()
}

func (r *LSMRepository) Compact() error {
	//Snapshot: Get list of files to compact
	r.mu.RLock()
	if len(r.sstables) <= 1 {
		r.mu.RUnlock()
		return nil // Nothing to do
	}
	// Copy the list so we can work without blocking queries
	oldFiles := make([]string, len(r.sstables))
	copy(oldFiles, r.sstables)
	r.mu.RUnlock()

	fmt.Printf("🧹 Compacting %d files...\n", len(oldFiles))

	// Merge: Read all files into a map
	// Iterate REVERSE (Oldest -> Newest) so newer keys overwrite older ones
	mergedData := make(map[string][]byte)
	for i := len(oldFiles) - 1; i >= 0; i-- {
		sst := &SSTable{Filename: oldFiles[i]}
		data, err := sst.Scan() // Ensure you added Scan() to sstable.go!
		if err != nil {
			return err
		}
		for k, v := range data {
			mergedData[k] = v
		}
	}

	//  Write: Create the new compacted file
	newFilename := fmt.Sprintf("%s/compacted_%d.db", r.storageDir, time.Now().UnixNano())
	if _, err := WriteSSTable(mergedData, newFilename); err != nil {
		return err
	}

	//Swap: Update the active list atomically
	r.mu.Lock()
	newFilesCount := len(r.sstables) - len(oldFiles)
	newSSTables := make([]string, 0)
	if newFilesCount > 0 {
        newSSTables = append(newSSTables, r.sstables[:newFilesCount]...)
    }
	newSSTables = append(newSSTables, newFilename)
	r.sstables = newSSTables
	r.mu.Unlock()

	//Cleanup: Delete old files (Delayed for safety)
	go func() {
		time.Sleep(10 * time.Second) // Wait for pending queries to finish
		for _, f := range oldFiles {
			os.Remove(f)
		}
		fmt.Println(" Old SSTables deleted.")
	}()

	return nil
}
