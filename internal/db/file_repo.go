package db

import (
	"chill-db/internal/domain"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileRepository struct {
	DataDir string
	mu      sync.RWMutex // many reads, one write
}

func NewFileRepository(dir string) (*FileRepository, error) { //initializes the repo once, similar to a singleton pattern, this returns a pointer the the fileRepo address
	cleanPath := filepath.Clean(dir) // provides a function file path e.g. ../myproject//data/ -> ../myproject/data

	err := os.MkdirAll(cleanPath, 0755) // MkdirAll creates a directory only if it doesn't exist, 0755 means Owner: rwx while Group and other can only rx

	if err != nil {
		return nil, fmt.Errorf("failed to create data root: %w", err)
	}

	return &FileRepository{DataDir: cleanPath}, nil
}

func (r *FileRepository) resolvePath(segments ...string) (string, error) { // a private function of the FileRepository struct, takes zero or more string parameters
	fullPath := filepath.Join(r.DataDir, filepath.Join(segments...))
	if !strings.HasPrefix(fullPath, r.DataDir) {
		return "", fmt.Errorf("security violation: invalid path")
	}
	return fullPath, nil
}

func (r *FileRepository) ListDatabases(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.Unlock()

	entries, err := os.ReadDir(r.DataDir)

	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	var dbs []string

	for _, entry := range entries {
		if entry.IsDir() {
			dbs = append(dbs, entry.Name())
		}
	}
	return dbs, nil
}

func (r *FileRepository) CreateDatabase(ctx context.Context, name string) error {
	r.mu.Lock() // Write lock
	defer r.mu.Unlock()

	dbPath, err := r.resolvePath(name)

	if err != nil {
		return err
	}

	err = os.Mkdir(dbPath, 0755)

	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("database '%s' already exists", name)
		}
		return fmt.Errorf("fs error: %w", err)
	}
	return nil
}

func (r *FileRepository) CreateTable(ctx context.Context, dbName string, table domain.TableMetaData) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Structure: ./data/{dbName}/{tableName}.meta
	metaPath, err := r.resolvePath(dbName, table.Name+".meta")
	if err != nil {
		return err
	}

	// Structure: ./data/{dbName}/{tableName}.data
	dataPath, err := r.resolvePath(dbName, table.Name+".data")
	if err != nil {
		return err
	}

	if _, err := os.Stat(metaPath); err == nil {
		return fmt.Errorf("table '%s' already exists in '%s'", table.Name, dbName)
	}

	metaFile, err := os.Create(metaPath)
	if err != nil {
		// If the DB directory doesn't exist, this fails with PathError
		return fmt.Errorf("failed to create table meta: %w", err)
	}

	defer metaFile.Close()

	// We write: Name, Type (e.g., "id,int")
	writer := csv.NewWriter(metaFile)
	for _, col := range table.Columns {
		if err := writer.Write([]string{col.Name, col.Type}); err != nil { // Writes a line like: "id,int" or "username,string"
			return err
		}
	}
	writer.Flush() // Pushes any buffered data to the file"disc"

	dataFile, err := os.Create(dataPath)
	if err != nil {
		return fmt.Errorf("failed to create table data: %w", err)
	}
	dataFile.Close() // Close immediately, we just needed to create it
	return nil
}

// resolvepath, os.OpenFile, csv.NewWriter
func (r *FileRepository) InsertRow(ctx context.Context, dbName, tableName string, row domain.Row) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	dataPath, err := r.resolvePath(dbName, tableName+".data")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(dataPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("table '%s' does not exist", tableName)
		}
		return fmt.Errorf("failed to open table: %w", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)

	if err := writer.Write(row); err != nil {
		return fmt.Errorf("failed to write row: %w", err)
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return err
	}
	return nil
}

func (r *FileRepository) Query(ctx context.Context, dbName, tableName string) ([]domain.Row, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dataPath, err := r.resolvePath(dbName, tableName+".data")
	if err != nil {
		return nil, err
	}

	file, err := os.Open(dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("table '%s' does not exist", tableName)
		}
		return nil, fmt.Errorf("failed to open table: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		// Empty file is not an error, just return empty list
		return []domain.Row{}, nil
	}

	var rows []domain.Row
	for _, record := range records {
		rows = append(rows, domain.Row(record))
	}

	return rows, nil

}
