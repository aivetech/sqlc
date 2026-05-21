package main

import (
	"context"
	"fmt"

	"github.com/sqlc-dev/plugin-sdk-go/codegen"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func main() {
	codegen.Run(generate)
}

func generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	required := map[string]map[string]bool{}

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
				if col.NotNull {
					required[tableName][col.Name] = true
				}
			}
		}
	}

	for _, q := range req.Queries {
		if q.InsertIntoTable == nil {
			continue
		}

		table := q.InsertIntoTable.Name
		reqCols := required[table]

		if len(reqCols) == 0 {
			continue
		}

		provided := map[string]bool{}
		for _, c := range q.Params {
			provided[c.Column.Name] = true
		}

		var missing []string
		for col := range reqCols {
			if !provided[col] {
				missing = append(missing, col)
			}
		}

		if len(missing) > 0 {
			return nil, fmt.Errorf(
				"❌ INSERT into %s missing NOT NULL columns: %v",
				table,
				missing,
			)
		}
	}

	return &plugin.GenerateResponse{
		Files: []*plugin.File{
			{
				Name:     "debug.txt",
				Contents: []byte("Plugin validation passed"),
			},
		},
	}, nil
}
