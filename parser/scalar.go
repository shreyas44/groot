package parser

import "reflect"

type ScalarKind int

const (
	BuiltInScalar ScalarKind = iota
	CustomScalar
)

type Scalar struct {
	reflect.Type
}

func NewScalar(t reflect.Type) (*Scalar, error) {
	if err := validateTypeKind(t, KindScalar, KindCustomScalar); err != nil {
		panic(err)
	}

	scalar := &Scalar{Type: t}
	cache.set(t, scalar)
	return scalar, nil
}

func (s Scalar) ReflectType() reflect.Type {
	return s.Type
}
