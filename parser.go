package groot

import (
	"fmt"
	"reflect"
)

type ParserType int

const (
	ParserScalar ParserType = iota
	ParserCustomScalar
	ParserObject
	ParserInterface
	ParserInterfaceDefinition
	ParserUnion
	ParserEnum
	ParserList
	ParserNullable
	ParserInvalidType
)

func (t ParserType) String() string {
	stringMap := map[ParserType]string{
		ParserScalar:              "ParserScalar",
		ParserCustomScalar:        "ParserCustomScalar",
		ParserObject:              "ParserObject",
		ParserInterface:           "ParserInterface",
		ParserInterfaceDefinition: "ParserInterfaceDefinition",
		ParserUnion:               "ParserUnion",
		ParserEnum:                "ParserEnum",
		ParserList:                "ParserList",
		ParserNullable:            "ParserNullable",
		ParserInvalidType:         "ParserInvalidType",
	}

	return stringMap[t]
}

func isTypeUnion(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous && field.Type == reflect.TypeOf(UnionType{}) {
			return true
		}
	}

	return false
}

func isInterfaceDefinition(t reflect.Type) bool {
	interfaceType := reflect.TypeOf(InterfaceType{})

	if t.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < t.NumField(); i++ {
		if field := t.Field(i); field.Anonymous && field.Type == interfaceType {
			return true
		}
	}

	return false
}

func getParserType(t reflect.Type) (ParserType, error) {
	var (
		enumType   = reflect.TypeOf((*EnumType)(nil)).Elem()
		scalarType = reflect.TypeOf((*ScalarType)(nil)).Elem()
	)

	if ptrT := reflect.PtrTo(t); ptrT.Implements(scalarType) {
		return ParserCustomScalar, nil
	}

	switch t.Kind() {
	case reflect.Ptr:
		return ParserNullable, nil

	case reflect.Slice:
		return ParserList, nil

	case reflect.Interface:
		return ParserInterface, nil

	case reflect.Struct:
		if isTypeUnion(t) {
			return ParserUnion, nil
		}

		if isInterfaceDefinition(t) {
			return ParserInterfaceDefinition, nil
		}

		return ParserObject, nil

	case
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Float32, reflect.Float64,
		reflect.Bool:
		return ParserScalar, nil

	case reflect.String:
		if t.Name() == "string" || !t.Implements(enumType) {
			return ParserScalar, nil
		}

		return ParserEnum, nil
	}

	return ParserInvalidType, fmt.Errorf("couldn't parse type %s", t.Name())
}
