package parser

import (
	"fmt"
	"reflect"
)

type Argument struct {
	structField  reflect.StructField
	validator    *ArgumentValidator
	input        *Input
	type_        Type
	jsonName     string
	defaultValue string
	description  string
}

func NewArgument(input *Input, field reflect.StructField) (*Argument, error) {
	if field.Tag.Get("json") == "-" || !field.IsExported() {
		return nil, nil
	}

	argument := &Argument{
		structField:  field,
		input:        input,
		description:  field.Tag.Get("description"),
		jsonName:     field.Tag.Get("json"),
		defaultValue: field.Tag.Get("default"),
	}

	type_, err := getOrCreateArgumentType(field.Type)
	if err != nil {
		return nil, err
	}

	validator, err := NewArgumentValidator(argument)
	if err != nil {
		return nil, err
	}

	argument.validator = validator
	argument.type_ = type_
	return argument, nil
}

func (arg *Argument) Type() Type {
	return arg.type_
}

func (arg *Argument) Input() *Input {
	return arg.input
}

func (arg *Argument) Description() string {
	return arg.description
}

func (arg *Argument) JSONName() string {
	if arg.jsonName != "" {
		return arg.jsonName
	}

	return arg.structField.Name
}

func (arg *Argument) DefaultValue() string {
	return arg.defaultValue
}

func (arg *Argument) Validator() *ArgumentValidator {
	return arg.validator
}

func (arg *Argument) StructField() reflect.StructField {
	return arg.structField
}

func getOrCreateArgumentType(t reflect.Type) (Type, error) {
	unupportedErr := fmt.Errorf("interface and union not supported for argument type")
	kind, err := getTypeKind(t)
	if err != nil {
		return nil, err
	}

	parserType, ok := cache.get(t)
	if ok {
		switch kind {
		case KindObject:
			if _, ok := parserType.(*Input); ok {
				return parserType, nil
			}
		case KindInterface, KindUnion, KindInterfaceDefinition:
			return nil, unupportedErr
		default:
			return parserType, nil
		}
	}

	switch kind {
	case KindScalar, KindCustomScalar:
		return NewScalar(t)
	case KindObject:
		return NewInput(t)
	case KindEnum:
		return NewEnum(t)
	case KindList:
		return NewArray(t, true)
	case KindNullable:
		return NewNullable(t, true)
	case KindInterface, KindUnion, KindInterfaceDefinition:
		return nil, unupportedErr
	}

	panic("parser: unexpected error occurred")
}
