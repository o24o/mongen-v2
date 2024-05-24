package gen

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

type gen[T any] struct {
	obj        T
	basePath   string
	objectName string
	fields     []Field
}

func (s *gen[T]) SetBasePath(basePath string) {
	s.basePath = basePath
}

type Field struct {
	Name string
	Bson string
	Type string
}

func (s *gen[T]) parseStruct(strucct interface{}) error {
	typ := reflect.TypeOf(strucct)
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("input is not a struct pointer")
	}

	var fields []Field

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			// 如果是匿名内嵌结构体字段,递归获取其字段
			embeddedFields, err := s.parseEmbeddedFields(field.Type)
			if err != nil {
				return err
			}
			fields = append(fields, embeddedFields...)
		} else {
			bson := toLowerCamelCase(field.Name)
			tmp := field.Tag.Get("bson")
			if tmp != "" && !strings.HasSuffix(tmp, ",") {
				bsonArr := strings.Split(tmp, ",")
				if len(bsonArr) > 0 {
					bson = bsonArr[0]
				}
			}
			// 否则直接添加该字段
			fields = append(fields, Field{
				Name: field.Name,
				Bson: bson,
				Type: field.Type.String(),
			})
		}
	}

	s.fields = fields
	s.objectName = typ.Name()
	return nil
}

// parseEmbeddedFields 递归解析内嵌结构体字段
func (s *gen[T]) parseEmbeddedFields(typ reflect.Type) ([]Field, error) {
	var fields []Field

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			// 如果是嵌套的内嵌结构体字段,继续递归
			embeddedFields, err := s.parseEmbeddedFields(field.Type)
			if err != nil {
				return nil, err
			}
			fields = append(fields, embeddedFields...)
		} else {
			bson := toLowerCamelCase(field.Name)
			tmp := field.Tag.Get("bson")
			if tmp != "" && !strings.HasSuffix(tmp, ",") {
				bsonArr := strings.Split(tmp, ",")
				if len(bsonArr) > 0 {
					bson = bsonArr[0]
				}
			}
			// 否则直接添加该字段
			fields = append(fields, Field{
				Name: field.Name,
				Type: field.Type.String(),
				Bson: bson,
			})
		}
	}

	return fields, nil
}
func (s *gen[T]) genDao() error {
	if s.basePath == "" {
		s.basePath = "dist"
	}
	var code strings.Builder
	code.WriteString(fmt.Sprintf(`
package dao

import (
	"gen/common"
	"model"
)

type %s struct {
	common.Q[model.%s]
}
var %s = &%s{}
`, toLowerCamelCase(s.objectName), toUpperCamelCase(s.objectName), toUpperCamelCase(s.objectName), toLowerCamelCase(s.objectName)))

	for _, field := range s.fields {
		fieldName := field.Name
		fieldType := field.Type
		// 生成字段变量
		code.WriteString(fmt.Sprintf(`
func (q *%s) %s() common.Field[%s] {
	return common.Field[%s]{Name: "%s",Bson: "%s"}
}`, toLowerCamelCase(s.objectName), fieldName, fieldType, fieldType, fieldName, field.Bson))
	}
	daoPath := filepath.Join(s.basePath, "dao")
	err := os.MkdirAll(daoPath, 0755)
	if err != nil {
		return err
	}
	filename := filepath.Join(daoPath, fmt.Sprintf("%s.go", toLowerCamelCase(s.objectName)))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(code.String())
	return err
}

func (s *gen[T]) genModel() error {
	var err error
	if s.basePath == "" {
		s.basePath = "dist"
	}
	fieldsArr := []string{}
	for _, field := range s.fields {
		fieldName := field.Name
		fieldType := field.Type

		tag := toLowerCamelCase(fieldName)
		// 生成字段变量
		f := fmt.Sprintf("%s %s `bson:\"%s\"`", fieldName, fieldType, tag)
		fieldsArr = append(fieldsArr, f)
	}
	filedStr := strings.Join(fieldsArr, "\n")
	var code strings.Builder
	code.WriteString(fmt.Sprintf(`
package model

type %s struct {
%s
}
`, s.objectName, filedStr))

	modelPath := filepath.Join(s.basePath, "model")
	err = os.MkdirAll(modelPath, 0755)
	if err != nil {
		return err
	}
	filename := filepath.Join(modelPath, fmt.Sprintf("%s.go", toLowerCamelCase(s.objectName)))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(code.String())
	return err
}

func Gen[T any](in T, path string) error {
	g := gen[T]{}
	g.SetBasePath(path)
	err := g.parseStruct(in)
	if err != nil {
		return err
	}
	if err != nil {
		fmt.Println("Error parsing struct:", err)
		return err
	}
	err = g.genDao()
	if err != nil {
		fmt.Println("Error generating code:", err)
		return err
	}
	err = g.genModel()
	if err != nil {
		fmt.Println("Error generating code:", err)
		return err
	}
	fmt.Println("Code generated successfully")
	return nil
}

func toLowerCamelCase(s string) string {
	var result string
	words := strings.Split(s, "_")
	for i, word := range words {
		if i == 0 {
			result = strings.ToLower(word)
		} else {
			result += toTitleCase(word)
		}
	}
	return result
}

func toUpperCamelCase(str string) string {
	words := strings.Split(str, "_")
	upperCamelCase := make([]string, len(words))

	for i, word := range words {
		if word == "" {
			continue
		}
		upperCamelCase[i] = toTitleCase(word)
	}

	return strings.Join(upperCamelCase, "")
}

func toTitleCase(str string) string {
	var result string
	wordBuf := make([]rune, 0, len(str))
	upNextWord := true

	for len(str) > 0 {
		r, size := utf8.DecodeRuneInString(str)
		str = str[size:]

		if unicode.IsLetter(r) {
			if upNextWord {
				r = unicode.ToUpper(r)
				upNextWord = false
			} else {
				r = unicode.ToLower(r)
			}
		} else {
			upNextWord = true
		}

		wordBuf = append(wordBuf, r)
	}

	result = string(wordBuf)
	return result
}
