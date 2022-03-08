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
	grootType, err := getOrCreateType(t)
	if err != nil {
		return nil, err
	}

	if object, ok := grootType.(*Object); ok {
		return object, nil
	}

	return nil, ErrNotObject
}

func ParseInputObject(t reflect.Type) (*Input, error) {
	grootType, err := getOrCreateArgumentType(t)
	if err != nil {
		return nil, err
	}

	if inputObject, ok := grootType.(*Input); ok {
		return inputObject, nil
	}

	return nil, ErrNotInputObject
}

func ParseUnion(t reflect.Type) (*Union, error) {
	grootType, err := getOrCreateType(t)
	if err != nil {
		return nil, err
	}

	if union, ok := grootType.(*Union); ok {
		return union, nil
	}

	return nil, ErrNotUnion
}

func ParseInterface(t reflect.Type) (*Interface, error) {
	grootType, err := getOrCreateType(t)
	if err != nil {
		return nil, err
	}

	if interfaceType, ok := grootType.(*Interface); ok {
		return interfaceType, nil
	}

	return nil, ErrNotInterface
}

func ParseScalar(t reflect.Type) (*Scalar, error) {
	grootType, err := getOrCreateType(t)
	if err != nil {
		return nil, err
	}

	if scalar, ok := grootType.(*Scalar); ok {
		return scalar, nil
	}

	return nil, ErrNotScalar
}

func ParseEnum(t reflect.Type) (*Enum, error) {
	grootType, err := getOrCreateType(t)
	if err != nil {
		return nil, err
	}

	if enum, ok := grootType.(*Enum); ok {
		return enum, nil
	}

	return nil, ErrNotEnum
}
