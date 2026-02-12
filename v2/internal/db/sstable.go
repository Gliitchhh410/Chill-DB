package db

import (
	"encoding/binary"
	"io"
	"os"
	"sort"
)

type SSTable struct {
    Filename string
    // BloomFilter *BloomFilter // Future upgrade!
}


func (sst *SSTable) Search(searchkey string) ([]byte, bool, error) {
	f, err := os.Open(sst.Filename)
    if err != nil {
        return nil, false, err
    }
    defer f.Close()

	for {
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

		keyBytes := make([]byte, int(keyLen))
		_, err = io.ReadFull(f, keyBytes)
		if err != nil {
			return nil, false, err
		}

		if string(keyBytes) == searchkey {
			valBytes := make([]byte, int (valLen))
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
}


func (sst *SSTable) Scan() (map[string][]byte, error) {
	data := make(map[string][]byte)

	f, err := os.Open(sst.Filename)
    if err != nil {
        return nil, err
    }
	defer f.Close()

	for {
		var keyLen int32
		var valLen int32

		if err := binary.Read(f, binary.LittleEndian, &keyLen); err == io.EOF {
			break // End of file, we are done
		} else if err != nil {
			return nil, err
		}
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
