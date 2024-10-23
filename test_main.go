package sqlseeder

import (
	"os"
	"testing"
)

var (
	seeder SeederInterface = NewSeeder(SeederConfig{
		OneToManyDelimiter:     "**",
		ManyToManyRowDelimiter: "|",
		ManyToManyDelimiter:    "***",
	})
	generator GeneratorInterface = seeder.GetGenerator()
	adapter   AdapterInterface   = seeder.GetAdapter()
)

func NewTestAdapter() {

}
func TestMain(m *testing.M) {

	os.Exit(m.Run())
}
