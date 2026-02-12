# Chill-DB Storage Engine Documentation

## Overview

Chill-DB is a **Log-Structured Merge (LSM) Tree-based key-value storage engine** written in Go. It implements a modern database architecture with durability guarantees, concurrent access, and efficient data organization.

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    Client Application                    │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
        ┌────────────────────────────┐
        │   SQL Parser (parser.go)   │
        │   Converts SQL → Commands  │
        └────────────┬───────────────┘
                     │
                     ▼
        ┌────────────────────────────┐
        │  API Handler (handler.go)  │
        │  HTTP Endpoints            │
        └────────────┬───────────────┘
                     │
        ┌────────────┴────────────┐
        ▼                         ▼
   ┌─────────────┐        ┌──────────────┐
   │  Memtable   │        │ Write-Ahead  │
   │ (In-Memory) │        │ Log (WAL)    │
   └──────┬──────┘        └──────┬───────┘
          │                      │
          └──────────┬───────────┘
                     │
          ┌──────────▼──────────┐
          │  Compaction Engine  │
          │ (Flushes & Merges)  │
          └──────────┬──────────┘
                     │
                     ▼
        ┌────────────────────────────┐
        │  SSTables on Disk          │
        │  (Sorted String Tables)    │
        │  - Level 0                 │
        │  - Level 1                 │
        │  - Level N                 │
        └────────────────────────────┘
```

---

## Core Components

### 1. **Memtable** (`memtable.go`)

**Purpose**: In-memory buffer for write operations

**Characteristics**:

- Stores key-value pairs in memory before flushing to disk
- Provides fast write performance
- Sorted data structure (likely Red-Black Tree or Skip List)
- Fixed size limit—triggers flush when exceeded

**Operations**:

```
Put(key, value)     → O(log n) insertion
Get(key)            → O(log n) lookup
Range(startKey, endKey) → O(log n + k) where k is results
Flush()             → Converts to SSTable
```

**Lifecycle**:

1. Writes are appended to WAL
2. Insert into memtable
3. When size threshold exceeded → flush to disk

---

### 2. **Write-Ahead Log (WAL)** (`wal.go`)

**Purpose**: Ensures durability and enables recovery

**How It Works**:

```
Write Request
    ↓
Write to WAL (disk) ← Synchronous
    ↓
Return ACK to client (only if WAL write succeeds)
    ↓
Write to memtable (async)
```

**Key Features**:

- **Durability Guarantee**: Before acknowledging a write, it's safely on disk
- **Recovery**: On crash, replay WAL to restore memtable state
- **Format**: Sequential log entries (timestamp, key, value, operation type)

**Entry Structure** (typical):

```
[Timestamp][Operation][Key Length][Key][Value Length][Value][Checksum]
```

**Recovery Process**:

```
On Startup:
1. Check for incomplete/corrupted entries
2. Replay entries sequentially
3. Rebuild memtable state
4. Resume normal operations
```

---

### 3. **SSTable (Sorted String Table)** (`sstable.go`)

**Purpose**: Immutable, sorted on-disk storage

**Structure**:

```
┌─────────────────────────────────┐
│     Data Block (Sorted KVs)     │
├─────────────────────────────────┤
│     Index Block (Key Offsets)   │
├─────────────────────────────────┤
│     Bloom Filter (Optional)     │
├─────────────────────────────────┤
│     Metadata Footer             │
│  (Version, Timestamps, etc.)    │
└─────────────────────────────────┘
```

**Advantages**:

- **Sorted**: Enables fast binary search and range queries
- **Immutable**: No locking needed for reads
- **Compressed**: Can use compression algorithms (optional)

**Read Path**:

```
Get(key):
1. Check Bloom filter (quick negative check)
2. Binary search in index block
3. Jump to data block offset
4. Return value
```

---

### 4. **Compaction Engine** (`compaction.go`)

**Purpose**: Manages LSM tree levels and keeps system efficient

**LSM Tree Structure**:

```
Level 0: [SSTable] [SSTable] [SSTable]  ← Recent flushes from memtable
         (may overlap, small)

Level 1: [SSTable] [SSTable]             ← Compacted from L0
         (no overlap, larger)

Level 2: [SSTable] [SSTable]             ← Further compacted
         (no overlap, larger)

