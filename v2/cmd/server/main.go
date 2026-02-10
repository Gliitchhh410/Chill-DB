package main

import (
	"chill-db/internal/api"
	"chill-db/internal/db"
	"fmt"
	"log"
	"net/http"
)

func main() {
	repo, err := db.NewFileRepository("./data")
	if err != nil {
		log.Fatalf("Failed to initialize database engine: %v", err)
	}

	handler := api.NewHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/databases", handler.ListDatabases)
	mux.HandleFunc("/database/create", handler.CreateDatabase)
	mux.HandleFunc("/sql", handler.HandleSQL)
	mux.HandleFunc("/database/delete", handler.DropDatabase)

	fmt.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))

}
