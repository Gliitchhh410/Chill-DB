package db

import (
	"chill-db/internal/domain"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type LSMRepository struct {
	memTable *MemTable
	wal      *WAL
	storageDir string
}


func NewLSMRepository(storageDir string) (*LSMRepository, error) {
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		os.Mkdir(storageDir, 0755) // Use MkdirAll just in case
	}

	wal, err := NewWAL(storageDir + "/wal.log")
	if err != nil {
		return nil, err
	}

	return &LSMRepository{
		memTable: NewMemTable(),
		wal:      wal,
		storageDir: storageDir,
	}, nil
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
	_,err := WriteSSTable(r.memTable.data, filename)
	if err != nil {
		return err
	}

	r.memTable.data = make(map[string][]byte)
	r.memTable.size = 0
	if err := r.wal.Truncate(); err != nil {
		return fmt.Errorf("failed to truncate WAL: %w", err)
	}
	return nil
	
}


func (r *LSMRepository) CreateTable(t domain.TableMetaData) error { return nil }
func (r *LSMRepository) DropDatabase() error { return nil }
func (r *LSMRepository) Query(tableName string, conditions string) ([]domain.Row, error) { return nil, nil }