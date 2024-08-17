package gen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"

	"github.com/bryce-gif/gen/internal/model"
)

// GenerateMode generate mode
type GenerateMode uint

const (
	// WithDefaultQuery create default query in generated code
	WithDefaultQuery GenerateMode = 1 << iota

	// WithoutContext generate code without context constrain
	WithoutContext

	// WithQueryInterface generate code with exported interface object
	WithQueryInterface
)

// Config generator's basic configuration
type Config struct {
	db *gorm.DB // db connection

	OutPath      string // 查询代码路径
	OutFile      string // 查询代码文件名，默认: gen.go
	ModelPkgPath string // 生成的模型代码的包名
	WithUnitTest bool   // 为查询代码生成单元测试

	// 生成模型全局配置
	FieldNullable     bool // 当字段可为空时生成指针
	FieldCoverable    bool // 当字段有默认值时生成指针，以修复无法分配零值的问题: https://gorm.io/docs/create.html#Default-Values
	FieldSignable     bool // 检测整数字段的无符号类型，调整生成的数据类型
	FieldWithIndexTag bool // 使用 gorm 索引标签生成
	FieldWithTypeTag  bool // 使用 gorm 列类型标签生成

	Mode GenerateMode // generate mode

	queryPkgName   string // generated query code's package name
	modelPkgPath   string // model pkg path in target project
	dbNameOpts     []model.SchemaNameOpt
	importPkgPaths []string

	// name strategy for syncing table from db
	tableNameNS func(tableName string) (targetTableName string)
	modelNameNS func(tableName string) (modelName string)
	fileNameNS  func(tableName string) (fileName string)

	dataTypeMap    map[string]func(columnType gorm.ColumnType) (dataType string)
	fieldJSONTagNS func(columnName string) (tagContent string)

	modelOpts []ModelOpt
}

// WithOpts set global  model options
func (cfg *Config) WithOpts(opts ...ModelOpt) {
	if cfg.modelOpts == nil {
		cfg.modelOpts = opts
	} else {
		cfg.modelOpts = append(cfg.modelOpts, opts...)
	}
}

// WithDbNameOpts set get database name function
func (cfg *Config) WithDbNameOpts(opts ...model.SchemaNameOpt) {
	if cfg.dbNameOpts == nil {
		cfg.dbNameOpts = opts
	} else {
		cfg.dbNameOpts = append(cfg.dbNameOpts, opts...)
	}
}

// WithTableNameStrategy specify table name naming strategy, only work when syncing table from db
func (cfg *Config) WithTableNameStrategy(ns func(tableName string) (targetTableName string)) {
	cfg.tableNameNS = ns
}

// WithModelNameStrategy specify model struct name naming strategy, only work when syncing table from db
func (cfg *Config) WithModelNameStrategy(ns func(tableName string) (modelName string)) {
	cfg.modelNameNS = ns
}

// WithFileNameStrategy specify file name naming strategy, only work when syncing table from db
func (cfg *Config) WithFileNameStrategy(ns func(tableName string) (fileName string)) {
	cfg.fileNameNS = ns
}

// WithDataTypeMap specify data type mapping relationship, only work when syncing table from db
func (cfg *Config) WithDataTypeMap(newMap map[string]func(columnType gorm.ColumnType) (dataType string)) {
	cfg.dataTypeMap = newMap
}

// WithJSONTagNameStrategy specify json tag naming strategy
func (cfg *Config) WithJSONTagNameStrategy(ns func(columnName string) (tagContent string)) {
	cfg.fieldJSONTagNS = ns
}

// WithImportPkgPath specify import package path
func (cfg *Config) WithImportPkgPath(paths ...string) {
	for i, path := range paths {
		path = strings.TrimSpace(path)
		if len(path) > 0 && path[0] != '"' && path[len(path)-1] != '"' { // without quote
			path = `"` + path + `"`
		}
		paths[i] = path
	}
	cfg.importPkgPaths = append(cfg.importPkgPaths, paths...)
}

// Revise format path and db
func (cfg *Config) Revise() (err error) {
	if strings.TrimSpace(cfg.ModelPkgPath) == "" {
		cfg.ModelPkgPath = model.DefaultModelPkg
	}

	cfg.OutPath, err = filepath.Abs(cfg.OutPath)
	if err != nil {
		return fmt.Errorf("outpath is invalid: %w", err)
	}
	if cfg.OutPath == "" {
		cfg.OutPath = fmt.Sprintf(".%squery%s", string(os.PathSeparator), string(os.PathSeparator))
	}
	if cfg.OutFile == "" {
		cfg.OutFile = filepath.Join(cfg.OutPath, "gen.go")
	} else if !strings.Contains(cfg.OutFile, string(os.PathSeparator)) {
		cfg.OutFile = filepath.Join(cfg.OutPath, cfg.OutFile)
	}
	cfg.queryPkgName = filepath.Base(cfg.OutPath)

	if cfg.db == nil {
		cfg.db, _ = gorm.Open(tests.DummyDialector{})
	}

	return nil
}

func (cfg *Config) judgeMode(mode GenerateMode) bool { return cfg.Mode&mode != 0 }
