package common

type Field[T any] struct {
	Name string
	Bson string
	Type string
}

type QueryCondition struct {
	Field string
	Op    string
	Value interface{}
}

// Eq 等于条件
func (f Field[T]) Eq(value T) QueryCondition {
	return QueryCondition{Field: f.Bson, Op: "$eq", Value: value}
}

// In 包含条件
func (f Field[T]) In(values ...T) QueryCondition {
	return QueryCondition{Field: f.Bson, Op: "$in", Value: values}
}
