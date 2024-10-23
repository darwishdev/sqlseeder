package sqlseeder

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

// GeneratorInterface defines methods for generating SQL queries and related data.
type GeneratorInterface interface {
	IsLastIndex(index int, a interface{}) bool
	// GenerateTableData generates SQLData from a slice of maps (representing rows)
	// for a given schema and table.
	GenerateTableData(data []map[string]interface{}, schemaName string, tableName string) (*SQLData, error)

	// GenerateRootTableData generates a map representing a single row of data for root columns
	// (columns that are not part of many-to-many relationships).
	GenerateRootTableDataRow(rootColumns []string, row map[string]interface{}, tableName string) (map[string]interface{}, error)
	// GetColumnName extracts the base column name (the part before any delimiters).
	GetColumnName(column string) string

	// Generate generates the SQL insert statements from the provided SQLData.
	Generate(model SQLData) (string, error)

	// GenerateOneToManySubquery generates a subquery for a one-to-many relationship column.
	GenerateOneToManySubquery(columnName string, tableName string, value string) (string, error)
}

type Generator struct {
	TemplatePath        string
	ManyToManyDelimiter string
	OneToManyDelimiter  string
	Delimiter           string
	HashFunc            func(string) string
	Adapter             AdapterInterface
}

func NewGenerator(adapter AdapterInterface, delimiter string, oneToManyDelimiter string, manyToManyDelimiter string, hashFunc func(string) string) GeneratorInterface {
	execPath, err := os.Executable()
	if err != nil {
		panic(err) // Handle the error appropriately
	}

	tmpleatePath := fmt.Sprintf("%s/insert.tmpl", execPath)
	return &Generator{
		TemplatePath:        tmpleatePath, // Path to the SQL template file
		Delimiter:           delimiter,
		ManyToManyDelimiter: manyToManyDelimiter,
		OneToManyDelimiter:  oneToManyDelimiter,
		HashFunc:            hashFunc,

		Adapter: adapter,
	}
}

// GenerateOneToManySubquery generates a subquery for a one-to-many relationship column.
// It takes the column name, table name, and the value to search for.
// If the value is "*", it selects the primary key from the related table without any WHERE clause.
// Otherwise, it generates a subquery to select the primary key where the search key equals the provided value.
func (g *Generator) GenerateOneToManySubquery(columnName string, tableName string, value string) (string, error) {
	relation, err := g.Adapter.ParseOneToMany(columnName, tableName)
	if err != nil {
		return "", err
	}
	if value == "*" {
		return fmt.Sprintf("(SELECT %s FROM %s )", relation.PrimaryKey, relation.Table), nil
	}
	return fmt.Sprintf("(SELECT %s FROM %s WHERE %s = '%s')", relation.PrimaryKey, relation.Table, relation.SearchKey, value), nil

}

func (g *Generator) IsLastIndex(index int, a interface{}) bool {
	return index == g.Adapter.GetLastIndex(a)
}

// GetColumnName extracts the base column name (the part before any delimiters).
func (g *Generator) GetColumnName(column string) string {
	if g.Adapter.IsOneToMany(column) {
		parts := strings.Split(column, g.OneToManyDelimiter)
		return parts[0]
	}
	if g.Adapter.IsHashedColumn(column) {
		parts := strings.Split(column, "#")
		return parts[0]
	}
	return column
}

// GenerateRootTableDataRow generates a map representing a single row of data for root columns.
// It handles one-to-many relationships by generating subqueries.
func (g *Generator) GenerateRootTableDataRow(rootColumns []string, row map[string]interface{}, tableName string) (map[string]interface{}, error) {
	rootRow := make(map[string]interface{})
	for _, rootColumn := range rootColumns {
		var err error
		value := row[rootColumn].(string)
		isOneToMany := g.Adapter.IsOneToMany(rootColumn)
		if isOneToMany {
			value, err = g.GenerateOneToManySubquery(rootColumn, tableName, row[rootColumn].(string))
			if err != nil {
				return nil, err
			}
		}
		rootRow[rootColumn] = value
	}

	return rootRow, nil
}

