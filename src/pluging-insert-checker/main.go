package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/codegen"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func main() {
	codegen.Run(generate)
}

type Options struct {
	IgnoreColumns []string `json:"ignore_columns"`
}

func generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	required := map[string]map[string]bool{}

	// f, err := os.OpenFile("./sqlc-plugin.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	return nil, err
	// }
	// defer f.Close()

	var defaultCols = make(map[string][]string)
	if len(req.PluginOptions) > 0 {
		cols, err := parseOptions(req)
		if err != nil {
			return nil, err
		}
		defaultCols = cols
	}

	for _, schema := range req.Catalog.Schemas {
		if schema.Name != "public" {
			continue
		}

		for _, table := range schema.Tables {
			tableName := table.Rel.Name

			if required[tableName] == nil {
				required[tableName] = map[string]bool{}
			}

			for _, col := range table.Columns {
				if col.NotNull && (defaultCols[tableName] == nil || !slices.Contains(defaultCols[tableName], col.Name)) {
					required[tableName][col.Name] = true
				}
			}
		}
	}

	for _, query := range req.Queries {
		if query.InsertIntoTable == nil {
			continue
		}
		insertInfo, err := extractInsertParts(query.Text)
		if err != nil {
			continue
		}

		reqCols := required[insertInfo.Table]
		if len(reqCols) == 0 {
			continue
		}

		provided := map[string]bool{}
		for _, c := range insertInfo.Columns {
			provided[c] = true
		}

		var missing []string
		for col := range reqCols {
			if !provided[col] {
				missing = append(missing, col)
			}
		}

		if len(missing) > 0 {
			return nil, fmt.Errorf(
				"❌ File: %s INSERT into %s missing NOT NULL columns: %v",
				query.Filename,
				insertInfo.Table,
				missing,
			)
		}

	}

	return &plugin.GenerateResponse{}, nil
}

type InsertInfo struct {
	Table   string
	Columns []string
}

// extractInsertParts parses the SQL string to find the table and columns of an INSERT statement
func extractInsertParts(sqlQuery string) (*InsertInfo, error) {
	// Clean up whitespaces and newlines to make matching easier
	normalized := regexp.MustCompile(`\s+`).ReplaceAllString(sqlQuery, " ")

	// Regex breakdown:
	// (?i)                 : Case-insensitive match
	// INSERT\s+INTO\s+     : Matches "INSERT INTO "
	// ([a-zA-Z0-9_]+)      : Capture Group 1: The table name
	// \s*\((.*?)\)         : Capture Group 2: Everything inside the columns parentheses (...)
	// \s*(VALUES|SELECT)   : Matches up to VALUES or SELECT (stopping there)
	re := regexp.MustCompile(`(?i)INSERT\s+INTO\s+([a-zA-Z0-9_]+)\s*\((.*?)\)\s*(VALUES|SELECT)`)

	matches := re.FindStringSubmatch(normalized)
	if len(matches) < 3 {
		return nil, fmt.Errorf("could not parse a valid INSERT INTO statement with columns")
	}

	tableName := matches[1]
	columnsRaw := matches[2]

	// Split columns by comma and trim spaces
	rawCols := strings.Split(columnsRaw, ",")
	var columns []string
	for _, col := range rawCols {
		trimmed := strings.TrimSpace(col)
		if trimmed != "" {
			columns = append(columns, trimmed)
		}
	}

	return &InsertInfo{
		Table:   tableName,
		Columns: columns,
	}, nil
}

func parseOptions(req *plugin.GenerateRequest) (map[string][]string, error) {
	var opts Options
	if err := json.Unmarshal(req.PluginOptions, &opts); err != nil {
		return nil, fmt.Errorf("failed to parse options: %w", err)
	}

	defaults := make(map[string][]string)
	for _, entry := range opts.IgnoreColumns {
		parts := strings.SplitN(entry, ".", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid column format %q, expected table.column", entry)
		}
		if defaults[parts[0]] == nil {
			defaults[parts[0]] = []string{}
		}
		defaults[parts[0]] = append(defaults[parts[0]], parts[1])
	}
	return defaults, nil
}
