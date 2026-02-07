package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"chill-db/internal/api"
	"chill-db/internal/db"
)

// Define the request structures locally.
type DBRequest struct {
	Name string `json:"name"`
}

type SQLRequest struct {
	DBName string `json:"db_name"`
	Query  string `json:"query"`
}

func TestEndToEnd(t *testing.T) {
	// 1. SETUP: Create a temporary directory for the test data
	tempDir, err := os.MkdirTemp("", "chill-db-integration")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up files after test runs

	// 2. INITIALIZE: Wire up the components (Repo -> Handler)
	repo, err := db.NewFileRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}
	handler := api.NewHandler(repo)

	// 3. ROUTER: Recreate the server routing logic
	mux := http.NewServeMux()
	mux.HandleFunc("/database/create", handler.CreateDatabase)
	mux.HandleFunc("/sql", handler.HandleSQL)
	// Add other routes if needed, e.g., /databases

	// 4. HELPER: A function to reduce code repetition
	sendRequest := func(method, target string, body interface{}) *httptest.ResponseRecorder {
		var bodyBytes []byte
		if body != nil {
			bodyBytes, _ = json.Marshal(body)
		}

		req := httptest.NewRequest(method, target, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		return rr
	}

	// --- STEP 1: Create Database ---
	t.Run("1. Create Database", func(t *testing.T) {
		req := DBRequest{Name: "integration_test_db"}
		resp := sendRequest("POST", "/database/create", req)

		if resp.Code != http.StatusOK && resp.Code != http.StatusCreated {
			t.Errorf("Expected 200/201, got %d. Response: %s", resp.Code, resp.Body.String())
		}
	})

	// --- STEP 2: Create Table ---
	t.Run("2. Create Table", func(t *testing.T) {
		req := SQLRequest{
			DBName: "integration_test_db",
			Query:  "CREATE TABLE users (id int, name string, age int)",
		}
		resp := sendRequest("POST", "/sql", req)

		if resp.Code != http.StatusOK {
			t.Fatalf("Failed to create table. Code: %d, Body: %s", resp.Code, resp.Body.String())
		}
	})

	// --- STEP 3: Insert Data ---
	t.Run("3. Insert Data", func(t *testing.T) {
		req := SQLRequest{
			DBName: "integration_test_db",
			Query:  "INSERT INTO users VALUES (1, alice, 30)",
		}
		resp := sendRequest("POST", "/sql", req)

		if resp.Code != http.StatusOK {
			t.Fatalf("Failed to insert row. Code: %d, Body: %s", resp.Code, resp.Body.String())
		}
	})

	// --- STEP 4: Select Data ---
	t.Run("4. Select Data", func(t *testing.T) {
		req := SQLRequest{
			DBName: "integration_test_db",
			Query:  "SELECT * FROM users",
		}
		resp := sendRequest("POST", "/sql", req)

		if resp.Code != http.StatusOK {
			t.Fatalf("Select failed. Code: %d, Body: %s", resp.Code, resp.Body.String())
		}

		// Validation: Ensure the data we inserted actually came back
		if !strings.Contains(resp.Body.String(), "alice") {
			t.Errorf("Expected response to contain 'alice', got: %s", resp.Body.String())
		}
	})
}
