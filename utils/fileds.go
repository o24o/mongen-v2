package utils

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
)

// 获取字段名的函数
func getFieldName(structPtr interface{}, fieldPtr interface{}) (string, error) {
	structVal := reflect.ValueOf(structPtr).Elem()
	for i := 0; i < structVal.NumField(); i++ {
		if structVal.Field(i).Addr().Interface() == fieldPtr {
			return structVal.Type().Field(i).Name, nil
		}
	}
	return "", fmt.Errorf("field not found in struct")
}

// Bson 定义函数获取字段的 BSON 标签值
func Bson(structPtr interface{}, fieldPtr interface{}) string {
	fieldName, err := getFieldName(structPtr, fieldPtr)
	if err != nil {
		logrus.Fatal(err)
	}
	structType := reflect.TypeOf(structPtr).Elem()
	if field, found := structType.FieldByName(fieldName); found {
		return field.Tag.Get("bson")
	}
	return fieldName
}
