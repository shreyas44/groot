package parser

import (
	"fmt"
	"reflect"
)

var (
	_ Type = (*Object)(nil)
	_ Type = (*Array)(nil)
	_ Type = (*Union)(nil)
	_ Type = (*Interface)(nil)
	_ Type = (*Nullable)(nil)
	_ Type = (*Scalar)(nil)
	_ Type = (*Enum)(nil)
	_ Type = (*Input)(nil)
)

var (
	_ TypeWithFields = (*Object)(nil)
	_ TypeWithFields = (*Interface)(nil)
)

var (
	_ TypeWithElement = (*Array)(nil)
	_ TypeWithElement = (*Nullable)(nil)
)

type Type interface {
	reflect.Type
	ReflectType() reflect.Type
}

type TypeWithFields interface {
	Type
	Fields() []*Field
}

type TypeWithElement interface {
	Type
	Element() Type
}

var (
	ErrNotObject      = fmt.Errorf("type is not object")
	ErrNotInputObject = fmt.Errorf("type is not input object")
	ErrNotUnion       = fmt.Errorf("type is not union")
	ErrNotInterface   = fmt.Errorf("type is not interface")
	ErrNotScalar      = fmt.Errorf("type is not scalar")
	ErrNotEnum        = fmt.Errorf("type is not enum")
)

func ParseObject(t reflect.Type) (*Object, error) {
	t, err := getOrCreateType(t)
	if err != nil {
		return nil, err
	}

	if object, ok := t.(*Object); ok {
		return object, nil
	}

	return nil, ErrNotObject
}

func ParseInputObject(t reflect.Type) (*Input, error) {
	t, err := getOrCreateArgumentType(t)
	if err != nil {
		return nil, err
	}

	if inputObject, ok := t.(*Input); ok {
		return inputObject, nil
	}

	return nil, ErrNotInputObject
}

func ParseUnion(t reflect.Type) (*Union, error) {
	t, err := getOrCreateType(t)
	if err != nil {
		return nil, err
	}

	if union, ok := t.(*Union); ok {
		return union, nil
	}

	return nil, ErrNotUnion
}

func ParseInterface(t reflect.Type) (*Interface, error) {
	t, err := getOrCreateType(t)
	if err != nil {
		return nil, err
	}

	if interfaceType, ok := t.(*Interface); ok {
		return interfaceType, nil
	}

	return nil, ErrNotInterface
}

func ParseScalar(t reflect.Type) (*Scalar, error) {
	t, err := getOrCreateType(t)
	if err != nil {
		return nil, err
	}

	if scalar, ok := t.(*Scalar); ok {
		return scalar, nil
	}

	return nil, ErrNotScalar
}

func ParseEnum(t reflect.Type) (*Enum, error) {
	t, err := getOrCreateType(t)
	if err != nil {
		return nil, err
	}

	if enum, ok := t.(*Enum); ok {
		return enum, nil
	}

	return nil, ErrNotEnum
}
