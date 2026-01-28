package main

import (
	"fmt"
	"os/exec"
)

func main() {
	fmt.Println("Go is speaking, SILENCE!!")
	cmd := exec.Command("./db_ops.sh", "list")

	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("a crash Occurred")
		fmt.Printf("Error details: %s\n", err)
	}
	fmt.Println("✅ Bash responded:")
	fmt.Println(string(output))

}
