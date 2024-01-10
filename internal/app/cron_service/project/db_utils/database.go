package db_utils

import (
	"encoding/json"
	"github.com/glebarez/sqlite"
	"github.com/invopop/jsonschema"
	jsonschema2 "github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"gorm.io/gorm"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var logger = comfy_log.New("[cron_service database]")

type Database struct {
	Tables []Table `json:"tables" jsonschema:"title=tables, description=tables of database"`
}

type Table struct {
	Name    string   `json:"name" jsonschema:"title=table name, description=name of table, required=true"`
	Columns []Column `json:"columns" jsonschema:"title=columns, description=columns of table, required=true"`
}

type Column struct {
	Name string `json:"name" jsonschema:"title=column name, description=name of column, required=true"`
	Type string `json:"type" jsonschema:"title=column data type, enum=int, enum=uint, enum=time, enum=float, enum=boolean, enum=string, required=true"`
}

var DatabaseSchema string

func init() {
	reflector := jsonschema.Reflector{
		Anonymous: true,
	}
	schema := reflector.Reflect(&Database{})
	bytes, err := json.MarshalIndent(schema, "", " ")
	if err != nil {
		logger.Fatal("generate database schema failed, %w", err)
	}

	DatabaseSchema = string(bytes)
}

func OpenOrCreate(dsn string) (*gorm.DB, error) {
	err := os.MkdirAll(filepath.Dir(dsn), os.ModePerm)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB, database Database) error {
	for _, table := range database.Tables {
		if err := db.Table(table.Name).AutoMigrate(TableStructToGormModel(table)); err != nil {
			return err
		}
	}

	return nil
}

func ValidateDatabaseJson(r io.Reader) error {
	compiler := jsonschema2.NewCompiler()
	if err := compiler.AddResource("database_schema.json", strings.NewReader(DatabaseSchema)); err != nil {
		return err
	}
	sch, err := compiler.Compile("database_schema.json")
	if err != nil {
		return err
	}

	var databaseStruct map[string]any
	if err := json.NewDecoder(r).Decode(&databaseStruct); err != nil {
		return err
	}
	delete(databaseStruct, "$schema")

	if err := sch.Validate(databaseStruct); err != nil {
		return err
	}

	return nil
}
