package groot

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

type (
	fieldResolverArgType int
)

type CustomName interface {
	GraphQLName() string
}

const (
	fieldResolverArgInputArg fieldResolverArgType = iota
	fieldResolverArgContext
	fieldResolverArgInfo
	fieldResolverArgOther
)

type Field struct {
	name              string
	description       string
	deprecationReason string
	type_             GrootType
	arguments         []*Argument
	resolve           graphql.FieldResolveFn
	subscribe         graphql.FieldResolveFn
	field             *graphql.Field
}

func NewField(structType reflect.Type, structField reflect.StructField, builder *SchemaBuilder) (*Field, error) {
	if parserType, _ := getParserType(structType); parserType != ParserObject && parserType != ParserInterfaceDefinition {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to NewField must have parser type ParserObject or ParserInterfaceDefinition, received %s",
			structType.Name(),
			parserType,
		)
		panic(err)
	}

	var (
		name              string
		description       string
		depracationReason string
		arguments         []*Argument
		resolver          graphql.FieldResolveFn
		subscribe         graphql.FieldResolveFn = nil
	)

	grootType, err := getFieldGrootType(structType, structField, builder)
	if err != nil {
		return nil, err
	}

	if ignoreTag := structField.Tag.Get("groot_ignore"); ignoreTag == "true" {
		return nil, nil
	}

	if nameTag := structField.Tag.Get("json"); nameTag != "" {
		name = nameTag
	} else {
		name = structField.Name
	}

	if descTag := structField.Tag.Get("description"); descTag != "" {
		description = descTag
	}

	if deprecate := structField.Tag.Get("deprecate"); deprecate != "" {
		depracationReason = deprecate
	}

	// default resolver
	resolver = func(p graphql.ResolveParams) (interface{}, error) {
		value := reflect.ValueOf(p.Source).FieldByName(structField.Name)
		return value.Interface(), nil
	}

	// subscription resolver
	if structType.Name() == "Subscription" {
		method, _ := structType.MethodByName(fmt.Sprintf("Subscribe%s", structField.Name))
		returnType := reflect.ChanOf(reflect.RecvDir, structField.Type)
		subscribe, err = buildSubscriptionResolver(method, returnType, grootType)
		if err != nil {
			return nil, err
		}

		resolver = func(p graphql.ResolveParams) (interface{}, error) {
			return p.Source, nil
		}

		argsSignature := getResolverArgumentSignature(method)
		for i, arg := range argsSignature {
			if arg == fieldResolverArgInputArg {
				arguments, err = getArguments(method.Func.Type().In(i+1), builder)
				if err != nil {
					return nil, err
				}
			}
		}
	} else if method, exists := structType.MethodByName(fmt.Sprintf("Resolve%s", structField.Name)); exists {
		// custom resolver

		returnType := structField.Type
		resolver, err = buildResolver(method, returnType, grootType)
		if err != nil {
			return nil, err
		}

		argsSignature := getResolverArgumentSignature(method)
		for i, arg := range argsSignature {
			if arg == fieldResolverArgInputArg {
				arguments, err = getArguments(method.Func.Type().In(i+1), builder)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	field := &Field{
		name:              name,
		description:       description,
		type_:             grootType,
		resolve:           resolver,
		subscribe:         subscribe,
		deprecationReason: depracationReason,
		arguments:         arguments,
	}

	return field, nil
}

func (field *Field) GraphQLField() *graphql.Field {
	if field.field != nil {
		return field.field
	}

	args := graphql.FieldConfigArgument{}

	for _, argument := range field.arguments {
		args[argument.name] = argument.GraphQLArgument()
	}

	field.field = &graphql.Field{
		Name:              field.name,
		Type:              field.type_.GraphQLType(),
		Description:       field.description,
		Resolve:           field.resolve,
		DeprecationReason: field.deprecationReason,
		Args:              args,
	}

	if field.subscribe != nil {
		field.field.Subscribe = field.subscribe
	}

	return field.field
}

func getArguments(t reflect.Type, builder *SchemaBuilder) ([]*Argument, error) {
	arguments := []*Argument{}
	fieldCount := t.NumField()

	if parserType, _ := getParserType(t); parserType != ParserObject && parserType != ParserInterfaceDefinition {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to getArguments must have parser type ParserObject or ParserInterfaceDefinition, received %s",
			t.Name(),
			parserType,
		)

		panic(err)
	}

	for i := 0; i < fieldCount; i++ {
		fieldType := t.Field(i)
		argument, err := NewArgument(t, fieldType, builder)
		if err != nil {
			return nil, err
		}

		if argument != nil {
			arguments = append(arguments, argument)
		}
	}

	return arguments, nil
}

func validateFieldType(structType reflect.Type, structField reflect.StructField) error {
	parserType, err := getParserType(structField.Type)
	if err != nil {
		return fmt.Errorf(
			"field type %s not supported for field %s on struct %s \nif you think this is a mistake please open an issue at github.com/shreyas44/groot",
			structField.Type.Name(),
			structField.Name,
			structType.Name(),
		)
	}

	if parserType == ParserInterfaceDefinition {
		return fmt.Errorf(
			"received an interface definition for field type %s for field %s on struct %s\n"+
				"create a Go interface corresponding to the GraphQL interface and use that instead\n"+
				"see https://groot.shreyas44.com/type-definitions/interface for more info",
			structField.Type.Name(),
			structField.Name,
			structType.Name(),
		)
	}

	return nil
}

func getFieldGrootType(structType reflect.Type, structField reflect.StructField, builder *SchemaBuilder) (GrootType, error) {
	if parserType, _ := getParserType(structType); parserType != ParserObject && parserType != ParserInterfaceDefinition {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to getFieldGrootType must have parser type ParserObject or ParserInterfaceDefinition, received %s",
			structType.Name(),
			parserType,
		)
		panic(err)
	}

	parserType, _ := getParserType(structField.Type)
	err := validateFieldType(structType, structField)
	if err != nil {
		return nil, err
	}

	if grootType := builder.getType(structField.Type); grootType != nil {
		return NewNonNull(grootType), nil
	}

	var fieldType GrootType
	switch parserType {
	case ParserScalar, ParserCustomScalar:
		fieldType, err = NewScalar(structField.Type, builder)
	case ParserObject:
		fieldType, err = NewObject(structField.Type, builder)
	case ParserInterface, ParserInterfaceDefinition:
		fieldType, err = NewInterface(structField.Type, builder)
	case ParserUnion:
		fieldType, err = NewUnion(structField.Type, builder)
	case ParserEnum:
		fieldType, err = NewEnum(structField.Type, builder)
	case ParserList:
		field := structField
		field.Type = field.Type.Elem()
		item, err := getFieldGrootType(structType, field, builder)
		if err != nil {
			return nil, err
		}

		return NewNonNull(NewArray(item)), nil
	case ParserNullable:
		field := structField
		field.Type = field.Type.Elem()
		item, err := getFieldGrootType(structType, field, builder)
		if err != nil {
			return nil, err
		}

		return GetNullable(item), nil
	}

	if err != nil {
		return nil, err
	}

	return NewNonNull(fieldType), nil
}

