package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func ExecuteSQL(dbName string, query string) (string, error) {
	query = strings.TrimSpace(query)
	upperQuery := strings.ToUpper(query)
	if strings.HasPrefix(upperQuery, "SELECT") {
		return parseSelect(dbName, query)
	} else if strings.HasPrefix(upperQuery, "INSERT") {
		return parseInsert(dbName, query)
	} else if strings.HasPrefix(upperQuery, "DELETE") {
		return parseDelete(dbName, query)
	}

	return "", fmt.Errorf("unknown command: %s", query)
}

func parseInsert(dbName string, query string) (string, error) {
	re := regexp.MustCompile(`(?i)^INSERT\s+INTO\s+(\w+)\s+VALUES\s*\((.+)\)$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(query))

	if len(matches) < 3 {
		return "", fmt.Errorf("syntax error. usage: INSERT INTO <table> VALUES (v1, v2...)")
	}

	tableName := matches[1]
	rawValues := matches[2]

	parts := strings.Split(rawValues, ",")
	var cleanParts []string

	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "'\"")
		cleanParts = append(cleanParts, p)
	}

	finalValues := strings.Join(cleanParts, ",")
	cmd := exec.Command("./scripts/data_ops.sh", "insert", dbName, tableName, finalValues)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("system error: %s", string(output))
	}

	return "Row inserted successfully", nil
}

func parseDelete(dbName string, query string) (string, error) {
	re := regexp.MustCompile(`(?i)^DELETE\s+FROM\s+(\w+)(?:\s+WHERE\s+(\w+)\s*=\s*(.+))?$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(query))

	if len(matches) < 2 {
		return "", fmt.Errorf("syntax error. usage: DELETE FROM <table> WHERE <filter>")
	}

	tableName := strings.TrimSpace(matches[1])
	var filterCol, filterVal string

	if len(matches) == 4 && matches[2] != "" {
		filterCol = strings.TrimSpace(matches[2])
		filterVal = strings.TrimSpace(matches[3])
		filterVal = strings.Trim(filterVal, "'\"")
	}

	var cmd *exec.Cmd
	if filterCol == "" {
		cmd = exec.Command("./scripts/data_ops.sh", "delete", dbName, tableName)
	} else {
		cmd = exec.Command("./scripts/data_ops.sh", "delete", dbName, tableName, filterVal)
	}
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("system error: %s", string(output))
	}

	return string(output), nil

}
func parseSelect(dbName string, query string) (string, error) {
	re := regexp.MustCompile(`(?i)^SELECT\s+(.*?)\s+FROM\s+(\w+)(?:\s+WHERE\s+(\w+)\s*=\s*(.+))?$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(query))

	if len(matches) < 3 {
		return "", fmt.Errorf("syntax error. usage: SELECT * FROM <table>")
	}

	columnsStr := strings.TrimSpace(matches[1])
	tableName := strings.TrimSpace(matches[2])

	var filterCol, filterVal string
	if len(matches) == 4 {
		filterCol = strings.TrimSpace(matches[3])
		filterVal = strings.TrimSpace(matches[4])
		filterVal = strings.Trim(filterVal, "'\"")
	}
	cmd := exec.Command("./scripts/data_ops.sh", "select", dbName, tableName, filterCol, filterVal)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("system error: %s", string(output))
	}

	rawData := string(output)
	if rawData == "" {
		return "", nil
	}

	if columnsStr == "*" {
		return rawData, nil
	}

	requestedCols := strings.Split(columnsStr, ",")
	indices, err := getColumnIndices(dbName, tableName, requestedCols)
	if err != nil {
		return "", err
	}

	var resultBuilder strings.Builder
	lines := strings.Split(rawData, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		cells := strings.Split(line, ",")
		var newRow []string

		for _, index := range indices {
			if index < len(cells) {
				newRow = append(newRow, cells[index])
			}
		}

		resultBuilder.WriteString(strings.Join(newRow, ",") + "\n")
	}

	return resultBuilder.String(), nil
}

func getColumnIndices(dbName, tableName string, reqColumns []string) ([]int, error) {
	metaPath := fmt.Sprintf("./data/%s/%s.meta", dbName, tableName)
	content, err := os.ReadFile(metaPath)

	if err != nil {
		absPath, _ := filepath.Abs(metaPath)
		return nil, fmt.Errorf("debug: tried to open '%s' (Absolute: %s). System Error: %v", metaPath, absPath, err)
	}

	rawHeaders := strings.Split(strings.TrimSpace(string(content)), ",")
	headerMap := make(map[string]int)

	for i, h := range rawHeaders {
		h = strings.TrimSpace(h)
		parts := strings.Split(h, ":")
		colName := parts[0]
		headerMap[colName] = i
	}

	var indices []int
	for _, col := range reqColumns {
		col = strings.TrimSpace(col)
		index, exists := headerMap[col]

		if !exists {
			return nil, fmt.Errorf("column '%s' does not exist in table '%s'", col, tableName)
		}

		indices = append(indices, index)
	}

	return indices, nil
}
