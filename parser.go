package main

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func ExecuteSQL(dbName string, query string) (string, error) {
	query = strings.TrimSpace(query)
	upperQuery := strings.ToUpper(query)
	if strings.HasPrefix(upperQuery, "SELECT") {
		return parseSelect(dbName, query)
	} else if strings.HasPrefix(upperQuery, "INSERT") {
		return "", errors.New("INSERT command is under construction")
	} else if strings.HasPrefix(upperQuery, "DELETE") {
		return "", errors.New("DELETE command is under construction")
	}

	return "", fmt.Errorf("unknown command: %s", query)
}

func parseSelect(dbName string, query string) (string, error) {
	re := regexp.MustCompile(`(?i)^SELECT\s+(.*?)\s+FROM\s+(\w+)$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(query))

	if len(matches) < 3 {
		return "", fmt.Errorf("syntax error. usage: SELECT * FROM <table>")
	}

	columns := strings.TrimSpace(matches[1])
	tableName := strings.TrimSpace(matches[2])

	if columns == "*" {
		cmd := exec.Command("./scripts/data_ops.sh", "select", dbName, tableName, "", "")
		output, err := cmd.CombinedOutput()

		if err != nil {
			return "", fmt.Errorf("system error: %s", string(output))
		}
		return string(output), nil
	}
	return "", errors.New("selecting specific columns is not supported yet")
}
