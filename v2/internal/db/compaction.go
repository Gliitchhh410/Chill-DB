package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func (r *LSMRepository) RunCompaction() {
	fmt.Println(" Compaction Started....")

	files, _ := os.ReadDir(r.storageDir)

	var sstFiles []string

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".db" {
			sstFiles = append(sstFiles, filepath.Join(r.storageDir, f.Name()))
		}
	}

	if len(sstFiles) <= 1 {
		fmt.Println("Nothing to compact.")
		return
	}

	sort.Strings(sstFiles)

	mergedData := make(map[string][]byte)

	for _, filename := range sstFiles {
		sst := &SSTable{Filename: filename}
		fileData, err := sst.Scan()
		if err != nil {
			fmt.Printf("❌ Compaction failed reading %s: %v\n", filename, err)
			return
		}
		// Merge logic: Simple map assignment.
		// Since we iterate Old->New, newer values automatically replace older ones.
		for k, v := range fileData {
			mergedData[k] = v
		}
	}

	// Writing the new Compacted SSTable
	newFileName := fmt.Sprintf("%s/sst_%d.db", r.storageDir, time.Now().UnixNano())
	if _, err := WriteSSTable(mergedData, newFileName); err != nil {
		fmt.Printf("❌ Compaction failed writing %s: %v\n", newFileName, err)
		return
	}

	for _, filename := range sstFiles {
		os.Remove(filename)
	}
	fmt.Printf("✅ Compaction complete! Merged %d files into %s\n", len(sstFiles), newFileName)
}
