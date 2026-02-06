package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)


type FileRepository struct {
	DataDir string
	mu sync.RWMutex // many reads, one write
}


func NewFileRepository(dir string) (*FileRepository, error){ //initializes the repo once, similar to a singleton pattern, this returns a pointer the the fileRepo address
	cleanPath := filepath.Clean(dir) // provides a function file path e.g. ../myproject//data/ -> ../myproject/data

	err := os.MkdirAll(cleanPath, 0755) // MkdirAll creates a directory only if it doesn't exist, 0755 means Owner: rwx while Group and other can only rx

	if err != nil {
		return nil, fmt.Errorf("failed to create data root: %w", err)
	}

	return &FileRepository{DataDir: cleanPath}, nil
}


func (r *FileRepository) resolvePath(segments ...string) (string, error){ // a private function of the FileRepository struct, takes zero or more string parameters
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

	var dbs[] string

	for _, entry := range entries{
		if entry.IsDir(){
			dbs = append(dbs, entry.Name())
		}
	}
	return dbs, nil
}


func (r *FileRepository) CreateDatabase(ctx context.Context, name string)  error {
	r.mu.Lock() // Write lock
	defer r.mu.Unlock()

	dbPath, err := r.resolvePath(name)

	if err != nil {
		return err
	}

	err = os.Mkdir(dbPath, 0755)

	if err != nil {
		if os.IsExist(err){
			return fmt.Errorf("database '%s' already exists", name)
		}
		return fmt.Errorf("fs error: %w", err)
	}
	return nil
}


func 

