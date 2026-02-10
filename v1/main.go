package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type DBRequest struct {
	Name string `json:"name"`
}

type DBDropRequest struct {
	Name string `json:"name"`
}

type TableDropRequest struct {
	DBName    string `json:"db_name"`
	TableName string `json:"table_name"`
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

type SQLRequest struct {
	DBName string `json:"db_name"`
	Query  string `json:"query"`
}

type TableResponse struct {
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
}

func main() {
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	http.HandleFunc("/sql", handleSQL)
	http.HandleFunc("/databases", listDatabases)

	http.HandleFunc("/database/create", createDatabase)

	http.HandleFunc("/database/delete", dropDatabase)

	http.HandleFunc("/table/delete", dropTable)

	http.HandleFunc("/tables", listTables)

	http.HandleFunc("/table/create", createTable)

	http.HandleFunc("/data/insert", insertRow)

	http.HandleFunc("/data/query", queryTable)

	http.HandleFunc("/data/update", updateRow)

	http.HandleFunc("/data/delete", deleteRow)

	fmt.Println("🚀 Server is running on http://localhost:8080 ...")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Error starting the server", err)
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


	valueArgs := strings.Split(req.Values, ",")

	for i := range valueArgs {
		valueArgs[i] = strings.TrimSpace(valueArgs[i])
	}


	cmdArgs := append([]string{"insert", req.DBName, req.TableName}, valueArgs...)

	cmd := exec.Command("./scripts/data_ops.sh", cmdArgs...)


	output, err := cmd.CombinedOutput()

	if err != nil {
		http.Error(w, string(output), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Success")
}
func queryTable(w http.ResponseWriter, r *http.Request) {
	// 1. Validate Method
	if r.Method != http.MethodPost {
		http.Error(w, "Only Post method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Decode Request Body
	var req SelectRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.TableName == "" {
		http.Error(w, "Missing db_name or table_name", http.StatusBadRequest)
		return
	}

	// 3. Get Metadata (Columns)
	metaPath := fmt.Sprintf("./data/%s/%s.meta", req.DBName, req.TableName)
	metaContent, err := os.ReadFile(metaPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Table metadata not found: %v", err), http.StatusNotFound)
		return
	}

	metaReader := csv.NewReader(strings.NewReader(string(metaContent)))
	metaRecords, err := metaReader.ReadAll()

	var columns []string
	if err == nil {
		for _, record := range metaRecords {
			// FIX: Iterate over ALL fields in the row (e.g., "id:int", "name:string")
			// The previous code only read record[0], missing subsequent columns.
			for _, field := range record {
				// Split "name:string" -> ["name", "string"] -> take "name"
				parts := strings.Split(field, ":")
				if len(parts) > 0 && parts[0] != "" {
					columns = append(columns, parts[0])
				}
			}
		}
	}

	// 4. Get Data (Rows)
	cmd := exec.Command("./scripts/data_ops.sh", "select", req.DBName, req.TableName, req.ColumnName, req.Value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, string(output), http.StatusNotFound)
		return
	}

	rowReader := csv.NewReader(strings.NewReader(string(output)))
	rows, _ := rowReader.ReadAll()

	if rows == nil {
		rows = [][]string{}
	}

	// 5. Send Response
	response := TableResponse{
		Columns: columns,
		Rows:    rows,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

func listTables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DBRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}
	cmd := exec.Command("./scripts/table_ops.sh", "list", req.Name)
	output, err := cmd.CombinedOutput()

	if err != nil {
		http.Error(w, string(output), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, string(output))
}

func dropDatabase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DBDropRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}
	cmd := exec.Command("./scripts/db_ops.sh", "drop", req.Name)
	output, err := cmd.CombinedOutput()

	if err != nil {
		http.Error(w, string(output), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, string(output))

}

func dropTable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TableDropRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.TableName == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}
	cmd := exec.Command("./scripts/table_ops.sh", "drop", req.DBName, req.TableName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		http.Error(w, string(output), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, string(output))
}

func handleSQL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("Processing SQL: %s (DB: %s)\n", req.Query, req.DBName)

	result, err := ExecuteSQL(req.DBName, req.Query)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, result)
}

func handleGetTable(w http.ResponseWriter, r *http.Request) {
	db := r.URL.Query().Get("db")
	table := r.URL.Query().Get("table")

	// 1. Read the Meta file (Columns)
	metaPath := fmt.Sprintf("./data/%s/%s.meta", db, table)
	metaContent, err := os.ReadFile(metaPath)
	if err != nil {
		http.Error(w, "Table metadata not found", 404)
		return
	}

	// Parse Meta CSV
	// Assuming format: "ColumnName,Type" or just "ColumnName"
	metaReader := csv.NewReader(strings.NewReader(string(metaContent)))
	metaRecords, err := metaReader.ReadAll()
	if err != nil {
		http.Error(w, "Failed to parse metadata", 500)
		return
	}

	// Extract just the column names (assuming index 0 is the name)
	var columns []string
	for _, record := range metaRecords {
		if len(record) > 0 {
			columns = append(columns, record[0])
		}
	}

	// 2. Read the Data file (Rows)
	// (I am assuming you have logic to read rows, let's call it readRows logic)
	// For this example, I'll assume you read the .csv or .data file here
	rows := [][]string{}
	dataPath := fmt.Sprintf("./data/%s/%s.data", db, table) // Example path
	dataContent, err := os.ReadFile(dataPath)
	if err == nil {
		dataReader := csv.NewReader(strings.NewReader(string(dataContent)))
		rows, _ = dataReader.ReadAll()
	}

	// 3. Send combined JSON response
	response := map[string]interface{}{
		"columns": columns,
		"rows":    rows,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
