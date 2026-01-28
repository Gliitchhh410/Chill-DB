package main

import (
	"fmt"
	"net/http"
	"os/exec"
)

func main() {
	http.HandleFunc("/databases", listDatabases)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to Chill-DB! Visit /databases to see your data.")
	})

	fmt.Println("Server is running on port 8080...")

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