Level N: [SSTable] ...                   ← Deep history
```

**Compaction Types**:

1. **Memtable Flush** (L0):
   - Triggered when memtable reaches size limit
   - Converts memtable → SSTable
   - Placed in Level 0

2. **Level Compaction**:
   - When Level i exceeds threshold, compact with Level i+1
   - Merge overlapping SSTables
   - Remove deleted keys
   - Write merged result to Level i+1

**Compaction Algorithm** (Typical):

```
Compact(level):
1. Select SSTable(s) from level_i that exceed size limit
2. Find overlapping SSTables in level_i+1
3. Merge-sort all key-value pairs
4. Write to new SSTables in level_i+1
5. Delete old SSTables
6. Update manifest
```

**Write Amplification**:

- Each key may be rewritten multiple times during compaction
- Trade-off: Better read performance, higher write cost

---

### 5. **File Repository** (`file_repo.go`)

**Purpose**: Manages SSTable files on disk

**Responsibilities**:

- Create new SSTable files
- Delete obsolete SSTables
- Track file metadata (version, timestamps)
- Recover file list on startup
- Prevent concurrent access conflicts

**File Organization**:

```
data/
├── testdb/
│   ├── level0/
│   │   ├── 0000000001.sst
│   │   ├── 0000000002.sst
│   │   └── ...
│   ├── level1/
│   │   ├── 0000000010.sst
│   │   └── ...
│   ├── manifest.log      ← Tracks SSTable versions
│   └── CURRENT           ← Latest manifest file
```

---

### 6. **Repository Interface** (`repository.go`)

**Purpose**: Abstract data access layer

**Core Operations**:

```go
Put(key, value) error       // Insert/Update
Get(key) (value, error)     // Retrieve
Delete(key) error           // Remove
Range(start, end) []KV      // Scan range
Close() error               // Cleanup
```

**Implementation Flow**:

```
Put Request:
1. Write to WAL
2. Insert into memtable
3. Check if flush needed
4. Return success

Get Request:
1. Check memtable (hot data)
2. Search Level 0 SSTables
3. Search Level 1 SSTables
4. ... continue levels
5. Return value or NOT_FOUND
```

---

### 7. **LSM Repository** (`lsm_repo.go`)

**Purpose**: Main entry point implementing LSM tree logic

**Key Methods**:

- `Put()`: Write with durability
- `Get()`: Multi-level search
- `Delete()`: Mark as deleted (tombstone)
- `Compact()`: Trigger compaction cycle
- `Recover()`: Restore from WAL on startup

---

### 8. **Domain Models** (`domain/models.go`)

**Purpose**: Business logic entities

**Typical Entities**:

```go
type KeyValue struct {
    Key       string
    Value     []byte
    Timestamp int64
    Deleted   bool  // Tombstone for deletions
}

type Table struct {
    Name    string
    Columns []Column
    Rows    []KeyValue
}
```

---

### 9. **SQL Parser** (`sql/parser.go`)

**Purpose**: Convert SQL queries to storage engine commands

**Supported Operations**:

- `SELECT * FROM table WHERE key = 'x'`
- `INSERT INTO table (key, value) VALUES (...)`
- `DELETE FROM table WHERE key = 'x'`
- `UPDATE table SET value = 'y' WHERE key = 'x'`

**Parsing Flow**:

```
"SELECT * FROM users WHERE id = 5"
    ↓
Tokenize
    ↓
Parse SQL Grammar
    ↓
Generate Query AST
    ↓
Convert to Repository Operations
    ↓
Execute against LSM tree
```

---

### 10. **API Handler** (`api/handler.go`)

**Purpose**: HTTP endpoints for client interaction

**Typical Endpoints**:

```
POST   /api/put      → Store key-value
GET    /api/get?key=x → Retrieve value
DELETE /api/del?key=x → Delete key
GET    /api/range    → Range scan
```

---

## Write Path (Detailed)

```
Client: PUT(user_123, {"name": "Alice", "age": 30})
  │
  ├─ 1. API Handler receives HTTP POST
  │  └─ Validates input (key, value)
  │
  ├─ 2. Write to WAL
  │  └─ Synchronous disk write (durability guarantee)
  │     └─ Returns success only if WAL persistent
  │
  ├─ 3. Insert into Memtable
  │  └─ In-memory sorted structure
  │     └─ O(log n) insertion
  │
  └─ 4. Check Memtable Size
     ├─ If size < threshold:
     │  └─ Return ACK to client (write done)
     │
     └─ If size ≥ threshold:
        └─ Trigger Memtable Flush:
           ├─ Convert memtable to SSTable
           ├─ Write to disk (Level 0)
           ├─ Clear memtable
           └─ Check if Level 0 needs compaction
```

---

## Read Path (Detailed)

```
Client: GET(user_123)
  │
  ├─ 1. Check Memtable (hot path)
  │  └─ Binary search in sorted structure
  │     ├─ Found: Return value (fast)
  │     └─ Not found: Continue to SSTables
  │
  ├─ 2. Check Level 0 SSTables
  │  ├─ Read index block
  │  ├─ Bloom filter check (quick negative test)
  │  └─ Binary search in data block
  │
  ├─ 3. Check Level 1 SSTables (if needed)
  │  └─ Similar process
  │
  ├─ 4. Check Level N SSTables
  │  └─ Continue until found or exhausted
  │
  └─ 5. Return Result
     ├─ Found: Return value
     └─ Not found: Return NULL/NOT_FOUND
