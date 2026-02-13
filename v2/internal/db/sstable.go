package db

import (
	"encoding/binary"
	"io"
	"os"
	"sort"
)

type SSTable struct {
	Filename string
	Filter   *BloomFilter
	// BloomFilter *BloomFilter // Future upgrade!
}

func (sst *SSTable) LoadFilter() error {
	f, err := os.Open(sst.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Seek to the end to find the Filter Length
	// The file structure is: [Data] [FilterLen(8 bytes)] [FilterData]
	stat, _ := f.Stat()
	fileSize := stat.Size()

	if fileSize < 8 {
		return nil // File too small to have a filter
	}
	// Seek to "End - 8 bytes"
	// 0 = Origin (Start), 1 = Current, 2 = End
	// Read Footer (Last 8 bytes) = Filter Length
	if _, err := f.Seek(-8, 2); err != nil {
		return err
	}

	var filterLen uint64
	if err := binary.Read(f, binary.LittleEndian, &filterLen); err != nil {
		return err
	}

	// Calculate Position of Filter Data
	// File: [ ... Data ... ] [FilterData (len)] [Footer (8)]
	// Position = FileSize - 8 - filterLen
	offset := fileSize - 8 - int64(filterLen)
	if offset < 0 {
		return nil
	}

	// Read Filter Data
	if _, err := f.Seek(offset, 0); err != nil {
		return err
	}

	filterData := make([]byte, filterLen)
	if _, err := f.Read(filterData); err != nil {
		return err
	}

	sst.Filter = DecodeBloomFilter(filterData)
	return nil
}
func (sst *SSTable) Search(searchkey string) ([]byte, bool, error) {
	if sst.Filter != nil {
		if !sst.Filter.Contains([]byte(searchkey)) {
			return nil, false, nil
		}
	}
	f, err := os.Open(sst.Filename)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()


	stat, _ := f.Stat()
	fileSize := stat.Size()
	dataEnd := fileSize
	if fileSize > 8 {
		f.Seek(-8, io.SeekEnd)
		var filterLen uint64
		if err := binary.Read(f, binary.LittleEndian, &filterLen); err == nil {
			dataEnd = fileSize - 8 - int64(filterLen)
		}
		f.Seek(0, io.SeekStart) // Rewind to start after checking footer
	}


	for {

		currentPos, _ := f.Seek(0, io.SeekCurrent)
		if currentPos >= dataEnd {
			break
		}

		var keyLen int32
		var valLen int32

		err := binary.Read(f, binary.LittleEndian, &keyLen)
		if err == io.EOF {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}

		err = binary.Read(f, binary.LittleEndian, &valLen)
		if err != nil {
			return nil, false, err
		}

		if int64(keyLen) + currentPos > dataEnd {
			break
		}
		
		keyBytes := make([]byte, int(keyLen))
		_, err = io.ReadFull(f, keyBytes)
		if err != nil {
			return nil, false, err
		}

		if string(keyBytes) == searchkey {
			valBytes := make([]byte, int(valLen))
			_, err = io.ReadFull(f, valBytes)
			if err != nil {
				return nil, false, err
			}
			return valBytes, true, nil
		} else {
			// NO MATCH - SKIP THE VALUE!
			// We seek forward by valLen bytes from the current position
			_, err = f.Seek(int64(valLen), io.SeekCurrent)
			if err != nil {
				return nil, false, err
			}
		}
	}
	return nil, false, nil
}


func (sst *SSTable) Scan() (map[string][]byte, error) {
	f, err := os.Open(sst.Filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 1. Calculate where the Data actually ends
	stat, _ := f.Stat()
	fileSize := stat.Size()
	dataEnd := fileSize // Default: Read whole file

	// Check if there is a footer (Filter Length)
	if fileSize > 8 {
		// Read last 8 bytes
		if _, err := f.Seek(-8, 2); err == nil {
			var filterLen uint64
			if err := binary.Read(f, binary.LittleEndian, &filterLen); err == nil {
				// Sanity check: Filter + Footer shouldn't be larger than the file
				if int64(filterLen+8) < fileSize {
					// FOUND IT! The data stops before the filter.
					dataEnd = fileSize - 8 - int64(filterLen)
				}
			}
		}
	}

	// 2. Rewind to start
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}

	data := make(map[string][]byte)

	// 3. Read Loop with strict limit
	for {
		// STOP CHECK: Have we reached the Bloom Filter?
		currentPos, _ := f.Seek(0, 1)
		if currentPos >= dataEnd {
			break
		}

		var keyLen int32
		if err := binary.Read(f, binary.LittleEndian, &keyLen); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		// Double check: Did reading keyLen push us past the limit?
		// (This handles corrupted files nicely)
		if int64(keyLen)+currentPos > dataEnd {
			break
		}

		var valLen int32
		if err := binary.Read(f, binary.LittleEndian, &valLen); err != nil {
			return nil, err
		}

		keyBytes := make([]byte, int(keyLen))
		valBytes := make([]byte, int(valLen))

		if _, err := io.ReadFull(f, keyBytes); err != nil {
			return nil, err
		}
		if _, err := io.ReadFull(f, valBytes); err != nil {
			return nil, err
		}

		data[string(keyBytes)] = valBytes
	}

	return data, nil
}

func WriteSSTable(data map[string][]byte, filename string) (*SSTable, error) {
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

	// Rule of thumb: Size = NumKeys * 10 (gives ~1% false positive rate)
	// HashCount = 7 (optimal for size*10)
	bf := NewBloomFilter(uint64(len(keys)*10), 7)
	for _, k := range keys {
		val := data[k]
		bf.Add(k)
		// Write Key Len (int32)
		if err := binary.Write(f, binary.LittleEndian, int32(len(k))); err != nil {
			return nil, err
		}
		// Write Val Len (int32)
		if err := binary.Write(f, binary.LittleEndian, int32(len(val))); err != nil {
			return nil, err
		}
		// Write Key Bytes
		if _, err := f.WriteString(k); err != nil {
			return nil, err
		}
		// Write Val Bytes
		if _, err := f.Write(val); err != nil {
			return nil, err
		}

	}
	// Write Bloom Filter to the END of the file
	bfData := bf.Encode()
	// Write the length of the filter first
	if _, err := f.Write(bfData); err != nil {
		return nil, err
	}

	// Write the filter itself
	if err := binary.Write(f, binary.LittleEndian, uint64(len(bfData))); err != nil {
		return nil, err
	}

	if err := f.Sync(); err != nil {
		return nil, err
	}

	return &SSTable{Filename: filename, Filter: bf}, nil

}
