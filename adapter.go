package sqlseeder

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/tangzero/inflector"
)

// AdapterInterface defines methods for adapting data and generating SQL-related information.
type AdapterInterface interface {
	// GetLastIndex returns the last index of the given data structure (slice or array).
	GetLastIndex(a interface{}) int

	// WrapWithSingleQoute wraps a value in single quotes (e.g., 'value').
	WrapWithSingleQoute(value string) string

	// ParseManyToManyColumns parses a list of many-to-many column names and returns
	// a map of ManyToManyRelation structs.
	// This map will hold the column name as a key and an object of ManyToManyRelation as the value.
	ParseManyToManyColumns(columns []string, schemaName string, tableName string) (map[string]ManyToManyRelation, error)

	// ParseManyToMany parses a single many-to-many column name and returns a
	// ManyToManyRelation struct.
	ParseManyToMany(columnName string, schemaName string, tableName string) (ManyToManyRelation, error)

	// IsOneToMany checks if a column represents a one-to-many relationship.
	IsOneToMany(columnName string) bool
	// IsHashedColumn checks if a column represents a password so it should be hashed
	IsHashedColumn(columnName string) bool
	// GetFullTableName returns the full table name with the schema (if provided).
	GetFullTableName(schemaName string, tableName string) string

	// GetPrimaryKeyFromTableName generates the primary key column name from a table name.
	GetPrimaryKeyFromTableName(tableName string) string

	// SplitColumnsToStatemntParts splits the columns of a row into root columns and
	// many-to-many columns.
	SplitColumnsToStatemntParts(row map[string]interface{}) ColumnsStatemntParts

	// ParseOneToMany parses a one-to-many column name and returns an
	// OneToManyRelation struct.
	ParseOneToMany(columnName string, tableName string) (OneToManyRelation, error)
}

// Adapter implements the AdapterInterface.
type Adapter struct {
	OneToManyDelimiter  string
	ManyToManyDelimiter string
}

// NewAdapter creates a new Adapter with the specified delimiters.
func NewAdapter(oneToManyDelimiter string, manyToManyDelimiter string) AdapterInterface {
	return &Adapter{
		OneToManyDelimiter:  oneToManyDelimiter,
		ManyToManyDelimiter: manyToManyDelimiter,
	}
}

// GetLastIndex returns the last index of the given data structure.
func (a *Adapter) GetLastIndex(data interface{}) int {
	return reflect.ValueOf(data).Len() - 1
}

// GetFullTableName returns the full table name with the schema (if provided).
func (a *Adapter) GetFullTableName(schemaName string, tableName string) string {
	name := tableName

	if schemaName != "" {
		name = fmt.Sprintf("%s.%s", schemaName, tableName)
	}
	return name
}

// IsOneToMany checks if a column represents a one-to-many relationship.
func (a *Adapter) IsHashedColumn(columnName string) bool {
	return strings.Contains(columnName, "#") && len(strings.Split(columnName, "#")) == 2
}

// IsOneToMany checks if a column represents a one-to-many relationship.
func (a *Adapter) IsOneToMany(columnName string) bool {
	return strings.Contains(columnName, a.OneToManyDelimiter) && !strings.Contains(columnName, a.ManyToManyDelimiter)
}

// WrapWithSingleQoute wraps a value in single quotes.
func (a *Adapter) WrapWithSingleQoute(value string) string {
	if value == "" || value == "NULL" || value == "null" {
		return "NULL"
	}
	if strings.Contains(value, "(") {
		return value
	}
	return fmt.Sprintf("'%s'", value)
}

