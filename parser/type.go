package parser

import (
	"fmt"
	"reflect"
)

type Kind int

const (
	KindScalar Kind = iota
	KindCustomScalar
	KindObject
	KindInterface
	KindInterfaceDefinition
	KindUnion
	KindEnum
	KindList
	KindNullable
	KindInvalidType
)

func (kind Kind) String() string {
	typeMap := map[Kind]string{
		KindScalar:              "Scalar",
		KindCustomScalar:        "CustomScalar",
		KindObject:              "Object",
		KindInterface:           "Interface",
		KindInterfaceDefinition: "InterfaceDefinition",
		KindUnion:               "Union",
		KindEnum:                "Enum",
		KindList:                "List",
		KindNullable:            "Nullable",
		KindInvalidType:         "InvalidType",
	}

	return typeMap[kind]
}

func validateTypeKind(t reflect.Type, expected ...Kind) error {
	kindString := ""

	if len(expected) == 0 {
		return nil
	}

	kind, err := getTypeKind(t)
	if err != nil {
		return err
	}

	for _, expectedKind := range expected {
		kindString += fmt.Sprintf("%s, ", expectedKind.String())

		if kind == expectedKind {
			return nil
		}
	}

	kindString = kindString[:len(kindString)-2]
	return fmt.Errorf("reflect.Type of kind %s was expected, got %s", kindString, kind.String())
}

func getOrCreateType(t reflect.Type) (Type, error) {
	parserType, ok := cache.get(t)
	if ok {
		return parserType, nil
	}

	kind, err := getTypeKind(t)
	if err != nil {
		return nil, err
	}

	switch kind {
	case KindScalar, KindCustomScalar:
		return NewScalar(t)
	case KindInterface:
		return NewInterface(t)
	case KindInterfaceDefinition:
		return NewInterfaceFromDefinition(t)
	case KindObject:
		return NewObject(t)
	case KindUnion:
		return NewUnion(t)
	case KindEnum:
		return NewEnum(t)
	case KindList:
		return NewArray(t, false)
	case KindNullable:
		return NewNullable(t, false)
	}

	panic("groot: unexpected error occurred")
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

func getTypeKind(t reflect.Type) (Kind, error) {
	var (
		enumType   = reflect.TypeOf((*EnumType)(nil)).Elem()
		scalarType = reflect.TypeOf((*ScalarType)(nil)).Elem()
	)

	if parserType, ok := t.(Type); ok {
		t = parserType.ReflectType()
	}

	if ptrT := reflect.PtrTo(t); ptrT.Implements(scalarType) {
		return KindCustomScalar, nil
	}

	switch t.Kind() {
	case reflect.Ptr:
		return KindNullable, nil

	case reflect.Slice:
		return KindList, nil

	case reflect.Interface:
		return KindInterface, nil

	case reflect.Struct:
		if isTypeUnion(t) {
			return KindUnion, nil
		}

		if isInterfaceDefinition(t) {
			return KindInterfaceDefinition, nil
		}

		return KindObject, nil

	case
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Float32, reflect.Float64,
		reflect.Bool:
		return KindScalar, nil

	case reflect.String:
		if t.Name() == "string" || !t.Implements(enumType) {
			return KindScalar, nil
		}

		return KindEnum, nil
	}

	return KindInvalidType, fmt.Errorf("couldn't parse type %s", t.Name())
}
