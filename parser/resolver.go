package parser

import (
	"context"
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

type Subscriber = Resolver

type (
	ResolverArgType int
	ResolveType     int
)

const (
	ResolverArgOther ResolverArgType = iota
	ResolverArgInput
	ResolverArgContext
	ResolverArgInfo
)

type Resolver struct {
	reflect.Method
	field     *Field
	signature []ResolverArgType
}

func NewResolver(field *Field) (*Resolver, error) {
	var (
		methodName string
		fieldName  = field.Name
		object     = field.Object()
	)

	if object.Name() == "Subscription" {
		methodName = fmt.Sprintf("Subscribe%s", fieldName)
	} else {
		methodName = fmt.Sprintf("Resolve%s", fieldName)
	}

	method, hasMethod := object.MethodByName(methodName)

	if object.Name() == "Subscription" {
		if !hasMethod {
			return nil, fmt.Errorf(
				"subscription field %s must have a subscriber method with name %s defined",
				fieldName,
				methodName,
			)
		}

		if err := validateFieldSubscriber(method, reflect.ChanOf(reflect.RecvDir, field.StructField.Type)); err != nil {
			return nil, err
		}
	} else if hasMethod {
		if err := validateFieldResolver(method, field.StructField.Type); err != nil {
			return nil, err
		}
	} else {
		return nil, nil
	}

	return &Resolver{
		Method:    method,
		field:     field,
		signature: getResolverArgumentSignature(method),
	}, nil
}

func (r *Resolver) ArgsSignature() []ResolverArgType {
	return r.signature
}

func (r *Resolver) Field() *Field {
	return r.field
}

func (r *Resolver) ReturnsThunk() bool {
	return r.Type.Out(0).Kind() == reflect.Func
}

func getResolverArgumentSignature(method reflect.Method) []ResolverArgType {
	// method doesn't exis
	if method.Type == nil {
		return []ResolverArgType{}
	}

	var (
		arguments        = []ResolverArgType{}
		contextInterface = reflect.TypeOf((*context.Context)(nil)).Elem()
		resolverInfoType = reflect.TypeOf(graphql.ResolveInfo{})
		funcType         = method.Func.Type()
	)

	// start from 1 to ignore receiver
	for i := 1; i < method.Func.Type().NumIn(); i++ {
		arg := funcType.In(i)
		if arg.Implements(contextInterface) {
			arguments = append(arguments, ResolverArgContext)
		} else if arg == resolverInfoType {
			arguments = append(arguments, ResolverArgInfo)
		} else if arg.Kind() == reflect.Struct {
			arguments = append(arguments, ResolverArgInput)
		} else {
			arguments = append(arguments, ResolverArgOther)
		}
	}

	return arguments
}

func validateResolverArguments(method reflect.Method) error {
	var (
		funcType = method.Func.Type()
		inCount  = funcType.NumIn()

		structType          = funcType.In(0)
		argsSignature       = getResolverArgumentSignature(method)
		validArgPermuations = [][]ResolverArgType{
			{ResolverArgInput, ResolverArgContext, ResolverArgInfo},
			{ResolverArgInput, ResolverArgContext},
			{ResolverArgInput, ResolverArgInfo},
			{ResolverArgContext, ResolverArgInfo},
			{ResolverArgInput},
			{ResolverArgContext},
			{ResolverArgInfo},
			{},
		}
	)

	if inCount > 4 {
		return fmt.Errorf(
			"resolver %s on struct %s can accept only up to 3 arguments of type (Args, context.Context, graphql.ResolveInfo)",
			method.Name,
			structType.Name(),
		)
	}

	isValid := false
	for _, permutation := range validArgPermuations {
		if reflect.DeepEqual(argsSignature, permutation) {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf(
			"resolver %s on struct %s can accept either no arguments or one of the following permutations: \n"+
				"(args, context, info)\n"+
				"(args, context)\n"+
				"(args, info)\n"+
				"(context, info)\n"+
				"(args)\n"+
				"(context)\n"+
				"(info)",
			method.Name,
			structType.Name(),
		)
	}

	return nil
}

func validateResolverOutput(method reflect.Method, returnType reflect.Type) error {
	var (
		funcType       = method.Func.Type()
		errorInterface = reflect.TypeOf((*error)(nil)).Elem()
		outCount       = funcType.NumOut()
		structType     = funcType.In(0)
	)

	err := func() error {
		returnMsg := ""
		for i := 0; i < outCount; i++ {
			returnMsg += funcType.Out(i).String() + ", "
		}
		returnMsg = returnMsg[:len(returnMsg)-2]

		return fmt.Errorf(
			"one of the below return types was expect for resolver %s on struct %s, got (%s)\n"+
				"(%s, error)\n"+
				"(func() (%s, error), error)",
			method.Name,
			structType.Name(),
			returnMsg,
			returnType,
			returnType,
		)
	}()

	if outCount != 2 || !method.Type.Out(1).Implements(errorInterface) {
		return err
	}

	actualReturnType := method.Type.Out(0)

	if actualReturnType.Kind() == reflect.Func {
		returnMethod := method
		returnMethod.Type = actualReturnType

		if subErr := validateFieldSubscriber(returnMethod, returnType); subErr != nil {
			return err
		}

		return nil
	}

	if actualReturnType != returnType {
		return err
	}

	return nil
}

func validateSubscriberOutput(method reflect.Method, returnType reflect.Type) error {
	var (
		funcType       = method.Func.Type()
		errorInterface = reflect.TypeOf((*error)(nil)).Elem()
		outCount       = funcType.NumOut()
		structType     = funcType.In(0)
	)

	if outCount != 2 {
		return fmt.Errorf(
			"return type of (%s, error) was expected for resolver %s on struct %s, got %s",
			returnType.Name(),
			method.Name,
			structType.Name(),
			funcType.In(1),
		)
	}

	if method.Type.Out(0) != returnType || !method.Type.Out(1).Implements(errorInterface) {
		return fmt.Errorf(
			"return type of (%s, error) was expected for resolver %s on struct %s, got (%s, %s)",
			returnType.String(),
			method.Name,
			structType.Name(),
			funcType.Out(0).String(),
			funcType.Out(1).String(),
		)
	}

	return nil
}

func validateFieldResolver(method reflect.Method, returnType reflect.Type) error {
	if err := validateResolverArguments(method); err != nil {
		return err
	}

	if err := validateResolverOutput(method, returnType); err != nil {
		return err
	}

	return nil
}

func validateFieldSubscriber(method reflect.Method, returnType reflect.Type) error {
	if err := validateResolverArguments(method); err != nil {
		return err
	}

	if err := validateSubscriberOutput(method, returnType); err != nil {
		return err
	}

	return nil
}

func getResolverArguments(resolver *Resolver) ([]*Argument, error) {
	signature := resolver.ArgsSignature()
	if len(signature) == 0 || signature[0] != ResolverArgInput {
		return []*Argument{}, nil
	}

	reflectType := resolver.Method.Type.In(1)

	// this input type will not be created in the schema
	input, err := getOrCreateArgumentType(reflectType)
	if err != nil {
		return nil, err
	}

	args, err := getArguments(input.(*Input), reflectType)
	if err != nil {
		return nil, err
	}

	return args, nil
}
