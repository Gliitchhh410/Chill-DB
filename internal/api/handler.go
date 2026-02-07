package api

import (
	"chill-db/internal/db"
	"chill-db/internal/sql"
	"encoding/json"
	"net/http"
)

type Handler struct {
	Repo db.Repository
}

func NewHandler(repo db.Repository) *Handler { // NewHandler is the constructor/factory for Handler, always take the convention "NewStruct", it's also a form of dependency injection
	return &Handler{Repo: repo}
}

type DBRequest struct { // DTO (Data Transfer Object)
	Name string `json:"name"` // It specifically tells the Go JSON encoder/decoder: "When reading JSON, look for a key named name (lowercase) and map its value to this struct field."
}

type SQLRequest struct {
	DBName string `json:"db_name"`
	Query  string `json:"query"`
}

// CreateDatabase handles POST /database/create
func (h *Handler) CreateDatabase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DBRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Database name is required", http.StatusBadRequest)
		return
	}
	// Call the Pure Go Engine
	if err := h.Repo.CreateDatabase(r.Context(), req.Name); err != nil {
		http.Error(w, "Failed to create database", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Database created successfully"))
}

// ListDatabases handles GET /databases

func (h *Handler) ListDatabases(w http.ResponseWriter, r *http.Request) {
	// Call Go engine
	dbs, err := h.Repo.ListDatabases(r.Context())

	if err != nil {
		http.Error(w, "Failed to list databases", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbs)
}

func (h *Handler) HandleSQL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SQLRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	//Call the SQL Logic Layer
	//We pass the Repo to the parser so it can do the work
	result, err := sql.Execute(r.Context(), h.Repo, req.DBName, req.Query)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(result))
}
