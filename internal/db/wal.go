package db

import (
	"encoding/binary"
	"os"
	"sync"
)

// Before we touch the MemTable, we write the operation to a file on disk. 
// Speed: We only append to the end of the file. Appending is extremely fast (almost as fast as RAM) because the disk head doesn't have to jump around.
// Recovery: If we crash, we just read this file from top to bottom on restart to rebuild the MemTable.

type WAL struct {
	file *os.File
	mu sync.Mutex
}

func NewWAL(path string) (*WAL, error) {
	// O_APPEND: Always write to the end
	// O_CREATE: Create if it doesn't exist
	// O_WRONLY: We only write here (reading is for recovery startup)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &WAL{file: f}, nil
}
// 1. Lock the mutex (w.mu.Lock())
// 2. Defer unlock
// 3. Write Key Length (int32) using binary.Write
// 4. Write Value Length (int32) using binary.Write
// 5. Write the Key bytes
// 6. Write the Value bytes
// 7. Call w.file.Sync() to force the OS to save to disk immediately
// Hint: binary.Write(w.file, binary.LittleEndian, int32(len(key)))
func (w *WAL) Append(key string, value []byte) error{
	w.mu.Lock()
	defer w.mu.Unlock()

	err := binary.Write(w.file, binary.LittleEndian, int32(len(key))) // Why binary.LittleEndian? Computers store numbers in different ways (big-endian vs little-endian). We choose one standard so that if you move the file to a different computer, it can still be read.
	if err != nil {
		return err
	}
	err = binary.Write(w.file, binary.LittleEndian, int32(len(value))) // Why int32? We use a fixed size (4 bytes) for the length so the reader knows exactly how many bytes to read next.
	if err != nil {
		return err
	}
	_, err = w.file.Write([]byte(key))
	if err != nil {
		return err
	}
	_, err = w.file.Write(value)
	if err != nil {
		return err
	}
	w.file.Sync()	
	return nil
}

func (w *WAL) Close() error {
	return w.file.Close()
}


