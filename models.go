package sqlseeder

// SQLStatement represents an individual SQL statement with multiple rows of data
type SQLStatement struct {
	Schema  string
	Table   string
	Columns []string
	Rows    []map[string]interface{}
}
type ManyToManyRelation struct {
	Table              string
	FirstTable         string
	SecondTable        string
	FirstSearchColumn  string
	SecondSearchColumn string
	Columns            []string
}
type OneToManyRelation struct {
	Table      string
	PrimaryKey string
	SearchKey  string
}
type ColumnsStatemntParts struct {
	RootColumns       []string
	ManyToManyColumns []string
}

// SQLData represents all SQL statements to be executed
type SQLData struct {
	Statements []SQLStatement
}
