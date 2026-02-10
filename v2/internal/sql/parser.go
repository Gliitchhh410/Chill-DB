package sql


import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"chill-db/internal/db"
	"chill-db/internal/domain"
)


func Execute(ctx context.Context, repo db.Repository, dbName, query string) (string, error) {
	query = strings.TrimSpace(query)
	upperQuery := strings.ToUpper(query)

	switch {
	case strings.HasPrefix(upperQuery, "CREATE"):
		return parseCreate(ctx, repo, dbName, query)
	case strings.HasPrefix(upperQuery, "INSERT"):
		return parseInsert(ctx, repo, dbName, query)
	case strings.HasPrefix(upperQuery, "SELECT"):
		return parseSelect(ctx, repo, dbName, query)
	default:
		return "", fmt.Errorf("unknown or unsupported command: %s", query)
	}
}


// parseCreate: "CREATE TABLE users (id int, name string)"
func parseCreate(ctx context.Context, repo db.Repository, dbName, query string) (string, error) {
	re := regexp.MustCompile(`(?i)^CREATE\s+TABLE\s+(\w+)\s*\((.+)\)$`)
	matches := re.FindStringSubmatch(query)
	if len(matches) < 3 {
		return "", fmt.Errorf("syntax error: CREATE TABLE <table> (<columns>)")
	}

	tableName := matches[1]
	colDefs := strings.Split(matches[2], ",")
	
	var columns []domain.ColumnDefinition
	for _, def := range colDefs {
		parts := strings.Fields(strings.TrimSpace(def))
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid column def: %s", def)
		}
		columns = append(columns, domain.ColumnDefinition{Name: parts[0], Type: parts[1]})
	}

	err := repo.CreateTable(ctx, dbName, domain.TableMetaData{
		Name:    tableName,
		Columns: columns,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Table '%s' created.", tableName), nil
}

// parseInsert: "INSERT INTO users VALUES ('1', 'john')"
func parseInsert(ctx context.Context, repo db.Repository, dbName, query string) (string, error) {
	re := regexp.MustCompile(`(?i)^INSERT\s+INTO\s+(\w+)\s+VALUES\s*\((.+)\)$`)
	matches := re.FindStringSubmatch(query)
	if len(matches) < 3 {
		return "", fmt.Errorf("syntax error: INSERT INTO <table> VALUES (<values>)")
	}

	tableName := matches[1]
	rawValues := strings.Split(matches[2], ",")
	
	var row domain.Row
	for _, val := range rawValues {
		// Clean up quotes: 'john' -> john
		clean := strings.TrimSpace(val)
		clean = strings.Trim(clean, "'\"")
		row = append(row, clean)
	}

	if err := repo.InsertRow(ctx, dbName, tableName, row); err != nil {
		return "", err
	}
	return "Row inserted.", nil
}

// parseSelect: "SELECT * FROM users"
// Note: We are doing a simple SELECT * for now. Filtering (WHERE) happens here in memory.
func parseSelect(ctx context.Context, repo db.Repository, dbName, query string) (string, error) {
	re := regexp.MustCompile(`(?i)^SELECT\s+(.*?)\s+FROM\s+(\w+)`)
	matches := re.FindStringSubmatch(query)
	if len(matches) < 3 {
		return "", fmt.Errorf("syntax error: SELECT <cols> FROM <table>")
	}

	// 1. Fetch ALL data from the engine
	tableName := matches[2]
	rows, err := repo.Query(ctx, dbName, tableName)
	if err != nil {
		return "", err
	}

	// 2. Format the output (Simple CSV dump for now)
	// In a real app, you would apply WHERE clauses here by looping through 'rows'
	var sb strings.Builder
	for _, row := range rows {
		// Join the columns back with commas for display
		sb.WriteString(strings.Join(row, ",")) 
		sb.WriteString("\n")
	}

	return sb.String(), nil
}