func getResolverArgumentSignature(method reflect.Method) []fieldResolverArgType {
	var (
		arguments        = []fieldResolverArgType{}
		contextInterface = reflect.TypeOf((*context.Context)(nil)).Elem()
		resolverInfoType = reflect.TypeOf(graphql.ResolveInfo{})
		funcType         = method.Func.Type()
	)

	// start from 1 to ignore receiver
	for i := 1; i < method.Func.Type().NumIn(); i++ {
		arg := funcType.In(i)
		if arg.Implements(contextInterface) {
			arguments = append(arguments, fieldResolverArgContext)
		} else if arg == resolverInfoType {
			arguments = append(arguments, fieldResolverArgInfo)
		} else if arg.Kind() == reflect.Struct {
			arguments = append(arguments, fieldResolverArgInputArg)
		} else {
			arguments = append(arguments, fieldResolverArgOther)
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
		validArgPermuations = [][]fieldResolverArgType{
			{fieldResolverArgInputArg, fieldResolverArgContext, fieldResolverArgInfo},
			{fieldResolverArgInputArg, fieldResolverArgContext},
			{fieldResolverArgInputArg, fieldResolverArgInfo},
			{fieldResolverArgContext, fieldResolverArgInfo},
			{fieldResolverArgInputArg},
			{fieldResolverArgContext},
			{fieldResolverArgInfo},
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

// TODO: use groot type to get return type, we can construct a new type from it
func buildResolver(method reflect.Method, returnType reflect.Type, grootType GrootType) (graphql.FieldResolveFn, error) {
	argsSignature := getResolverArgumentSignature(method)

	if err := validateFieldResolver(method, returnType); err != nil {
		return nil, err
	}

	union, isUnion := GetNullable(grootType).(*Union)
	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := makeResolverArgs(method, argsSignature, p)
		if err != nil {
			return nil, err
		}

		response := method.Func.Call(args)
		value, resErr := response[0], response[1]

		if isUnion {
			p := graphql.ResolveTypeParams{
				Value:   value.Interface(),
				Info:    p.Info,
				Context: p.Context,
			}

			value = union.resolveValue(p)
		}

		if resErr.IsNil() {
			return value.Interface(), nil
		}

		return value.Interface(), resErr.Interface().(error)
	}, nil
}

// TODO: use groot type to get return type, we can construct a new type from it
func buildSubscriptionResolver(method reflect.Method, returnType reflect.Type, grootType GrootType) (graphql.FieldResolveFn, error) {
	argsSignature := getResolverArgumentSignature(method)

	if err := validateFieldResolver(method, returnType); err != nil {
		return nil, err
	}

	union, isUnion := GetNullable(grootType).(*Union)
	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := makeResolverArgs(method, argsSignature, p)
		if err != nil {
			return nil, err
		}

		response := method.Func.Call(args)
		resCh, resErr := response[0], response[1]

		if !resErr.IsNil() {
			return nil, resErr.Interface().(error)
		}

		ch := make(chan interface{})
		valueCh := make(chan reflect.Value)

		go func() {
			for {
				value, ok := resCh.Recv()
				if !ok {
					close(valueCh)
					return
				}

				select {
				case <-p.Context.Done():
					close(valueCh)
					return
				default:
					valueCh <- value
				}
			}
		}()

		go func() {
			for value := range valueCh {
				if isUnion {
					p := graphql.ResolveTypeParams{
						Value:   value.Interface(),
						Info:    p.Info,
						Context: p.Context,
					}

					value = union.resolveValue(p)
				}

				ch <- value.Interface()
			}

			close(ch)
		}()

		return ch, nil
	}, nil
}

func makeResolverArgs(method reflect.Method, argsSignature []fieldResolverArgType, p graphql.ResolveParams) ([]reflect.Value, error) {
	var (
		funcType   = method.Func.Type()
		structType = funcType.In(0)
		args       = []reflect.Value{}
	)

	// if it's a map, it's a root query
	if _, isMap := p.Source.(map[string]interface{}); isMap {
		args = append(args, reflect.Indirect(reflect.New(method.Type.In(0))))
	} else {
		args = append(args, reflect.ValueOf(p.Source).Convert(structType))
	}

	for _, arg := range argsSignature {
		switch arg {
		case fieldResolverArgInputArg:
			structInterface := reflect.New(funcType.In(1)).Interface()
			jsonBytes, err := json.Marshal(p.Args)
			if err != nil {
				return nil, err
			}

			json.Unmarshal(jsonBytes, &structInterface)
			args = append(args, reflect.Indirect(reflect.ValueOf(structInterface)))
		case fieldResolverArgContext:
			args = append(args, reflect.ValueOf(p.Context))
		case fieldResolverArgInfo:
			args = append(args, reflect.ValueOf(p.Info))
		}
	}

	return args, nil
}
