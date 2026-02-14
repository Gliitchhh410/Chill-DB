# Chill-DB ❄️

![Go Version](https://img.shields.io/github/go-mod/go-version/Gliitchhh410/Chill-DB?filename=v2/go.mod)
![License](https://img.shields.io/badge/license-MIT-blue)

**From Bash Scripts to Byte Slices: A journey into Database Internals.**

Chill-DB is a custom database project that evolved from a **process-based orchestrator (v1)** into a high-performance **embedded LSM-tree storage engine (v2)**.

---

## 🔄 The Evolution Story

This project represents my transition from **Full-Stack/DevOps integration** to **Low-Level Systems Engineering**.

| Feature           | **v1: The Prototype** (Legacy)      | **v2: The Engine** (Current)         |
| :---------------- | :---------------------------------- | :----------------------------------- |
| **Architecture**  | **Process Orchestrator**            | **Embedded Library**                 |
| **Language**      | Go (API) + Bash (Storage) + JS (UI) | Pure Go (100%)                       |
| **Storage Model** | Flat Files & Directories            | LSM Tree (Log-Structured Merge-tree) |
| **Query Engine**  | Regex Parsing -> `awk`/`sed`        | Binary Search + Bloom Filters        |
| **Durability**    | File System Reliability             | Write-Ahead Log (WAL) + `fsync`      |
| **Latency**       | ~10-50ms (Process Overhead)         | **~0.6ms** (Direct System Calls)     |

---

## 🛠 Chill-DB v1: The Full-Stack Prototype ([Video Demo](https://drive.google.com/file/d/1WimzsQ70XHQ0gvdck6OUQXkmCfv_nbxA/view?usp=sharing))

**v1** was developed as a final project for the **Bash Scripting Course** at **ITI**. It demonstrates how to use the Linux Operating System _as_ a database by orchestrating processes.

### 🏗 Architecture (v1)

<img height="1605" alt="v1Arch" src="https://github.com/user-attachments/assets/fda5a3f7-8381-41b1-9fe4-82cce1d72118" />

1.  **Frontend:** A JavaScript/Tailwind dashboard for executing SQL.
2.  **Middleware (Go):**
    - **Regex SQL Parser:** Tokenizes `SELECT`, `INSERT`, `UPDATE` queries.
    - **Process Forking:** Uses `exec.Command` to spawn Bash scripts.
3.  **Storage (Bash):**
    - Uses `awk` and `sed` for projection and selection.
    - Manages data/metadata directories.

### ⚡ v1 Features

- ✅ **SQL Engine:** Supports `SELECT`, `INSERT`, `UPDATE`, `DELETE`.
- ✅ **Complex Queries:** Handles `WHERE` clauses and Column Projections via Regex.
- ✅ **Interactive UI:** Visualizes data in responsive grids.

## 🚀 Chill-DB v2: The High-Performance Engine

**v2** is a complete rewrite focusing on **Performance**, **Durability**, and **Memory Efficiency**. It drops the UI and shell scripts to function as a raw, embedded key-value store similar to LevelDB or RocksDB.

<img width="2816" height="1536" alt="v2Arch" src="https://github.com/user-attachments/assets/5081aa0a-42c8-43a7-b493-9dcc96fe8a36" />


### ⚙️ Technical Architecture

v2 implements a classic LSM-Tree architecture to optimize for high write throughput:

1.  **WAL (Write-Ahead Log):** Every write is appended to a binary log file first to ensure **ACID durability** in case of a crash.
2.  **MemTable:** Data is written to an in-memory balanced structure for speed.
3.  **SSTables:** When memory is full, data is flushed to immutable **Sorted String Tables** on disk.
4.  **Bloom Filter:** A probabilistic data structure sits in front of disk reads to instantly reject requests for non-existent keys (eliminating unnecessary I/O).

### 📊 v2 Benchmarks (Intel i7-10750H)

The shift to binary I/O and algorithmic optimization resulted in massive gains:

| Metric           | Time / Op   | Optimization Impact             | Implementation Detail                                 |
| :--------------- | :---------- | :------------------------------ | :---------------------------------------------------- |
| **Inserts**      | **0.6 ms**  | —                               | Optimized zero-allocation write path (<10 allocs/op). |
| **Query (Hit)**  | **0.71 ms** | **80% Faster** (vs Linear Scan) | **Binary Search** on SSTables                         |
| **Query (Miss)** | **0.18 ms** | **94% Faster** (vs Disk Seek)   | **Bloom Filters** prevent disk access on misses.      |
| **Compaction**   | **63 ms**   | **66% Faster** (vs Sequential)  | Concurrent background merging via Goroutines.         |

### 📦 v2 Installation & Usage

Chill-DB v2 is designed to be imported as a Go library.

```bash
go get [github.com/Gliitchhh410/chill-db](https://github.com/Gliitchhh410/chill-db)
```

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
