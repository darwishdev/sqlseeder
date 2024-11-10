package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/darwishdev/sqlseeder"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(pass string) string {
	password, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(password)

}
func main() {
	// Create an instance of SQLGenerator
	seeder := sqlseeder.NewSeeder(sqlseeder.SeederConfig{
		HashFunc: hashPassword,
	})
	// Create an instance of SeederImpl with the generator

	// Example: Seed from JSON

	// Example data setup
	// data := sqlseeder.SQLData{
	// 	Statements: []sqlseeder.SQLStatement{
	// 		{
	// 			Schema:  "products_schema",
	// 			Table:   "products",
	// 			Columns: []string{"product_name", "category_id**products_schema.categories**category_name"},
	// 			Rows: []map[string]interface{}{
	// 				{"product_name": "Laptop", "category_id**products_schema.categories**category_name": "Electronics"},
	// 				{"product_name": "Phone", "category_id**products_schema.categories**category_name": "Mobile"},
	// 			},
	// 		},
	// 		{
	// 			Schema:  "tags_schema",
	// 			Table:   "tags",
	// 			Columns: []string{"product_id**products_schema.products**product_name", "tag_id**tags_schema.tags**tag_name"},
	// 			Rows: []map[string]interface{}{
	// 				{"product_id**products_schema.products**product_name": "Laptop", "tag_id**tags_schema.tags**tag_name": "Tech"},
	// 			},
	// 		},
	// 	},
	// }
	// resp, err := generator.Generate(data)
	// fmt.Println(resp, err)
	// jsonData := `[
	//
	// 	{
	// 	"product_description"  : "Laptop des" ,
	// 	"product_name"  : "Laptop" ,
	// 	"category_id**products_schema.categories**category_name" : "electronics",
	// 	"tag_id***product_tags***tags***tag_name***product_description" : "electronics|laotio|tag2",
	// 	"image_id***products_schena.product_images***storage_schmea.images***image_path***product_description" : "image/2.png|images/product1.ong"
	// 	},
	// 	{
	// 	"product_description"  : "pc des" ,
	// 	"product_name"  : "PC" ,
	// 	"category_id**products_schema.categories**category_name" : "electronics",
	// 	"tag_id***product_tags***tags***tag_name***product_description" : "electronics|laptops|hardware",
	// 	"image_id***products_schena.product_images***storage_schmea.images***image_path***product_description" : "image/1.png|images/2.png"
	// 	}
	//
	// ]`
	//
	// // Load JSON into a buffer
	// var jsonBuffer bytes.Buffer
	// jsonBuffer.WriteString(jsonData)
	//
	// // Seed from JSON
	// sqlResult, err := seeder.SeedFromJSON(jsonBuffer, "products_schema", "products")
	//
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// } else {
	//
	// 	fmt.Println("Generated SQL:", sqlResult)
	// }
	//
	// // // Example: Seed from Excel (using a buffer with Excel data)
	// // var excelBuffer bytes.Buffer
	// // Populate excelBuffer with Excel data here...
	// // sqlResult, err = seeder.SeedFromExcel(excelBuffer)
	// // Handle the result...
	// Test SeedFromExcel
	excelFile, err := os.ReadFile("accounts.xlsx") // Replace with your Excel file path
	if err != nil {
		log.Fatal().Err(err).Msg("cano open the file")
	}

	excelString, err := seeder.SeedFromExcel(*bytes.NewBuffer(excelFile), "accounts_schema", "roles", "roles", map[string]string{})
	if err != nil {
		log.Fatal().Err(err).Msg("cano seed from excel")
	}
	fmt.Println(excelString)
}
