package sqlseeder

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdapter_GetLastIndex(t *testing.T) {
	data := []string{"a", "b", "c"}
	fmt.Println("adapter", adapter)
	lastIndex := adapter.GetLastIndex(data)
	if lastIndex != 2 {
		t.Errorf("Expected last index to be 2, but got %d", lastIndex)
	}
}
func TestAdapter_GetFullTableName(t *testing.T) {
	// Test with schema name
	fullTableName := adapter.GetFullTableName("my_schema", "my_table")
	expected := "my_schema.my_table"
	if fullTableName != expected {
		t.Errorf("Expected full table name to be '%s', but got '%s'", expected, fullTableName)
	}

	// Test without schema name
	fullTableName = adapter.GetFullTableName("", "my_table")
	expected = "my_table"
	if fullTableName != expected {
		t.Errorf("Expected full table name to be '%s', but got '%s'", expected, fullTableName)
	}
}

func TestAdapter_IsOneToMany(t *testing.T) {

	// Test one-to-many column name
	if !adapter.IsOneToMany("category_id**categories**category_name") {
		t.Error("Expected 'category_id**categories**category_name' to be a one-to-many column")
	}

	// Test many-to-many column name
	if adapter.IsOneToMany("tag_id***product_tags***tags***tag_name***product_name") {
		t.Error("Expected 'tag_id***product_tags***tags***tag_name***product_name' to NOT be a one-to-many column")
	}
}

func TestAdapter_WrapWithSingleQoute(t *testing.T) {
	wrappedValue := adapter.WrapWithSingleQoute("test value")
	expected := "'test value'"
	if wrappedValue != expected {
		t.Errorf("Expected wrapped value to be '%s', but got '%s'", expected, wrappedValue)
	}
}

func TestAdapter_ParseManyToMany(t *testing.T) {
	relation, err := adapter.ParseManyToMany("tag_id***product_tags***tags***tag_name***product_name", "public", "products")
	if err != nil {
		t.Error(err)
	}

	expected := ManyToManyRelation{
		Table:              "product_tags",
		FirstTable:         "products",
		SecondTable:        "tags",
		FirstSearchColumn:  "product_name",
		SecondSearchColumn: "tag_name",
		Columns:            []string{"product_id**public.products**product_name", "tag_id**tags**tag_name"},
	}
	require.Equal(t, relation, expected)
}

func TestAdapter_ParseOneToMany(t *testing.T) {
	relation, err := adapter.ParseOneToMany("category_id**categories**category_name", "products")
	if err != nil {
		t.Error(err)
	}

	expected := OneToManyRelation{
		Table:      "categories",
		PrimaryKey: "category_id",
		SearchKey:  "category_name",
	}

	if relation != expected {
		t.Errorf("Expected OneToManyRelation to be %+v, but got %+v", expected, relation)
	}
}
func TestAdapter_SplitColumnsToStatemntParts(t *testing.T) {
	row := map[string]interface{}{
		"id":          1,
		"name":        "Product 1",
		"category_id": "1",
		"tag_id***product_tags***tags***tag_name***product_name": "tag1|tag2",
	}

	parts := adapter.SplitColumnsToStatemntParts(row)

	expectedRootColumns := []string{"id", "name", "category_id"}
	expectedManyToManyColumns := []string{"tag_id***product_tags***tags***tag_name***product_name"}
	sort.Strings(parts.RootColumns)
	sort.Strings(expectedRootColumns)
	sort.Strings(parts.ManyToManyColumns)
	sort.Strings(expectedManyToManyColumns)
	require.Equal(t, parts.RootColumns, expectedRootColumns)
	require.Equal(t, parts.ManyToManyColumns, expectedManyToManyColumns)
}

func TestAdapter_GetPrimaryKeyFromTableName(t *testing.T) {

	primaryKey := adapter.GetPrimaryKeyFromTableName("products")
	expected := "product_id"
	if primaryKey != expected {
		t.Errorf("Expected primary key to be '%s', but got '%s'", expected, primaryKey)
	}

	primaryKey = adapter.GetPrimaryKeyFromTableName("categories")
	expected = "category_id"
	if primaryKey != expected {
		t.Errorf("Expected primary key to be '%s', but got '%s'", expected, primaryKey)
	}
}