```

---

## Compaction Strategy (Leveled Compaction)

**Trigger Conditions**:

1. Memtable reaches size limit → Flush to Level 0
2. Level 0 has too many SSTables → Compact with Level 1
3. Level i exceeds size limit → Compact with Level i+1

**Example Compaction Cycle**:

```
Before:
Level 0: [1.sst][2.sst][3.sst]  (3 × 2MB = 6MB)
Level 1: [4.sst][5.sst]          (2 × 10MB = 20MB)

Trigger: Level 0 exceeds 3 SSTables

During Compaction:
1. Merge-sort: [1.sst][2.sst][3.sst] + [4.sst][5.sst]
2. Remove duplicates (keep latest)
3. Remove tombstones (deleted keys)
4. Write sorted output

After:
Level 0: []                       (empty)
Level 1: [6.sst][7.sst][8.sst]   (new merged SSTables)
```

---

## Recovery Mechanism

**Scenario**: Server crashes with uncompacted memtable

```
Crash occurs at:
  Memory: [key1, key2, key3] in memtable
  Disk:   [WAL entries for key1, key2, key3]

Startup Recovery:
1. Detect incomplete shutdown
2. Open WAL file
3. Read each entry:
   Entry 1: PUT key1=value1 → insert into memtable
   Entry 2: PUT key2=value2 → insert into memtable
   Entry 3: PUT key3=value3 → insert into memtable
4. Verify integrity (checksums)
5. Rebuild memtable state
6. Resume normal operations
7. Eventually flush recovered data to SSTables
```

---

## Performance Characteristics

| Operation      | Time Complexity | Notes                                 |
| -------------- | --------------- | ------------------------------------- |
| **Put**        | O(1) amortized  | WAL write + memtable insert           |
| **Get**        | O(log m + k)    | m = memtable size, k = SSTable levels |
| **Delete**     | O(1) amortized  | Tombstone + memtable                  |
| **Range**      | O(k × log n)    | k = result size, n = total records    |
| **Compaction** | O(n log n)      | Merge-sort of overlapping SSTables    |

---

## Key Design Decisions

| Decision               | Benefit               | Trade-off                         |
| ---------------------- | --------------------- | --------------------------------- |
| **LSM Tree**           | Fast writes           | Slower reads, write amplification |
| **Memtable**           | In-memory speed       | Limited by available RAM          |
| **WAL**                | Durability guarantee  | Extra I/O overhead                |
| **Leveled Compaction** | Predictable read cost | Higher compaction overhead        |
| **Immutable SSTables** | No locking needed     | Complex compaction logic          |

---

## Testing Strategy

### Crash-Test (`cmd/crash-test/main.go`)

**Purpose**: Verify recovery correctness after crashes

```
Scenario:
1. Write N key-value pairs
2. Simulate crash (kill process)
3. Restart
4. Verify all written data recovered correctly
5. Check data integrity
```

### Stress-Test (`cmd/stress-test/main.go`)

**Purpose**: Validate under high concurrency

```
Scenario:
1. Spawn multiple goroutines
2. Each performs random PUT/GET/DELETE
3. Monitor for race conditions
4. Check final consistency
5. Measure throughput/latency
```

### Server Tests (`tests/server_test.go`)

**Purpose**: Unit and integration tests

---

## Limitations & Future Improvements

**Current Limitations**:

- Single-node only (no replication)
- In-process only (no remote clients)
- Basic SQL support (no joins, aggregations)
- No transactions or ACID isolation
- Limited compaction tuning options

**Future Enhancements**:

1. Add transaction support (MVCC)
2. Implement range queries optimization
3. Add compression (Snappy, ZSTD)
4. Distributed replication
5. Advanced query optimization
6. Better memory management

---

## Data Directory Structure

```
data/
├── testdb/              # Database instance
│   ├── WAL/
│   │   └── wal.log     # Write-ahead log
│   ├── SST/
│   │   ├── level0/
│   │   ├── level1/
│   │   └── ...
│   ├── manifest.log     # Version tracking
│   └── CURRENT          # Latest manifest
data_crash/              # For testing crash recovery
```

---

## Conclusion

Chill-DB implements a **production-inspired LSM Tree storage engine** with:

- ✅ Durability (WAL)
- ✅ Efficiency (memtable + compaction)
- ✅ Scalability (leveled structure)
- ✅ Reliability (crash recovery)

This architecture powers real databases like RocksDB, LevelDB, and Cassandra.
