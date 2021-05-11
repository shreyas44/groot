package gql

import (
	"fmt"
	"reflect"
)

func GetType(gType interface{}) (reflect.Type, error) {
	reflectType := reflect.TypeOf(gType)

	if reflectType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is not a struct", reflectType.Name())
	}

	return reflectType, nil
}

func GetTypes(gTypes ...interface{}) ([]reflect.Type, error) {
	reflectTypes := []reflect.Type{}

	for _, gType := range gTypes {
		reflectType, err := GetType(gType)

		if err != nil {
			return nil, err
		}

		reflectTypes = append(reflectTypes, reflectType)
	}

	return reflectTypes, nil
}