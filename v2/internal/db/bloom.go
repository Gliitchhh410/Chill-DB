package db

import (
	"encoding/binary"
	"hash/fnv"
)

type BloomFilter struct {
	bitset    []byte
	size      uint64
	hashCount uint64
}

func NewBloomFilter(size uint64, hashCount uint64) *BloomFilter {
	// We need 'size' bits.
	// Since 1 byte = 8 bits, we need size/8 bytes.
	// We add 7 before dividing to round up (ceiling division).
	byteSize := (size + 7) / 8

	return &BloomFilter{
		bitset:    make([]byte, byteSize),
		size:      size,
		hashCount: hashCount,
	}
}

func (bf *BloomFilter) Add(key string) {
	h := hash(key)

	for i := uint64(0); i < bf.hashCount; i++ {
		// SIMPLIFIED HASHING:
		// Just add 'i' to the hash to get a "new" position
		position := (h + (i * 0x9e3779b9)) % bf.size
		byteIndex := position / 8
		bitIndex := position % 8
		// Turn ON the specific bit using bitwise OR (|)
		bf.bitset[byteIndex] |= (1 << bitIndex)
	}
}

// Contains checks if a key MIGHT be in the set
// Returns true if possibly present, false if definitely not.
func (bf *BloomFilter) Contains(key []byte) bool {
	h1 := hash(string(key))
	for i := uint64(0); i < bf.hashCount; i++ {
		pos := (h1 + (i * 0x9e3779b9)) % bf.size
		byteIndex := pos / 8
		bitIndex := pos % 8
		// Check if the bit is OFF
		if (bf.bitset[byteIndex] & (1 << bitIndex)) == 0 {
			return false // Definitely not here
		}
	}
	return true // Possibly here
}

func hash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func (bf *BloomFilter) Encode() []byte {
	buffer := make([]byte, 16+len(bf.bitset))
	binary.LittleEndian.PutUint64(buffer[0:8], bf.size)
	binary.LittleEndian.PutUint64(buffer[8:16], bf.hashCount)
	copy(buffer[16:], bf.bitset)
	return buffer
}

func DecodeBloomFilter(data []byte) *BloomFilter {
	if len(data) < 16 {
		return nil
	}
	size := binary.LittleEndian.Uint64(data[0:8])
	hashCount := binary.LittleEndian.Uint64(data[8:16])
	// The rest of the data is the bitset
	bitset := make([]byte, len(data)-16)
	copy(bitset, data[16:])
	return &BloomFilter{
		bitset:    bitset,
		size:      size,
		hashCount: hashCount,
	}
}
