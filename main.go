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

type InsertRequest struct {
	DBName    string `json:"db_name"`
	TableName string `json:"table_name"`
	Values    string `json:"values"`
}

type SelectRequest struct {
	DBName     string `json:"db_name"`
	TableName  string `json:"table_name"`
	ColumnName string `json:"column"` // Optional
	Value      string `json:"value"`  // Optional
}

type UpdateRequest struct {
	DBName    string `json:"db_name"`
	TableName string `json:"table_name"`
	PKValue   string `json:"pk_value"`
	Column    string `json:"column"`
	Value     string `json:"value"`
}

type DeleteRequest struct {
	DBName    string `json:"db_name"`
	TableName string `json:"table_name"`
	PKValue   string `json:"pk_value"`
}

func main() {
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	http.HandleFunc("/databases", listDatabases)

	http.HandleFunc("/database/create", createDatabase)

	http.HandleFunc("/table/create", createTable)

	http.HandleFunc("/data/insert", insertRow)

	http.HandleFunc("/data/query", queryTable)

	http.HandleFunc("/data/update", updateRow)

	http.HandleFunc("/data/delete", deleteRow)

	fmt.Println("🚀 Server is running on http://localhost:8080 ...")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Error starting the server", nil)
	}

}

func listDatabases(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request for /databases")
	cmd := exec.Command("./scripts/db_ops.sh", "list")
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
	for _, char := range req.Name {
		if char == ' ' {
			http.Error(w, "Database names cannot contain spaces", http.StatusBadRequest)
			return
		}
	}

	cmd := exec.Command("./scripts/db_ops.sh", "create", req.Name)
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

	cmd := exec.Command("./scripts/table_ops.sh", "create", req.DBName, req.TableName, req.Columns)
	output, err := cmd.CombinedOutput()

	if err != nil {
		http.Error(w, string(output), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Success: %s", string(output))
}

func insertRow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only Post method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InsertRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.TableName == "" || req.Values == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("./scripts/data_ops.sh", "insert", req.DBName, req.TableName, req.Values)
	output, err := cmd.CombinedOutput()

	if err != nil {
		http.Error(w, string(output), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Success: %s", string(output))
}

func queryTable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only Post method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SelectRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.TableName == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("./scripts/data_ops.sh", "select", req.DBName, req.TableName, req.ColumnName, req.Value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, string(output), http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, string(output))
}

func updateRow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.TableName == "" || req.PKValue == "" || req.Column == "" || req.Value == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("./scripts/data_ops.sh", "update", req.DBName, req.TableName, req.PKValue, req.Column, req.Value)
	output, err := cmd.CombinedOutput()

	if err != nil {
		http.Error(w, string(output), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Success: %s", string(output))
}

func deleteRow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.PKValue == "" || req.TableName == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("./scripts/data_ops.sh", "delete", req.DBName, req.TableName, req.PKValue)

	output, err := cmd.CombinedOutput()

	if err != nil {
		http.Error(w, string(output), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Success: %s", string(output))
}
