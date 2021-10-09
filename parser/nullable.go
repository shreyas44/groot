package parser

import (
	"reflect"
)

type Nullable struct {
	reflect.Type
	element Type
}

func NewNullable(t reflect.Type, isArgument bool) (*Nullable, error) {
	var element Type
	var err error

	if err := validateTypeKind(t, KindNullable); err != nil {
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

	nullable := &Nullable{t, element}
	return nullable, nil
}

func (n *Nullable) Element() Type {
	return n.element
}

func (n *Nullable) ReflectType() reflect.Type {
	return n.Type
}
