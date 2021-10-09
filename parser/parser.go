package parser

import "reflect"

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

func ParseType(t reflect.Type) (Type, error) {
	return getOrCreateType(t)
}
