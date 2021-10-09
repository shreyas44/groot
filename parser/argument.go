package parser

import (
	"fmt"
	"reflect"
)

type Argument struct {
	reflect.StructField
	input        *Input
	type_        Type
	jsonName     string
	defaultValue string
	description  string
}

func NewArgument(input *Input, field reflect.StructField) (*Argument, error) {
	argument := &Argument{
		StructField:  field,
		input:        input,
		description:  field.Tag.Get("description"),
		jsonName:     field.Tag.Get("json"),
		defaultValue: field.Tag.Get("default"),
	}

	if err := validateArgumentType(argument); err != nil {
		return nil, err
	}

	type_, err := getOrCreateArgumentType(field.Type)
	if err != nil {
		return nil, err
	}

	argument.type_ = type_
	return argument, nil
}

func (arg *Argument) ArgType() Type {
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

	return arg.Name
}

func (arg *Argument) DefaultValue() string {
	return arg.defaultValue
}

func (arg *Argument) ImplementsType() {}

func validateArgumentType(arg *Argument) error {
	kind, err := getTypeKind(arg.Type)
	if err != nil {
		return err
	}

	switch kind {
	case KindInterface, KindUnion, KindInterfaceDefinition:
		return fmt.Errorf(
			"argument type %s not supported for field %s on struct %s \nif you think this is a mistake please open an issue at github.com/shreyas44/groot",
			arg.StructField.Type.Name(),
			arg.StructField.Name,
			arg.Input().Name(),
		)
	}

	return nil
}

func getOrCreateArgumentType(t reflect.Type) (Type, error) {
	parserType, ok := cache.get(t)
	if ok {
		kind, err := getTypeKind(t)
		if err != nil {
			return nil, err
		}

		switch kind {
		case KindInterface, KindUnion, KindInterfaceDefinition:
			err := fmt.Errorf("")
			return nil, err
		}

		return parserType, nil
	}

	kind, err := getTypeKind(t)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("interface and union not supported for argument type")
	}

	panic("parser: unexpected error occurred")
}
