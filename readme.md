# Chill-DB ❄️

![Go Version](https://img.shields.io/github/go-mod/go-version/Gliitchhh410/Chill-DB?filename=v2/go.mod)
![License](https://img.shields.io/badge/license-MIT-blue)

A custom-built Relational Database Management System (RDBMS) that has evolved from a **Bash-based Storage Engine** (v1) to a high-performance **LSM Tree implementation in pure Go** (v2).

**Tech Stack:** Go (Golang), Bash Scripting, Vanilla JS, Tailwind CSS.

## 🚀 Why I Built This

I wanted to understand databases at a low level—not just how to use them, but how they work under the hood. I built `Chill-DB` to explore:

- **Lexing & Parsing:** Writing a custom SQL engine to understand `SELECT`, `UPDATE`, `INSERT`, `DELETE`.
- **Systems Programming:** Using Bash for raw file manipulation and storage efficiency.
- **Full Stack Integration:** connecting a low-level backend to a modern frontend via REST API.

## 🔄 Evolution: v1 vs v2

The project is divided into two distinct versions, showcasing the learning journey:

### v1: The Prototype (Bash-based)

- **Storage:** Relied on Bash scripts (`awk`, `sed`) for file manipulation.
- **Concurrency:** Limited by file locking.
- **Format:** Flat CSV files.

### v2: The Rewrite (Pure Go + LSM)

- **Storage:** Implements **LSM Trees** (Log-Structured Merge Trees) for high write throughput.
- **Durability:** Uses a **Write Ahead Log (WAL)** to ensure data safety on crash.
- **Performance:** In-memory **MemTable** buffering and binary storage.
- **Cross-Platform:** Runs natively on Windows, Linux, and macOS without WSL.

## 🛠 Architecture (v1)

The v1 system follows a 3-tier architecture:

1.  **The Frontend (Client):**
    - A JavaScript/Tailwind dashboard that sends SQL commands via JSON.
    - Visualizes data in responsive grids.
2.  **The Middleware (Go Server):**
    - Listens on port `8080`.
    - **SQL Parser:** Uses Regex and String Tokenizing to break down queries (e.g., extracting `WHERE` clauses).
    - **Orchestrator:** Validates syntax and spawns system processes.

3.  **The Storage Engine (Bash):**
    - Handles the physical data layer (`.data` and `.meta` files).
    - Implements **Projection** (selecting columns) and **Selection** (filtering rows) using `awk` stream processing.
    - Ensures data integrity with atomic file moves.

## 📡 API Reference (v2)

Chill-DB v2 exposes a REST API for programmatic access. Note that v2 consolidates table operations into the SQL engine.

| Method | Endpoint           | Description       | Body                                        |
| :----- | :----------------- | :---------------- | :------------------------------------------ |
| `POST` | `/database/create` | Create a new DB   | `{"name": "mydb"}`                          |
| `POST` | `/sql`             | Execute SQL Query | `{"db_name": "mydb", "query": "SELECT..."}` |
| `GET`  | `/databases`       | List databases    | None                                        |

## ⚡ Features

- ✅ **SQL Engine:** Supports `SELECT`, `INSERT`, `UPDATE`, `DELETE`.
- ✅ **Complex Queries:** Handles `WHERE` clauses and Column Projections.
- ✅ **Data Persistence:** Custom file-based storage format.
- ✅ **Interactive UI:** A CLI-style web interface for executing commands.

## 🔧 Installation & Usage

1. **Clone the repo**

   ```bash
   git clone https://github.com/Gliitchhh410/Chill-DB.git

   cd Chill-DB
   ```

2. **Start The Engine**

   ```bash
   go run .
   ```

3. **Open The Dashboard**

   Navigate to `http://localhost:8080` in your browser.

4. **Run a Query**

```SQL
-- Create a new user
INSERT INTO users VALUES (1, 'Ahmed', 'Admin')

-- Update user role
UPDATE users SET role='SuperUser' WHERE id=1

-- Select specific columns
SELECT name, role FROM users WHERE name='Ahmed'

-- Delete the user
DELETE FROM users WHERE id=1
```