// ParseManyToMany parses a many-to-many relationship column name.
//
// Formula: <joining_table_primary_key><ManyToManyDelimiter><joining_table_name><ManyToManyDelimiter><second_table_name><ManyToManyDelimiter><second_table_search_column><ManyToManyDelimiter><first_table_search_column>
// Example: tag_id***product_tags***tags***tag_name***product_product_name
// Should return:
//
//	ManyToManyRelation{
//	  Table:              "product_tags",
//	  FirstTable:         "products", // Assuming the current table is "products"
//	  SecondTable:        "tags",
//	  FirstSearchColumn:  "product_name",
//	  SecondSearchColumn: "tag_name",
//	  Columns:            ["product_id**products**product_name", "tag_id**tags**tag_name"],
//	}
func (a *Adapter) ParseManyToMany(columnName string, schemaName string, tableName string) (ManyToManyRelation, error) {
	parts := strings.Split(columnName, a.ManyToManyDelimiter)
	response := ManyToManyRelation{}
	if len(parts) != 5 {
		return response, fmt.Errorf("not valid many to many column name: %s", columnName)
	}
	fullTableName := a.GetFullTableName(schemaName, tableName)
	firstColumn := fmt.Sprintf("%s%s%s%s%s", a.GetPrimaryKeyFromTableName(tableName), a.OneToManyDelimiter, fullTableName, a.OneToManyDelimiter, parts[4])
	secondColumn := fmt.Sprintf("%s%s%s%s%s", parts[0], a.OneToManyDelimiter, parts[2], a.OneToManyDelimiter, parts[3])
	result := []string{firstColumn, secondColumn}
	response = ManyToManyRelation{
		Table:              parts[1],
		FirstTable:         tableName,
		SecondTable:        parts[2],
		FirstSearchColumn:  parts[4],
		SecondSearchColumn: parts[3],
		Columns:            result,
	}
	return response, nil
}

// ParseManyToManyColumns parses a list of many-to-many column names.
func (a *Adapter) ParseManyToManyColumns(columns []string, schemaName string, tableName string) (map[string]ManyToManyRelation, error) {
	result := make(map[string]ManyToManyRelation)
	for _, column := range columns {
		row, err := a.ParseManyToMany(column, schemaName, tableName)
		if err != nil {
			return nil, err
		}
		result[column] = row
	}
	return result, nil

}

// ParseOneToMany parses a one-to-many relationship column name.
//
// Formula: <primary_key_column><OneToManyDelimiter><table_name><OneToManyDelimiter><search_key_column>
// Example: category_id**categories**category_name
// Should return:
//
//	OneToManyRelation{
//	  Table:      "categories",
//	  PrimaryKey: "category_id",
//	  SearchKey:  "category_name",
//	}
func (a *Adapter) ParseOneToMany(columnName string, tableName string) (OneToManyRelation, error) {
	parts := strings.Split(columnName, a.OneToManyDelimiter)
	response := OneToManyRelation{}
	if len(parts) != 3 {
		return response, fmt.Errorf("not valid one to many column name: %s", columnName)
	}
	response = OneToManyRelation{
		Table:      parts[1],
		PrimaryKey: parts[0],
		SearchKey:  parts[2],
	}
	return response, nil
}

// SplitColumnsToStatemntParts splits the columns of a row into root columns and many-to-many columns.
func (a *Adapter) SplitColumnsToStatemntParts(row map[string]interface{}) ColumnsStatemntParts {
	manyToManyColumns := []string{}
	rootColumns := []string{}
	for key := range row {
		if strings.Contains(key, a.ManyToManyDelimiter) {
			manyToManyColumns = append(manyToManyColumns, key)
			continue
		}
		rootColumns = append(rootColumns, key)
	}
	return ColumnsStatemntParts{
		RootColumns:       rootColumns,
		ManyToManyColumns: manyToManyColumns,
	}
}

// GetPrimaryKeyFromTableName generates the primary key column name from a table name
// using the formula (singular name)_id.
//
// Examples:
//   - table: "products"  =>  "product_id"
//   - table: "categories"  =>  "category_id"
func (a *Adapter) GetPrimaryKeyFromTableName(tableName string) string {
	singluraizedName := inflector.Singularize(tableName)
	return fmt.Sprintf("%s_id", singluraizedName)
}
