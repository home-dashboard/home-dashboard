package db_utils

import (
	"encoding/json"
	"github.com/glebarez/sqlite"
	"github.com/invopop/jsonschema"
	jsonschema2 "github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"gorm.io/gorm"
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

var databaseSchema *jsonschema2.Schema

func init() {
	schemaString, err := generateJsonSchema(Database{})
	if err != nil {
		logger.Fatal("generate database schema failed, %w", err)
	}

	databaseSchema, err = generateJsonSchemaValidator(schemaString)
	if err != nil {
		logger.Fatal("generate database schema validator failed, %w", err)
	}
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

func ValidateDatabaseJson(str string) error {
	var databaseStruct map[string]any
	if err := json.Unmarshal([]byte(str), &databaseStruct); err != nil {
		return err
	}
	delete(databaseStruct, "$schema")

	if err := databaseSchema.Validate(databaseStruct); err != nil {
		return err
	}

	return nil
}

func ValidateJson(str string, schemaStruct any) error {
	var jsonStruct map[string]any
	if err := json.Unmarshal([]byte(str), &jsonStruct); err != nil {
		return err
	}
	delete(jsonStruct, "$schema")

	return ValidateJsonStruct(jsonStruct, schemaStruct)

}

func ValidateJsonStruct(jsonStruct map[string]any, schemaStruct any) error {
	schemaString, err := generateJsonSchema(schemaStruct)
	if err != nil {
		return err
	}

	schema, err := generateJsonSchemaValidator(schemaString)
	if err != nil {
		return err
	}

	delete(jsonStruct, "$schema")

	if err := schema.Validate(jsonStruct); err != nil {
		return err
	}

	return nil

}

func generateJsonSchema(s any) (string, error) {
	reflector := jsonschema.Reflector{
		Anonymous: true,
	}
	schema := reflector.Reflect(&s)
	bytes, err := json.MarshalIndent(schema, "", " ")
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func generateJsonSchemaValidator(schema string) (*jsonschema2.Schema, error) {
	compiler := jsonschema2.NewCompiler()
	err := compiler.AddResource("database_schema.json", strings.NewReader(schema))
	if err != nil {
		return nil, err
	}
	return compiler.Compile("database_schema.json")
}
