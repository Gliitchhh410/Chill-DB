package db

import (
	"chill-db/internal/domain"
	"encoding/json"
	"fmt"
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
	sstables   []string     // Cache of active SSTable filenames (sorted Newest -> Oldest)
	mu         sync.RWMutex // Protects sstables slice
}

func NewLSMRepository(storageDir string) (*LSMRepository, error) {
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		os.Mkdir(storageDir, 0755) // Use MkdirAll just in case
	}

	wal, err := NewWAL(storageDir + "/wal.log")
	if err != nil {
		return nil, err
	}

	repo :=  &LSMRepository{
		memTable:   NewMemTable(),
		wal:        wal,
		storageDir: storageDir,
		sstables:   []string{},
	}
	files, err := os.ReadDir(storageDir)
	if err != nil {
		return nil, err
	}
	var ssts []string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".db" {
			ssts = append(ssts, filepath.Join(storageDir, f.Name()))
		}
	}

	sort.Slice(ssts, func(i, j int) bool {
		return ssts[i] > ssts[j]
	})

	repo.sstables = ssts
	fmt.Printf("Loaded %d SSTables from disk.\n", len(repo.sstables))

	return repo, nil
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
	_, err := WriteSSTable(r.memTable.data, filename)
	if err != nil {
		return err
	}


	r.mu.Lock()

	// Prepend the new file (since it's the newest)
    r.sstables = append([]string{filename}, r.sstables...)
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
	activeFiles := make([]string, len(r.sstables))
	copy(activeFiles, r.sstables)
	r.mu.RUnlock()

	for _, filename := range activeFiles {

		sst := &SSTable{Filename: filename}
		val, ok, err := sst.Search(key)

		if err != nil {
			return nil, err
		}

		if ok {
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
