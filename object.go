package gql

import (
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
)

var types = map[string]graphql.Type{
	"string": graphql.String,
	"int": graphql.Int,
	"float32": graphql.Float,
	"bool": graphql.Boolean,
}

func GetArrayGType(t reflect.Type) graphql.Type {
	elem := t.Elem()

	if elem.Kind() == reflect.Slice {
		gType := GetArrayGType(elem)
		return graphql.NewList(gType)
	}

	return graphql.NewList(types[elem.Name()])
}

func GetGType(field reflect.StructField) graphql.Type {
	var gType graphql.Type
	
	t := field.Type
	
	if t.Kind() == reflect.Slice {
		gType = GetArrayGType(t)
	} else {
		typeName := t.Name()
		gType = types[typeName]
	}

	if field.Tag.Get("nullable") == "false" {
		gType = graphql.NewNonNull(gType)
	}

	return gType
}

type ObjectConfig struct {
	CamelCase bool
}

func NewObject(t reflect.Type, relationQueue *RelationQueue, config ObjectConfig) *graphql.Object {
	objectName := t.Name()
	gObjectConfig := graphql.ObjectConfig{ Name: objectName, Fields: graphql.Fields{} }
	gObject := graphql.NewObject(gObjectConfig)
	
	fieldsCount := t.NumField()
	for i := 0; i < fieldsCount; i++ {
		var fieldName string
		var gType graphql.Type
		field := t.Field(i)

		gField := graphql.Field{ 
			Type: gType,
			DeprecationReason: field.Tag.Get("deprecate"),
			Description: field.Tag.Get("description"),
		}

		if value := field.Tag.Get("json"); value != "" {
			fieldName = value
		} else {
			name := field.Name
			if config.CamelCase {
				name = strings.ToLower(string(name[0])) + name[1:]
			}
			
			fieldName = name
		}

		if field.Type.Kind() != reflect.Struct {
			gType = GetGType(field)
		} else if _, ok := types[field.Type.Name()]; ok {
			gType = GetGType(field)
		} else {
			relationQueue.add(fieldName, field, gObject)
			continue
		}

		gField.Type = gType
		gObject.AddFieldConfig(fieldName, &gField)
	}

	types[objectName] = gObject

	return gObject
}