// GenerateTableData generates SQLData from a slice of maps.
// It handles both root columns and many-to-many relationships.
func (g *Generator) GenerateTableData(data []map[string]interface{}, schemaName string, tableName string) (*SQLData, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	columnsStatemntParts := g.Adapter.SplitColumnsToStatemntParts(data[0])
	fullTableName := g.Adapter.GetFullTableName(schemaName, tableName)
	manyToManyRelations, err := g.Adapter.ParseManyToManyColumns(columnsStatemntParts.ManyToManyColumns, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	rootRows := make([]map[string]interface{}, 0)
	manyToManyRows := make(map[string][]map[string]interface{})
	for _, item := range data {

		rootRow, err := g.GenerateRootTableDataRow(columnsStatemntParts.RootColumns, item, fullTableName)
		if err != nil {
			return nil, err
		}
		rootRows = append(rootRows, rootRow)
		for key, manyToManyColumn := range manyToManyRelations {
			cellValue := item[key].(string)
			cellValueRows := strings.Split(cellValue, g.Delimiter)
			value1, err := g.GenerateOneToManySubquery(manyToManyColumn.Columns[0], manyToManyColumn.Table, item[manyToManyColumn.FirstSearchColumn].(string))
			if err != nil {
				return nil, err
			}
			for _, row := range cellValueRows {
				value2, err := g.GenerateOneToManySubquery(manyToManyColumn.Columns[1], manyToManyColumn.SecondTable, row)
				if err != nil {
					return nil, err
				}

				manyToManyRows[key] = append(manyToManyRows[key], map[string]interface{}{
					manyToManyColumn.Columns[0]: value1,
					manyToManyColumn.Columns[1]: value2,
				})
			}

		}
	}
	sqlData := SQLData{
		Statements: []SQLStatement{
			{
				Table:   tableName,
				Schema:  schemaName,
				Columns: columnsStatemntParts.RootColumns,
				Rows:    rootRows,
			},
		},
	}
	for key, rel := range manyToManyRelations {
		sqlData.Statements = append(sqlData.Statements, SQLStatement{
			Table:   rel.Table,
			Schema:  "",
			Columns: rel.Columns,
			Rows:    manyToManyRows[key],
		})
	}

	return &sqlData, nil
}

// Generate creates the SQL string from the provided SQLData using a template.
func (g *Generator) Generate(data SQLData) (string, error) {
	// Define the template functions.
	funcMap := template.FuncMap{
		"IsLastIndex":      g.IsLastIndex,
		"GetFullTableName": g.Adapter.GetFullTableName,
		"HashFunc":         g.HashFunc,
		"GetColumnName":    g.GetColumnName,
		"IsHashedColumn":   g.Adapter.IsHashedColumn,
	}

	// Read the SQL template from the template path.
	templateContent := `
{{- range $stmt := .Statements }}
INSERT INTO {{ GetFullTableName $stmt.Schema $stmt.Table }} (
  {{- range $index, $column := $stmt.Columns }} 
  {{ GetColumnName $column }} {{- if not (IsLastIndex $index $stmt.Columns) }}, {{ end }}

  {{- end }}
) VALUES
{{- range $rowIndex, $row := $stmt.Rows }}
  (
    {{- range $colIndex, $column := $stmt.Columns }}
      {{- $value := index $row $column }}
        {{- if IsHashedColumn $column }}
          '{{ HashFunc $value }}' {{- if not (IsLastIndex $colIndex $stmt.Columns) }}, {{ end }}
        {{- else }}
          {{ $value }} {{- if not (IsLastIndex $colIndex $stmt.Columns) }}, {{ end }}
        {{- end }}
      {{- end }}
  ) {{- if not (IsLastIndex $rowIndex $stmt.Rows) }}, {{ end }}
{{- end }};
{{- end }}
	`
	// Parse the SQL template.
	tmpl, err := template.New("sql").Funcs(funcMap).Parse(templateContent)
	if err != nil {
		return "", err
	}

	// Use a buffer to capture the generated SQL output.
	var sqlBuffer bytes.Buffer
	err = tmpl.Execute(&sqlBuffer, data)
	if err != nil {
		return "", err
	}

	// Return the generated SQL as a string.
	return sqlBuffer.String(), nil
}
