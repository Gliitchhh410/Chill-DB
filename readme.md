# Chill-DB ❄️ 

A custom-built Relational Database Management System (RDBMS) featuring a **SQL Parser in Go**, a **Bash-based Storage Engine**, and a **Web-based Dashboard**.


**Tech Stack:** Go (Golang), Bash Scripting, Vanilla JS, Tailwind CSS.

## 🚀 Why I Built This
I wanted to understand databases at a low level—not just how to use them, but how they work under the hood. I built `Chill-DB` to explore:
* **Lexing & Parsing:** Writing a custom SQL engine to understand `SELECT`, `UPDATE`, `INSERT`, `DELETE`.
* **Systems Programming:** Using Bash for raw file manipulation and storage efficiency.
* **Full Stack Integration:** connecting a low-level backend to a modern frontend via REST API.

## 🛠 Architecture

The system follows a 3-tier architecture:

1.  **The Frontend (Client):** 
    * A JavaScript/Tailwind dashboard that sends SQL commands via JSON.
    * Visualizes data in responsive grids.
    
2.  **The Middleware (Go Server):** 
    * Listens on port `8080`.
    * **SQL Parser:** Uses Regex and String Tokenizing to break down queries (e.g., extracting `WHERE` clauses).
    * **Orchestrator:** Validates syntax and spawns system processes.

3.  **The Storage Engine (Bash):** 
    * Handles the physical data layer (`.data` and `.meta` files).
    * Implements **Projection** (selecting columns) and **Selection** (filtering rows) using `awk` stream processing.
    * Ensures data integrity with atomic file moves.

## ⚡ Features
* ✅ **SQL Engine:** Supports `SELECT`, `INSERT`, `UPDATE`, `DELETE`.
* ✅ **Complex Queries:** Handles `WHERE` clauses and Column Projections.
* ✅ **Data Persistence:** Custom file-based storage format.
* ✅ **Interactive UI:** A CLI-style web interface for executing commands.

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