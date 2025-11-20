package sqlseeder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// DataLoader interface for loading data from different sources
type DataLoader interface {
	Load() ([]map[string]interface{}, error)
}

// JsonLoader loads data from JSON
type JsonLoader struct {
	Content bytes.Buffer
}

// ExcelLoader loads data from Excel
type ExcelLoader struct {
	Content       bytes.Buffer
	SheetName     string
	ColumnsMapper map[string]string
}

// CSVLoader loads data from CSV (example for future extensibility)
type CSVLoader struct {
	Content       bytes.Buffer
	ColumnsMapper map[string]string
}

// SeederConfig contains all seeding configuration
type SeederConfig struct {
	Loader       DataLoader
	SchemaName   string // optional - required for table-based insert
	TableName    string // optional - required for table-based insert
	FunctionName string // optional - if provided, uses function-based import
}

// Load implementation for JsonLoader
func (j JsonLoader) Load() ([]map[string]interface{}, error) {
	var data []map[string]interface{}
	if err := json.Unmarshal(j.Content.Bytes(), &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return data, nil
}

// Load implementation for ExcelLoader
func (e ExcelLoader) Load() ([]map[string]interface{}, error) {
	f, err := excelize.OpenReader(&e.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("failed to close Excel file:", err)
		}
	}()

	rows, err := f.GetRows(e.SheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get sheet '%s': %w", e.SheetName, err)
	}

	if len(rows) <= 1 {
		return nil, fmt.Errorf("sheet '%s' has no data", e.SheetName)
	}

	columns := rows[0]
	var data []map[string]interface{}

	for _, row := range rows[1:] {
		dataRow := make(map[string]interface{})
		for colIndex, colCell := range row {
			if colIndex >= len(columns) {
				break
			}
			currentColumnName := strings.ToLower(strings.TrimSpace(columns[colIndex]))
			mappedColumnName := currentColumnName
			if e.ColumnsMapper != nil {
				if mapped, ok := e.ColumnsMapper[currentColumnName]; ok {
					mappedColumnName = mapped
				}
			}
			dataRow[mappedColumnName] = colCell
		}
		data = append(data, dataRow)
	}

	return data, nil
}

// SeederInterface defines methods for generating SQL from various sources
type SeederInterface interface {
	// Seed is the unified method that accepts a SeederConfig
	Seed(config SeederConfig) (string, error)

	// Legacy methods - kept for backward compatibility
	// SeedFromJSON(jsonContent bytes.Buffer, schemaName string, tableName string) (string, error)
	// SeedFromExcel(excelContent bytes.Buffer, schemaName string, tableName string, sheetName string, columnsMapper map[string]string) (string, error)

	GetGenerator() GeneratorInterface
	GetAdapter() AdapterInterface
}

// Seeder implements the SeederInterface
type Seeder struct {
	Generator      GeneratorInterface
	Delimiter      string
	ArrayDelimiter string
	Embed          func(ctx context.Context, text string, model ...string) ([][]float32, error)
	EmbedBulk      func(ctx context.Context, text []string, model ...string) ([][][]float32, error)
	HashFunc       func(string) string
	Adapter        AdapterInterface
}

type SeederConfigInit struct {
	OneToManyDelimiter     string
	HashFunc               func(string) string
	Embed                  func(ctx context.Context, text string, model ...string) ([][]float32, error)
	EmbedBulk              func(ctx context.Context, text []string, model ...string) ([][][]float32, error)
	ColumnsMapper          map[string]string
	ManyToManyRowDelimiter string
	ArrayDelimiter         string
	ManyToManyDelimiter    string
}

func NewSeeder(config SeederConfigInit) SeederInterface {
	delimiter := "|"
	if config.ArrayDelimiter == "" {
		config.ArrayDelimiter = ","
	}
	oneToManyDelimiter := "**"
	manyToManyDelimiter := "***"
	if config.ManyToManyDelimiter != "" {
		manyToManyDelimiter = config.ManyToManyDelimiter
	}
	if config.OneToManyDelimiter != "" {
		oneToManyDelimiter = config.OneToManyDelimiter
	}
	if config.ManyToManyRowDelimiter != "" {
		delimiter = config.ManyToManyRowDelimiter
	}
	adapter := NewAdapter(oneToManyDelimiter, manyToManyDelimiter)
	generator := NewGenerator(adapter, config.ColumnsMapper, delimiter, config.ArrayDelimiter, oneToManyDelimiter, manyToManyDelimiter, config.HashFunc)
	return &Seeder{
		Adapter:        adapter,
		Embed:          config.Embed,
		EmbedBulk:      config.EmbedBulk,
		HashFunc:       config.HashFunc,
		Delimiter:      delimiter,
		ArrayDelimiter: config.ArrayDelimiter,
		Generator:      generator,
	}
}

func (s *Seeder) GetGenerator() GeneratorInterface {
	return s.Generator
}

func (s *Seeder) GetAdapter() AdapterInterface {
	return s.Adapter
}

// Seed is the unified method
func (s *Seeder) Seed(config SeederConfig) (string, error) {
	// Load data using the provided loader
	data, err := config.Loader.Load()
	if err != nil {
		return "", err
	}

	// If FunctionName is provided, use function-based import
	if config.FunctionName != "" {
		return s.generateFunctionCall(data, config.FunctionName)
	}

	// Otherwise, use table-based import
	if config.SchemaName == "" || config.TableName == "" {
		return "", fmt.Errorf("SchemaName and TableName are required when FunctionName is not provided")
	}

	sqlData, err := s.Generator.GenerateTableData(data, config.SchemaName, config.TableName)
	if err != nil {
		return "", err
	}

	return s.Generator.Generate(*sqlData)
}

// generateFunctionCall generates a SELECT statement calling a SQL function with JSON data
func (s *Seeder) generateFunctionCall(data []map[string]interface{}, functionName string) (string, error) {
	// Marshal data back to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	// Escape single quotes in JSON for SQL
	jsonStr := string(jsonBytes)
	jsonStr = strings.ReplaceAll(jsonStr, "'", "''")

	// Generate SELECT statement
	sql := fmt.Sprintf("SELECT %s('%s'::JSONB);", functionName, jsonStr)
	return sql, nil
}

//
// // Legacy method - SeedFromJSON
// func (s *Seeder) SeedFromJSON(jsonContent bytes.Buffer, schemaName string, tableName string) (string, error) {
// 	return s.Seed(SeederConfig{
// 		Loader:     JsonLoader{Content: jsonContent},
// 		SchemaName: schemaName,
// 		TableName:  tableName,
// 	})
// }
//
// // Legacy method - SeedFromExcel
// func (s *Seeder) SeedFromExcel(excelContent bytes.Buffer, schemaName string, tableName string, sheetName string, columnsMapper map[string]string) (string, error) {
// 	return s.Seed(SeederConfig{
// 		Loader: ExcelLoader{
// 			Content:       excelContent,
// 			SheetName:     sheetName,
// 			ColumnsMapper: columnsMapper,
// 		},
// 		SchemaName: schemaName,
// 		TableName:  tableName,
// 	})
// }
