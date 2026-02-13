package db

import (
	"chill-db/internal/domain"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type LSMRepository struct {
	memTable   *MemTable
	wal        *WAL
	storageDir string
	sstables   []*SSTable   // Cache of active SSTable filenames (sorted Newest -> Oldest)
	mu         sync.RWMutex // Protects sstables slice
}

func NewLSMRepository(storageDir string) (*LSMRepository, error) {
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		os.Mkdir(storageDir, 0755) // Use MkdirAll just in case
	}

	repo := &LSMRepository{
		memTable:   NewMemTable(),
		storageDir: storageDir,
		sstables:   []*SSTable{},
	}

	walPath := storageDir + "/wal.log"
	if _, err := os.Stat(walPath); err == nil {
		fmt.Println("FYI: Found existing WAL. Attempting recovery...")
		if err := repo.recoverFromWAL(walPath); err != nil {
			return nil, fmt.Errorf("WAL recovery failed: %w", err)
		}
	}

	wal, err := NewWAL(storageDir + "/wal.log")
	if err != nil {
		return nil, err
	}
	repo.wal = wal
	files, err := os.ReadDir(storageDir)
	if err != nil {
		return nil, err
	}
	var loadedSSTs []*SSTable
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".db" {
			fullPath := filepath.Join(storageDir, f.Name())
			sst := &SSTable{Filename: fullPath}
			_ = sst.LoadFilter()
			loadedSSTs = append(loadedSSTs, sst)
		}
	}

	sort.Slice(loadedSSTs, func(i, j int) bool {
		return loadedSSTs[i].Filename > loadedSSTs[j].Filename
	})

	repo.sstables = loadedSSTs
	return repo, nil
}
func (r *LSMRepository) Close() error {
	return r.wal.Close()
}
func (r *LSMRepository) recoverFromWAL(walPath string) error {

	f, err := os.Open(walPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var loadedCount int

	for {
		var keyLen int32
		var valLen int32

		// Read Key Length
		if err := binary.Read(f, binary.LittleEndian, &keyLen); err == io.EOF {
			break // Finished reading WAL
		} else if err != nil {
			return err
		}

		// Read Value Length
		if err := binary.Read(f, binary.LittleEndian, &valLen); err != nil {
			return err
		}

		//Read Data
		key := make([]byte, keyLen)
		val := make([]byte, valLen)
		if _, err := io.ReadFull(f, key); err != nil {
			return err
		}
		if _, err := io.ReadFull(f, val); err != nil {
			return err
		}

		//Put directly into MemTable
		// We use the map directly to avoid writing to the WAL again
		r.memTable.Put(string(key), val)
		loadedCount++
	}

	if loadedCount > 0 {
		fmt.Printf("🔄 Recovered %d records from WAL.\n", loadedCount)
	}
	return nil
}

func (r *LSMRepository) InsertRow(tableName string, row domain.Row) error {
	jsonRow, err := json.Marshal(row)
	if err != nil {
		return err
	}

	// Using the first value as the primary key for the key generation
	key := fmt.Sprintf("%s:%v", tableName, row[0])

	if err := r.wal.Append(key, jsonRow); err != nil {
		return err
	}
	r.memTable.Put(key, jsonRow)
	return nil
}

func (r *LSMRepository) Flush() error {
	r.memTable.mu.Lock()
	defer r.memTable.mu.Unlock()

	if r.memTable.size == 0 {
		return nil
	}

	filename := fmt.Sprintf("%s/sst_%d.db", r.storageDir, time.Now().UnixNano())
	newSST, err := WriteSSTable(r.memTable.data, filename)
	if err != nil {
		return err
	}
	if newSST.Filter == nil {
		fmt.Printf("⚠️ WriteSSTable didn't return a filter for %s. Attempting to load...\n", filename)

		// Force load and CHECK THE ERROR
		if err := newSST.LoadFilter(); err != nil {
			fmt.Printf("❌ LoadFilter FAILED: %v\n", err)
			return fmt.Errorf("failed to load filter: %w", err)
		}

		if newSST.Filter == nil {
			fmt.Printf("❌ Filter is STILL nil after loading!\n")
		} else {
			fmt.Printf("✅ Filter loaded successfully from disk.\n")
		}
	}

	r.mu.Lock()

	// Prepend the new file (since it's the newest)
	r.sstables = append([]*SSTable{newSST}, r.sstables...)
	r.mu.Unlock()

	r.memTable.data = make(map[string][]byte)
	r.memTable.size = 0
	if err := r.wal.Truncate(); err != nil {
		return fmt.Errorf("failed to truncate WAL: %w", err)
	}
	return nil

}

func (r *LSMRepository) Query(tableName string, conditions string) ([]domain.Row, error) {
	key := fmt.Sprintf("%s:%s", tableName, conditions)
	// check reading from memtable
	r.memTable.mu.RLock()

	if val, ok := r.memTable.data[key]; ok {
		r.memTable.mu.RUnlock() // RUnlock here is not defer because if we stopped writing till we read from sstable and disk it will cose alot in terms of performance
		var row domain.Row
		if err := json.Unmarshal(val, &row); err != nil {
			return nil, err
		}
		return []domain.Row{row}, nil
	}
	r.memTable.mu.RUnlock()
	// check reading from sstable
	r.mu.RLock()
	activeFiles := make([]*SSTable, len(r.sstables))
	copy(activeFiles, r.sstables)
	r.mu.RUnlock()

	for _, sst := range activeFiles {
		if sst.Filter == nil {
			fmt.Printf("⚠️ WARNING: Filter is NIL for %s\n", sst.Filename)
		} else {
			// DEBUG: Check what the filter says
			isPresent := sst.Filter.Contains([]byte(key))
			// fmt.Printf("DEBUG: Filter check for %s in %s: %v\n", key, sst.Filename, isPresent)

			if !isPresent {
				continue // Optimization working!
			}
		}
		val, found, err := sst.Search(key)

		if err != nil {
			return nil, err
		}

		if found {
			var row domain.Row
			if err := json.Unmarshal(val, &row); err != nil {
				return nil, err
			}
			return []domain.Row{row}, nil
		}

	}
	return []domain.Row{}, nil
}

func (r *LSMRepository) CreateTable(t domain.TableMetaData) error { return nil }
func (r *LSMRepository) DropDatabase() error                      { return nil }
