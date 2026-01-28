package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

type DBRequest struct {
	Name string `json:"name"`
}

type TableRequest struct {
	DBName    string `json:"db_name"`
	TableName string `json:"table_name"`
	Columns   string `json:"columns"` // e.g., "id:int,name:string"
}

func main() {
	http.HandleFunc("/database", listDatabases)

	http.HandleFunc("/database/create", createDatabase)

	http.HandleFunc("/table/create", createTable)

	fmt.Println("🚀 Server is running on http://localhost:8080 ...")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Error starting the server", nil)
	}

}

func listDatabases(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request for /databases")
	cmd := exec.Command("./db_ops.sh", "list")
	output, err := cmd.CombinedOutput()

	if err != nil {
		http.Error(w, "Failed to list databases", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, string(output))
}

func createDatabase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DBRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Database name is required", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("./db_ops.sh", "create", req.Name)
	output, err := cmd.CombinedOutput()

	if err != nil {
		http.Error(w, string(output), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated) // HTTP 201 Created
	fmt.Fprintf(w, "Success: %s", string(output))
}

func createTable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only Post method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TableRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.TableName == "" || req.Columns == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("./table_ops.sh", "create", req.DBName, req.TableName, req.Columns)
	output, err := cmd.CombinedOutput()

	if err != nil {
		http.Error(w, string(output), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Success: %s", string(output))
}
