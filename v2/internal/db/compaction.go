package db

import (
	"fmt"
	"os"
	"sync"
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

	// results[0] will hold the map from oldFiles[0], etc.
	results := make([]map[string][]byte, len(oldFiles))
	// Buffered channel to prevent goroutines from blocking if multiple errors occur
	errChan := make(chan error, len(oldFiles))

	var wg sync.WaitGroup
	for i, filename := range oldFiles {
		wg.Add(1)
		// We pass 'i' and 'filename' to capture them by value
		go func(index int, fname string) {
			defer wg.Done()
			// Create a temporary SSTable struct
			sst := &SSTable{Filename: fname}
			// Call Scan() to get the data
			data, err := sst.Scan()
			if err != nil {
				errChan <- fmt.Errorf("failed to scan %s: %w", fname, err)
				return
			}
			// Store data in the specific index for this file
			results[index] = data
		}(i, filename)
	}

	wg.Wait()
	close(errChan)
	if len(errChan) > 0 {
		return <-errChan
	}

	// Iterate BACKWARDS (Oldest File -> Newest File)
	// This ensures that if a key exists in both 'old' and 'new' files,
	// the value from the 'new' file (processed later) overwrites the 'old' one.
	mergedData := make(map[string][]byte)
	for i := len(results) - 1; i >= 0; i-- {
		fileData := results[i]
		if fileData == nil {
			continue
		}
		// Loop through fileData and put every k,v into mergedData
		for k, v := range fileData {
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

	// Calculate how many NEW files arrived during the compaction

	newFilesCount := len(r.sstables) - len(oldFiles)
	// Build the new list
	newSSTables := make([]string, 0)
	if newFilesCount > 0 {
		// Keep the new files!
		newSSTables = append(newSSTables, newFilename)
	}
	// Append our new compacted file
	newSSTables = append(newSSTables, r.sstables[len(oldFiles):]...)
	// Update the pointer
	r.sstables = newSSTables

	r.mu.Unlock()

	//Cleanup: Delete old files (Delayed for safety)
	go func() {
		time.Sleep(5 * time.Second)
		for _, f := range oldFiles {
			os.Remove(f)
		}
	}()

	return nil
}
