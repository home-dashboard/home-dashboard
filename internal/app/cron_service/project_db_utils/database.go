package project_db_utils

import (
	"encoding/json"
	"github.com/invopop/jsonschema"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
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
	Type string `json:"type" jsonschema:"title=column data type, enum=int, enum=uint, enum=time, enum=float, enum=boolean, required=true"`
}

var DatabaseSchema string

func init() {
	schema := jsonschema.Reflect(&Database{})
	bytes, err := json.MarshalIndent(schema, "", " ")
	if err != nil {
		logger.Fatal("generate database schema failed, %w", err)
	}

	DatabaseSchema = string(bytes)
}
