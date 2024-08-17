package tests

import (
	"log"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bryce-gif/gen"
)

var (
	dsn = "root:123456@tcp(127.0.0.1:3306)/dev?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
)

func InitDb() *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func TestGenModel(t *testing.T) {
	g := gen.NewGenerator(gen.Config{
		OutPath:           "./gen/query",
		ModelPkgPath:      "./gen/",
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:     true,
		FieldWithIndexTag: false,
		FieldWithTypeTag:  false,
	})

	g.UseDB(InitDb())

	g.WithDataTypeMap(map[string]func(gorm.ColumnType) (dataType string){
		"int": func(columnType gorm.ColumnType) (dataType string) {
			return "int64"
		},
		"bigint": func(columnType gorm.ColumnType) (dataType string) {
			return "int64"
		},
		"tinyint": func(columnType gorm.ColumnType) (dataType string) {
			return "int64"
		},
	})

	g.ApplyBasic(g.GenerateModelAs("users", "Users"))

	g.Execute()
}
