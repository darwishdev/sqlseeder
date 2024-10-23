# SQL Seeder

This Go package provides a tool for generating SQL INSERT statements from JSON or Excel data. It supports various relationships between tables (one-to-many and many-to-many) and allows you to customize the delimiters used in your data.

## Features

* **Seed from JSON or Excel:** Generate SQL from structured data in either JSON or Excel format.
* **Relationship Support:** Handles one-to-many and many-to-many relationships between tables.
* **Customizable Delimiters:** Configure the delimiters used in your data for flexible parsing.
* **Templating:** Uses Go templates to generate the SQL statements, allowing for customization.

## Installation

```bash
go get github.com/darwishdev/sqlseeder
````


## Usage

### 1\. Define your data

**JSON:**

```json
[
  {
    "id": "1",
    "name": "Product 1",
    "category_id**categories**category_name": "Electronics",
    "tag_id***product_tags***tags***tag_name***product_name": "tag1|tag2"
  },
  {
    "id": "2",
    "name": "Product 2",
    "category_id**categories**category_name": "Books",
    "tag_id***product_tags***tags***tag_name***product_name": "tag3"
  }
]
```

**Excel:**

| id | name       | category\_id**categories**category\_name | tag\_id***product\_tags***tags***tag\_name***product\_name |
|----|------------|---------------------------------------|---------------------------------------------------------|
| 1  | Product 1  | Electronics                            | tag1|tag2                                              |
| 2  | Product 2  | Books                                 | tag3                                                   |

### 2\. Create a Seeder

```go
import "github.com/your-module-path/sqlseeder"

// Configure delimiters (optional)
config := sqlseeder.SeederConfig{
  OneToManyDelimiter:     "**",
  ManyToManyRowDelimiter: "|",
  ManyToManyDelimiter:    "***",
}

seeder := sqlseeder.NewSeeder(config)
```

### 3\. Generate SQL

**From JSON:**

```go
jsonContent := // Load your JSON data
sqlString, err := seeder.SeedFromJSON(bytes.NewBuffer(jsonContent), "your_schema", "your_table")
if err != nil {
  // Handle error
}
fmt.Println(sqlString)
```

**From Excel:**

```go
excelContent := // Load your Excel data
sqlString, err := seeder.SeedFromExcel(bytes.NewBuffer(excelContent), "your_schema", "your_table", "Sheet1")
if err != nil {
  // Handle error
}
fmt.Println(sqlString)
```

## Column Name Formulas

  * **One-to-many:** `<primary_key_column><OneToManyDelimiter><table_name><OneToManyDelimiter><search_key_column>`
  * **Many-to-many:** `<joining_table_primary_key><ManyToManyDelimiter><joining_table_name><ManyToManyDelimiter><second_table_name><ManyToManyDelimiter><second_table_search_column><ManyToManyDelimiter><first_table_search_column>`

## Contributing

Contributions are welcome\! Feel free to open issues or submit pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](https://www.google.com/url?sa=E&source=gmail&q=LICENSE) file for details.

``
