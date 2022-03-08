package parser

import (
	"fmt"
	"reflect"
)

type Interface struct {
	reflectType reflect.Type
	fields      []*Field
}

func NewInterface(t reflect.Type) (*Interface, error) {
	if err := validateTypeKind(t, KindInterface); err != nil {
		panic(err)
	}

	if err := validateInterface(t); err != nil {
		return nil, err
	}

	interfaceDefReflectType := t.Method(0).Type.Out(0)
	interface_, err := getOrCreateType(interfaceDefReflectType)
	if err != nil {
		return nil, err
	}

	cache.set(t, interface_)
	return interface_.(*Interface), nil
}

func NewInterfaceFromDefinition(t reflect.Type) (*Interface, error) {
	if err := validateTypeKind(t, KindInterfaceDefinition); err != nil {
		panic(err)
	}

	interface_ := &Interface{
		reflectType: t,
	}

	cache.set(t, interface_)

	fields, err := getFields(interface_, t)
	if err != nil {
		return nil, err
	}

	interface_.fields = fields
	return interface_, nil
}

func (i *Interface) Name() string {
	name := i.reflectType.Name()
	name = name[:len(name)-len("Definition")]
	return name
}

func (i *Interface) Fields() []*Field {
	return i.fields
}

func (i *Interface) ReflectType() reflect.Type {
	return i.reflectType
}

func validateInterface(t reflect.Type) error {
	if t.NumMethod() != 1 {
		return fmt.Errorf(
			"interface %s can have only one method",
			t.Name(),
		)
	}

	method := t.Method(0)

	if method.Type.NumIn() != 0 {
		return fmt.Errorf(
			"method %s on interface %s should not have input arguments",
			method.Name,
			t.Name(),
		)
	}

	if method.Type.NumOut() != 1 {
		return fmt.Errorf(
			"method %s on interface %s should return exactly one value",
			method.Name,
			t.Name(),
		)
	}

	outType := method.Type.Out(0)

	if err := validateTypeKind(outType, KindInterfaceDefinition); err != nil {
		return fmt.Errorf(
			"method %s on interface %s should return a struct with groot.InterfaceType embedded",
			method.Name,
			t.Name(),
		)
	}

	if t.Name()+"Definition" != outType.Name() {
		return fmt.Errorf(
			"method %s on interface %s should return a struct named %sDefinition",
			method.Name,
			t.Name(),
			t.Name(),
		)
	}

	return nil
}
