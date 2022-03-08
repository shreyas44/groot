package parser

import "reflect"

type Array struct {
	reflectType reflect.Type
	element     Type
}

func NewArray(t reflect.Type, isArgument bool) (*Array, error) {
	var element Type
	var err error

	if err := validateTypeKind(t, KindList); err != nil {
		panic(err)
	}

	if isArgument {
		element, err = getOrCreateArgumentType(t.Elem())
	} else {
		element, err = getOrCreateType(t.Elem())
	}

	if err != nil {
		return nil, err
	}

	array := &Array{t, element}
	return array, nil
}

func (a *Array) Element() Type {
	return a.element
}

func (a *Array) ReflectType() reflect.Type {
	return a.reflectType
}
