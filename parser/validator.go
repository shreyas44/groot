package parser

import (
	"fmt"
	"reflect"
)

type ArgumentValidator struct {
	reflectMethod reflect.Method
	argument      *Argument
}

type InputValidator struct {
	reflectMethod reflect.Method
	input         *Input
}

func NewArgumentValidator(argument *Argument) (*ArgumentValidator, error) {
	method, hasMethod := argument.input.reflectType.MethodByName(
		fmt.Sprintf("Validate%s", argument.structField.Name),
	)

	if !hasMethod {
		return nil, nil
	}

	if err := validateArgValidator(method, argument); err != nil {
		return nil, err
	}

	return &ArgumentValidator{method, argument}, nil
}

func NewInputValidator(input *Input) (*InputValidator, error) {
	method, hasMethod := input.reflectType.MethodByName("Validate")

	if !hasMethod {
		return nil, nil
	}

	if err := validateInputValidator(method); err != nil {
		return nil, err
	}

	return &InputValidator{method, input}, nil
}

func (v *ArgumentValidator) ReflectMethod() reflect.Method {
	return v.reflectMethod
}

func (v *InputValidator) ReflectMethod() reflect.Method {
	return v.reflectMethod
}

func validateArgValidator(method reflect.Method, arg *Argument) error {
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	numIn := method.Type.NumIn()
	numOut := method.Type.NumOut()

	if numIn != 2 || (numIn > 1 && method.Type.In(1) != arg.structField.Type) {
		return fmt.Errorf(
			"method %s on struct %s expected to have 1 argument of type (%s)",
			method.Name,
			method.Type.In(0).Name(),
			arg.structField.Type,
		)
	}

	if numOut != 1 || (numOut > 0 && method.Type.Out(0) != errorInterface) {
		return fmt.Errorf(
			"method %s on struct %s expected to return only error",
			method.Name,
			method.Type.In(0).Name(),
		)
	}

	return nil
}

func validateInputValidator(method reflect.Method) error {
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()

	if method.Type.NumIn() > 1 {
		return fmt.Errorf(
			"method %s on struct %s expected to have 0 arguments",
			method.Name,
			method.Type.In(0).Name(),
		)
	}

	if method.Type.NumOut() != 1 || method.Type.Out(0) != errorInterface {
		return fmt.Errorf(
			"method %s on struct %s expected to return only error",
			method.Name,
			method.Type.In(0).Name(),
		)
	}

	return nil
}
