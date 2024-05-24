package common

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
)

// Q 结构体
type Q[T any] struct {
	collection *mongo.Collection
	ctx        context.Context
	filter     []QueryCondition
}

// Collection 设置集合
func (q *Q[T]) Collection(collection *mongo.Collection) *Q[T] {
	q.collection = collection
	return q
}

// WithContext 设置查询上下文
func (q *Q[T]) WithContext(ctx context.Context) *Q[T] {
	q.ctx = ctx
	return q
}

// Where 添加过滤条件
func (q *Q[T]) Where(condition ...QueryCondition) *Q[T] {
	q.filter = append(q.filter, condition...)
	return q
}

// First 返回第一个匹配的结果
func (q *Q[T]) First() (*T, error) {
	var result T
	filter := bson.D{}
	for _, c := range q.filter {
		filter = append(filter, bson.E{Key: c.Field, Value: bson.D{{c.Op, c.Value}}})
	}
	err := q.collection.FindOne(q.ctx, filter).Decode(&result)
	return &result, err
}

// Find 返回所有匹配的结果
func (q *Q[T]) Find() ([]*T, error) {
	filter := bson.D{}
	for _, c := range q.filter {
		filter = append(filter, bson.E{Key: c.Field, Value: bson.D{{c.Op, c.Value}}})
	}
	cursor, err := q.collection.Find(q.ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(q.ctx)

	var results []*T
	if err = cursor.All(q.ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// UpdateOne 更新匹配的第一个文档
func (q *Q[T]) UpdateOne(update T) (*mongo.UpdateResult, error) {
	filter := bson.D{}
	for _, c := range q.filter {
		filter = append(filter, bson.E{Key: c.Field, Value: bson.D{{c.Op, c.Value}}})
	}

	// 使用反射构建更新文档
	updateDoc := bson.D{}
	v := reflect.ValueOf(update)
	t := reflect.TypeOf(update)
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		if !fieldValue.IsZero() {
			fieldName := t.Field(i).Tag.Get("bson")
			if fieldName == "" {
				fieldName = t.Field(i).Name
			}
			updateDoc = append(updateDoc, bson.E{Key: fieldName, Value: fieldValue.Interface()})
		}
	}

	if len(updateDoc) == 0 {
		return nil, nil // 没有需要更新的字段
	}

	updateBson := bson.D{{"$set", updateDoc}}
	updateResult, err := q.collection.UpdateOne(q.ctx, filter, updateBson)
	return updateResult, err
}

// InsertOne 插入一个新文档
func (q *Q[T]) InsertOne(document *T) (*mongo.InsertOneResult, error) {
	insertResult, err := q.collection.InsertOne(q.ctx, document)
	return insertResult, err
}

// UpsertOne 插入一个新文档或更新现有文档
func (q *Q[T]) UpsertOne(document *T, uniqueFields []string) (*mongo.UpdateResult, error) {
	// 构建过滤条件
	filter := bson.D{}
	v := reflect.ValueOf(document).Elem()
	t := reflect.TypeOf(document).Elem()

	for _, field := range uniqueFields {
		fieldValue := v.FieldByName(field)
		if fieldValue.IsValid() {
			filter = append(filter, bson.E{Key: field, Value: fieldValue.Interface()})
		}
	}

	// 构建更新文档
	updateDoc := bson.D{}
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		if !fieldValue.IsZero() {
			fieldName := t.Field(i).Tag.Get("bson")
			if fieldName == "" {
				fieldName = t.Field(i).Name
			}
			updateDoc = append(updateDoc, bson.E{Key: fieldName, Value: fieldValue.Interface()})
		}
	}

	if len(updateDoc) == 0 {
		return nil, nil // 没有需要更新的字段
	}

	updateBson := bson.D{{"$set", updateDoc}}
	updateOptions := options.Update().SetUpsert(true)
	updateResult, err := q.collection.UpdateOne(q.ctx, filter, updateBson, updateOptions)
	return updateResult, err
}

// InsertMany 插入多个新文档
func (q *Q[T]) InsertMany(documents []*T) (*mongo.InsertManyResult, error) {
	var docs []interface{}
	for _, doc := range documents {
		docs = append(docs, doc)
	}
	insertResult, err := q.collection.InsertMany(q.ctx, docs)
	return insertResult, err
}
