package groot

import (
	"fmt"
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
		fieldTypeName := field.Type.Name()

		gField := graphql.Field{ 
			Type: gType,
			DeprecationReason: field.Tag.Get("deprecate"),
			Description: field.Tag.Get("description"),
		}

		// Get the field name on type
		if value := field.Tag.Get("json"); value != "" {
			fieldName = value
		} else {
			name := field.Name
			if config.CamelCase {
				name = strings.ToLower(string(name[0])) + name[1:]
			}
			
			fieldName = name
		}

		// Get the graphql.Type struct or add to queu and continue
		if field.Type.Kind() != reflect.Struct {
			gType = GetGType(field)
		} else if _, ok := types[fieldTypeName]; ok {
			gType = GetGType(field)
		} else {
			relationQueue.add(fieldName, field, gObject)
			continue
		}

		// Add resolvers
		resolver, ok := field.Type.MethodByName("Resolve" + field.Type.Name())
		if ok {
			// check resolver type
			numOut := resolver.Func.Type().NumOut()

			if numOut != 2 {
				panic(fmt.Sprintf("resolver for %v must have the return 2 values of type (%v, error), returned only %d value", fieldTypeName, fieldTypeName, numOut))
			}

			fReturnType := resolver.Func.Type().Out(0)
			sReturnType := resolver.Func.Type().Out(1)
			if fReturnType != field.Type || sReturnType != reflect.TypeOf(new(error)) {
				panic(fmt.Sprintf("resolver for %v must have a return type of (%v, error), got (%v, %v)", fieldTypeName, fieldTypeName, fReturnType, sReturnType))
			}

			// add resolver to graphql.Type

			gField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
				args := []reflect.Value{}

				// add context to args
				args = append(args, reflect.ValueOf(p.Context))

				// add input args to args

				response := resolver.Func.Call(args)

				return response[0], response[1].Interface().(error)
			}
		} else {
			// default resolver

			gField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
				return reflect.ValueOf(p.Source).FieldByName(field.Type.Name()).Interface(), nil
			}
		}

		gField.Type = gType
		gObject.AddFieldConfig(fieldName, &gField)
	}

	types[objectName] = gObject

	return gObject
}