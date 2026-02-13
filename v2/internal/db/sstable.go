package db

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"os"
	"sort"
)

type IndexEntry struct {
	Key    string
	Offset int64
}

type SSTable struct {
	Filename string
	Filter   *BloomFilter
	Index    []IndexEntry
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
	if fileSize > 16 {
		f.Seek(-16, io.SeekEnd)
		var filterLen, indexLen uint64
		binary.Read(f, binary.LittleEndian, &filterLen)
		binary.Read(f, binary.LittleEndian, &indexLen)
		dataEnd = fileSize - 16 - int64(filterLen) - int64(indexLen)
	}

	var startOffset int64 = 0
	// Index Lookup: Jump to the closest offset
	if len(sst.Index) > 0 {
		idx := sort.Search(len(sst.Index), func(i int) bool {
			return sst.Index[i].Key > searchkey
		})
		if idx > 0 {
			startOffset = sst.Index[idx-1].Offset
		}
	}
	f.Seek(startOffset, io.SeekStart)

	for {
		currentPos, _ := f.Seek(0, io.SeekCurrent)
		if currentPos >= dataEnd {
			break
		}

		var keyLen, valLen int32
		if err := binary.Read(f, binary.LittleEndian, &keyLen); err != nil {
			break
		}
		if err := binary.Read(f, binary.LittleEndian, &valLen); err != nil {
			break
		}

		keyBytes := make([]byte, int(keyLen))
		if _, err := io.ReadFull(f, keyBytes); err != nil {
			break
		}
		keyStr := string(keyBytes)

		// Optimization: Since SSTable is sorted, if we pass the key, it's not here
		if keyStr > searchkey {
			return nil, false, nil
		}

		if keyStr == searchkey {
			valBytes := make([]byte, int(valLen))
			io.ReadFull(f, valBytes)
			return valBytes, true, nil
		}

		// Skip value to next record
		f.Seek(int64(valLen), io.SeekCurrent)
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

	var index []IndexEntry
	bf := NewBloomFilter(uint64(len(keys)*10), 7)
	currentOffset := int64(0)

	for i, k := range keys {
		val := data[k]

		// 1. Update Index: Every 100 keys, record the offset
		if i%100 == 0 {
			index = append(index, IndexEntry{Key: k, Offset: currentOffset})
		}

		// 2. Update Bloom Filter
		bf.Add(k)

		// 3. Write Data
		keyLen := int32(len(k))
		valLen := int32(len(val))

		binary.Write(f, binary.LittleEndian, keyLen)
		binary.Write(f, binary.LittleEndian, valLen)
		f.WriteString(k)
		f.Write(val)

		// 4. Track Offset: 4+4 bytes for lengths + actual data
		currentOffset += int64(8 + keyLen + valLen)
	}

	// Write Bloom Filter and Footer (Same as your logic)
	bfData := bf.Encode()
	indexData, _ := EncodeIndex(index)
	f.Write(bfData)
	f.Write(indexData)
	binary.Write(f, binary.LittleEndian, uint64(len(bfData)))
	binary.Write(f, binary.LittleEndian, uint64(len(indexData)))

	f.Sync()

	return &SSTable{
		Filename: filename,
		Filter:   bf,
		Index:    index, // Now properly populated!
	}, nil
}

func (sst *SSTable) LoadMetadata() error {
	f, err := os.Open(sst.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, _ := f.Stat()
	size := stat.Size()
	if size < 16 {
		return nil
	} // Too small for metadata

	// 1. Read the two lengths from the very end
	f.Seek(-16, io.SeekEnd)
	var filterLen, indexLen uint64
	binary.Read(f, binary.LittleEndian, &filterLen)
	binary.Read(f, binary.LittleEndian, &indexLen)

	// 2. Load Filter
	// Offset calculation: TotalSize - Footer(16) - IndexLen - FilterLen
	filterOffset := size - 16 - int64(indexLen) - int64(filterLen)
	f.Seek(filterOffset, io.SeekStart)
	filterData := make([]byte, filterLen)
	f.Read(filterData)
	sst.Filter = DecodeBloomFilter(filterData)

	// 3. Load Index
	// The index starts right after the filter
	indexData := make([]byte, indexLen)
	f.Read(indexData)
	sst.Index, err = DecodeIndex(indexData) // Use a simple decoder (like JSON or binary)
	if err != nil {
		return err
	}

	return nil
}

func EncodeIndex(index []IndexEntry) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(index)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeIndex(data []byte) ([]IndexEntry, error) {
	var index []IndexEntry
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&index)
	if err != nil {
		return nil, err
	}
	return index, nil
}
