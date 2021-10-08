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
)

const (
	ResolverArgInput ResolverArgType = iota
	ResolverArgContext
	ResolverArgInfo
	ResolverArgOther
)

type Resolver struct {
	reflect.Method
	field     *ObjectField
	signature []ResolverArgType
}

func getResolverArgumentSignature(method reflect.Method) []ResolverArgType {
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

func validateFieldResolver(method reflect.Method, returnType reflect.Type) error {
	var (
		funcType       = method.Func.Type()
		errorInterface = reflect.TypeOf((*error)(nil)).Elem()

		outCount = funcType.NumOut()
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
			returnType.Name(),
			method.Name,
			structType.Name(),
			funcType.Out(0).Name(),
			funcType.Out(1).Name(),
		)
	}

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

func NewResolver(t *Type, field *ObjectField) (*Resolver, error) {
	var (
		methodName string
		returnType = field.fieldType.Type
		fieldName  = field.Name
	)

	if t.Name() == "Subscription" {
		methodName = fmt.Sprintf("Subscribe%s", fieldName)
		returnType = reflect.ChanOf(reflect.RecvDir, returnType)
	} else {
		methodName = fmt.Sprintf("Resolve%s", fieldName)
	}

	method, hasMethod := t.MethodByName(methodName)
	if !hasMethod {
		if t.Name() == "Subscription" {
			err := fmt.Errorf(
				"subscription field %s must have a subscriber method with name %s defined",
				fieldName,
				methodName,
			)

			return nil, err
		}

		return nil, nil
	}

	if err := validateFieldResolver(method, returnType); err != nil {
		return nil, err
	}

	return &Resolver{
		Method:    method,
		field:     field,
		signature: getResolverArgumentSignature(method),
	}, nil
}

func (r *Resolver) Signature() []ResolverArgType {
	return r.signature
}

func (r *Resolver) Field() *ObjectField {
	return r.field
}

func getArguments(resolver *Resolver) (*Type, error) {
	if resolver == nil {
		return nil, nil
	}

	signature := resolver.Signature()
	for i, arg := range signature {
		if arg == ResolverArgInput {
			reflectType := resolver.Method.Type.In(i + 1)

			t, ok := cache.get(reflectType)
			if ok {
				return t, nil
			}

			t, err := NewType(reflectType)
			if err != nil {
				return nil, err
			}

			return t, nil
		}
	}

	return nil, nil
}
