package db_utils

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"reflect"
	"strings"
	"time"
)

func TableStructToGormModel(table Table) any {
	columns := make([]reflect.StructField, 0)

	// 从 monitor_model.Model 中获取 ID, CreatedAt, UpdatedAt, DeletedAt 字段
	valueType := reflect.ValueOf(monitor_model.Model{}).Type()
	for i := 0; i < valueType.NumField(); i++ {
		columns = append(columns, valueType.Field(i))
	}

	for _, column := range table.Columns {
		columns = append(columns, reflect.StructField{
			// 首字母大写
			Name: strings.ToUpper(column.Name[0:1]) + column.Name[1:],
			Type: ColumnTypeToGormType(column.Type),
			//Tag:  ColumnToGormTag(column),
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
	return reflect.StructTag(`json:"` + column.Name + `" gorm:"column:` + column.Name + `"`)
}
