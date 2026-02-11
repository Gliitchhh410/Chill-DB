package db

import (
	"encoding/binary"
	"os"
	"sort"
)

type SSTable struct {
    Filename string
    // BloomFilter *BloomFilter // Future upgrade!
}


func WriteSSTable(data map[string][]byte, filename string) (*SSTable, error){
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	f, err := os.Create(filename)
    if err != nil {
        return nil, err
    }
    defer f.Close()


	for _, k := range keys {
		val := data[k]
		// Write Key Len (int32)
		if err := binary.Write(f, binary.LittleEndian, int32(len(k))); err != nil { return nil, err }
		// Write Val Len (int32)
		if err := binary.Write(f, binary.LittleEndian, int32(len(val))); err != nil { return nil, err }		
		// Write Key Bytes
		if _, err := f.WriteString(k); err != nil { return nil, err }
		// Write Val Bytes
		if _, err := f.Write(val); err != nil { return nil, err }



		// Bloom Filter to be added here
	}

	return &SSTable{Filename: filename}, nil

}


