package db_utils

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"gorm.io/gorm"
	"reflect"
	"time"
)

func TableStructToGormModel(db *gorm.DB, table Table) any {
	columns := make([]reflect.StructField, 0)

	// 从 monitor_model.Model 中获取 ID, CreatedAt, UpdatedAt, DeletedAt 字段
	valueType := reflect.ValueOf(monitor_model.Model{}).Type()
	for i := 0; i < valueType.NumField(); i++ {
		columns = append(columns, valueType.Field(i))
	}

	for _, column := range table.Columns {
		columns = append(columns, reflect.StructField{
			// 转换外部定义的列名为 Golang struct 属性名.(主要为首字母大写)
			Name: db.NamingStrategy.SchemaName(column.Name),
			Type: ColumnTypeToGormType(column.Type),
			Tag:  ColumnToGormTag(column),
		})
	}

	model := reflect.StructOf(columns)
	return reflect.New(model).Interface()
}

func ColumnTypeToGormType(columnType string) reflect.Type {
	switch columnType {
	case "int":
		return reflect.TypeOf(int(0))
	case "uint":
		return reflect.TypeOf(uint(0))
	case "time":
		return reflect.TypeOf(time.Now())
	case "float":
	case "number":
		return reflect.TypeOf(float64(0))
	case "boolean":
		return reflect.TypeOf(false)
	case "string":
	default:
		return reflect.TypeOf("")
	}

	return reflect.TypeOf("")
}

func ColumnToGormTag(column Column) reflect.StructTag {
	return reflect.StructTag(`json:"` + column.Name + `"`)
}
