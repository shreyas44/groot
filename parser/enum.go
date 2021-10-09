package parser

import "reflect"

type Enum struct {
	reflect.Type
	values []string
}

func NewEnum(t reflect.Type) (*Enum, error) {
	if err := validateTypeKind(t, KindEnum); err != nil {
		panic(err)
	}

	values := reflect.New(t).
		MethodByName("Values").
		Call([]reflect.Value{})[0].
		Interface().([]string)

	enum := &Enum{t, values}
	cache.set(t, enum)
	return enum, nil
}

func (e *Enum) Values() []string {
	return e.values
}

func (e *Enum) ReflectType() reflect.Type {
	return e.Type
}
