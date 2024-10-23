package sqlseeder

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// SeederInterface defines a method for generating a query string from a model
type SeederInterface interface {
	// SeedFromJSON generates SQL INSERT statements from JSON data.
	SeedFromJSON(jsonContent bytes.Buffer, schemaName string, tableName string) (string, error)
	// SeedFromExcel generates SQL INSERT statements from Excel data.
	SeedFromExcel(excelContent bytes.Buffer, schemaName string, tableName string, sheetName string) (string, error)
	GetGenerator() GeneratorInterface
	GetAdapter() AdapterInterface
}

// Seeder implements the SeederInterface and holds a reference to a Seeder
type Seeder struct {
	Generator GeneratorInterface
	Delimiter string
	HashFunc  func(string) string
	Adapter   AdapterInterface
}
type SeederConfig struct {
	OneToManyDelimiter     string
	HashFunc               func(string) string
	ManyToManyRowDelimiter string
	ManyToManyDelimiter    string
}

func NewSeeder(config SeederConfig) SeederInterface {
	delemiter := "|"
	oneToManyDelimiter := "**"
	manyToManyDelimiter := "***"
	if config.ManyToManyDelimiter != "" {
		manyToManyDelimiter = config.ManyToManyDelimiter
	}
	if config.OneToManyDelimiter != "" {
		oneToManyDelimiter = config.OneToManyDelimiter
	}
	if config.ManyToManyRowDelimiter != "" {
		delemiter = config.ManyToManyRowDelimiter
	}
	adapter := NewAdapter(oneToManyDelimiter, manyToManyDelimiter)
	generator := NewGenerator(adapter, delemiter, oneToManyDelimiter, manyToManyDelimiter, config.HashFunc)
	return &Seeder{
		Adapter:   adapter,
		HashFunc:  config.HashFunc,
		Delimiter: delemiter,
		Generator: generator,
	}
}
func (s *Seeder) GetGenerator() GeneratorInterface {
	return s.Generator

}

func (s *Seeder) GetAdapter() AdapterInterface {
	return s.Adapter

}

// SeedFromJSON parses the JSON content from the buffer and generates the SQL
func (s *Seeder) SeedFromJSON(jsonContent bytes.Buffer, schemaName string, tableName string) (string, error) {

	// Unmarshal JSON data into SQLData struct
	var data []map[string]interface{}
	err := json.Unmarshal(jsonContent.Bytes(), &data)
	if err != nil {
		return "", err
	}
	sqlData, err := s.Generator.GenerateTableData(data, schemaName, tableName)
	if err != nil {
		return "", err
	}
	result, err := s.Generator.Generate(*sqlData)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (s *Seeder) SeedFromExcel(excelContent bytes.Buffer, schemaName string, tableName string, sheetName string) (string, error) {
	// ... (code to open Excel file) ...
	f, err := excelize.OpenReader(&excelContent)
	if err != nil {
		return "", fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("failed to close Excel file:", err)
		}
	}()
	// Get the specified sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return "", fmt.Errorf("failed to get sheet '%s': %w", sheetName, err)
	}
	// Check if the sheet has any data
	if len(rows) <= 1 { // Must have at least 2 rows (header + data)
		return "", fmt.Errorf("sheet '%s' has no data", sheetName)
	}

	// Get the header row (column names)
	columns := rows[0]

	fmt.Println(columns, "header")
	// Prepare the data as a slice of maps
	var data []map[string]interface{}
	for _, row := range rows[1:] { // Start from the second row (index 1)
		dataRow := make(map[string]interface{})
		for colIndex, colCell := range row {
			dataRow[columns[colIndex]] = colCell
		}
		data = append(data, dataRow)
	}
	sqlData, err := s.Generator.GenerateTableData(data, schemaName, tableName)
	if err != nil {
		return "", err
	}
	result, err := s.Generator.Generate(*sqlData)
	if err != nil {
		return "", err
	}
	return result, nil
}
