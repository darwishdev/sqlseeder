package sqlseeder

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerator_IsLastIndex(t *testing.T) {

	// Test cases
	testCases := []struct {
		index    int
		data     interface{}
		expected bool
	}{
		{0, []string{"a"}, true},
		{1, []string{"a", "b"}, true},
		{0, []string{"a", "b"}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Index %d", tc.index), func(t *testing.T) {
			isLast := generator.IsLastIndex(tc.index, tc.data)
			if isLast != tc.expected {
				t.Errorf("Expected IsLastIndex to be %v, but got %v", tc.expected, isLast)
			}
		})
	}
}

func TestGenerator_GetColumnName(t *testing.T) {

	// Test one-to-many column
	columnName := generator.GetColumnName("category_id**categories**category_name")
	expected := "category_id"
	if columnName != expected {
		t.Errorf("Expected column name to be '%s', but got '%s'", expected, columnName)
	}

	// Test regular column
	columnName = generator.GetColumnName("name")
	expected = "name"
	if columnName != expected {
		t.Errorf("Expected column name to be '%s', but got '%s'", expected, columnName)
	}
}

func TestGenerator_GenerateRootTableDataRow(t *testing.T) {

	row := map[string]interface{}{
		"id":                                     "1",
		"name":                                   "Product 1",
		"category_id**categories**category_name": "Electronics",
	}

	rootColumns := []string{"id", "name", "category_id**categories**category_name"}

	result, err := generator.GenerateRootTableDataRow(rootColumns, row, "products")
	if err != nil {
		t.Error(err)
	}

	expected := map[string]interface{}{
		"id":                                     "'1'",
		"name":                                   "'Product 1'",
		"category_id**categories**category_name": "(SELECT category_id FROM categories WHERE category_name = 'Electronics')",
	}
	require.Equal(t, result, expected)
}

func TestGenerator_GenerateTableData(t *testing.T) {
	// Sample data
	data := []map[string]interface{}{
		{
			"id":                                     "1",
			"product_name":                           "Product 1",
			"category_id**categories**category_name": "Electronics",
			"tag_id***product_tags***tags***tag_name***product_name": "tag1|tag2",
		},
		{
			"id":                                     "2",
			"product_name":                           "Product 2",
			"category_id**categories**category_name": "Books",
			"tag_id***product_tags***tags***tag_name***product_name": "tag3",
		},
	}

	// Expected SQLData
	expected := &SQLData{
		Statements: []SQLStatement{
			{
				Table:   "products",
				Schema:  "public",
				Columns: []string{"id", "product_name", "category_id**categories**category_name"},
				Rows: []map[string]interface{}{
					{
						"id":                                     "'1'",
						"product_name":                           "'Product 1'",
						"category_id**categories**category_name": "(SELECT category_id FROM categories WHERE category_name = 'Electronics')",
					},
					{
						"id":                                     "'2'",
						"product_name":                           "'Product 2'",
						"category_id**categories**category_name": "(SELECT category_id FROM categories WHERE category_name = 'Books')",
					},
				},
			},
			{
				Table:   "product_tags",
				Schema:  "",
				Columns: []string{"product_id**public.products**product_name", "tag_id**tags**tag_name"},
				Rows: []map[string]interface{}{
					{
						"product_id**public.products**product_name": "(SELECT product_id FROM public.products WHERE product_name = 'Product 1')",
						"tag_id**tags**tag_name":                    "(SELECT tag_id FROM tags WHERE tag_name = 'tag1')",
					},
					{
						"product_id**public.products**product_name": "(SELECT product_id FROM public.products WHERE product_name = 'Product 1')",
						"tag_id**tags**tag_name":                    "(SELECT tag_id FROM tags WHERE tag_name = 'tag2')",
					},
					{
						"product_id**public.products**product_name": "(SELECT product_id FROM public.products WHERE product_name = 'Product 2')",
						"tag_id**tags**tag_name":                    "(SELECT tag_id FROM tags WHERE tag_name = 'tag3')",
					},
				},
			},
		},
	}

	// Generate SQLData
	result, err := generator.GenerateTableData(data, "public", "products")
	require.NoError(t, err)

	// Compare results
	require.Equal(t, len(expected.Statements), len(result.Statements))

	for i := range expected.Statements {
		// Sort the Rows slice to ignore order when comparing
		sort.Slice(expected.Statements[i].Rows, func(j, k int) bool {
			return fmt.Sprint(expected.Statements[i].Rows[j]) < fmt.Sprint(expected.Statements[i].Rows[k])
		})
		sort.Slice(result.Statements[i].Rows, func(j, k int) bool {
			return fmt.Sprint(result.Statements[i].Rows[j]) < fmt.Sprint(result.Statements[i].Rows[k])
		})

		require.Equal(t, expected.Statements[i].Table, result.Statements[i].Table)
		require.Equal(t, expected.Statements[i].Schema, result.Statements[i].Schema)
		require.Equal(t, expected.Statements[i].Columns, result.Statements[i].Columns)
		require.Equal(t, expected.Statements[i].Rows, result.Statements[i].Rows)
	}
}

func TestGenerator_Generate(t *testing.T) {
	// ... (This test requires reading from a template file and generating SQL, so it's also best to implement based on your specific template and logic) ...
}
