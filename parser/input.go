package parser

import (
	"reflect"
)

type Input struct {
	reflect.Type
	arguments []*Argument
}

func NewInput(t reflect.Type) (*Input, error) {
	if err := validateTypeKind(t, KindObject); err != nil {
		return nil, err
	}

	input := &Input{
		Type:      t,
		arguments: []*Argument{},
	}

	cache.set(t, input)

	arguments, err := getArguments(input, t)
	if err != nil {
		return nil, err
	}

	input.arguments = arguments
	return input, nil
}

func (i *Input) Arguments() []*Argument {
	return i.arguments
}

func (i *Input) ReflectType() reflect.Type {
	return i.Type
}

func getArguments(t *Input, reflectType reflect.Type) ([]*Argument, error) {
	args := []*Argument{}

	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)

		if kind, _ := getTypeKind(field.Type); field.Anonymous && kind == KindObject {
			embeddedArgs, err := getArguments(t, field.Type)
			if err != nil {
				return nil, err
			}

			args = append(args, embeddedArgs...)
			continue
		}

		arg, err := NewArgument(t, field)
		if err != nil {
			return nil, err
		}

		args = append(args, arg)
	}

	return args, nil
}