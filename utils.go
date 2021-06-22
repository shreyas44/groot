package groot

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

func GetType(t interface{}) (graphql.Type, error) {
	reflectType :=  reflect.TypeOf(t)

	if reflectType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is not a struct", reflectType.Name())
	}

	field := NewType(reflectType)

	return field.Type(), nil
}

func GetTypes(types ...interface{}) ([]graphql.Type, error) {
	reflectTypes := []graphql.Type{}

	for _, gType := range types {
		reflectType, err := GetType(gType)

		if err != nil {
			return nil, err
		}

		reflectTypes = append(reflectTypes, reflectType)
	}

	return reflectTypes, nil
}