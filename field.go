package groot

import (
	"encoding/json"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type CustomName interface {
	GraphQLName() string
}

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

func NewField(field *parser.Field, builder *SchemaBuilder) (*Field, error) {
	var (
		arguments []*Argument
		resolver  graphql.FieldResolveFn
		subscribe graphql.FieldResolveFn = nil
	)

	grootType, err := getOrCreateType(field.Type(), builder)
	if err != nil {
		return nil, err
	}

	// default resolver
	resolver = func(p graphql.ResolveParams) (interface{}, error) {
		value := reflect.ValueOf(p.Source).FieldByName(field.Name)
		return value.Interface(), nil
	}

	if field.Subscriber() != nil {
		// subscription resolver
		subscribe = buildSubscriptionResolver(field.Subscriber(), grootType)
		resolver = func(p graphql.ResolveParams) (interface{}, error) {
			return p.Source, nil
		}

	} else if field.Resolver() != nil {
		// custom resolver
		resolver = buildResolver(field.Resolver(), grootType)
	}

	if args := field.Arguments(); args != nil {
		if arguments, err = getArguments(args, builder); err != nil {
			return nil, err
		}
	}

	grootField := &Field{
		name:              field.JSONName(),
		description:       field.Description(),
		type_:             grootType,
		resolve:           resolver,
		subscribe:         subscribe,
		deprecationReason: field.DeprecationReason(),
		arguments:         arguments,
	}

	return grootField, nil
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

func getOrCreateType(t parser.Type, builder *SchemaBuilder) (GrootType, error) {
	if grootType, ok := builder.getType(t); ok {
		return NewNonNull(grootType), nil
	}

	var grootType GrootType
	var err error

	switch t := t.(type) {
	case *parser.Scalar:
		grootType, err = NewScalar(t, builder)
	case *parser.Enum:
		grootType, err = NewEnum(t, builder)
	case *parser.Object:
		grootType, err = NewObject(t, builder)
	case *parser.Interface:
		grootType, err = NewInterface(t, builder)
	case *parser.Union:
		grootType, err = NewUnion(t, builder)
	case *parser.Input:
		grootType, err = NewInputObject(t, builder)
	case *parser.Array:
		grootType, err = NewArray(t, builder)
	case *parser.Nullable:
		gType, err := getOrCreateType(t.Element(), builder)
		if err != nil {
			return nil, err
		}

		return GetNullable(gType), nil
	default:
		panic("groot: unexpected error occurred")
	}

	if err != nil {
		return nil, err
	}

	return NewNonNull(grootType), nil
}

func getArguments(args []*parser.Argument, builder *SchemaBuilder) ([]*Argument, error) {
	arguments := []*Argument{}

	for _, arg := range args {
		argument, err := NewArgument(arg, builder)
		if err != nil {
			return nil, err
		}

		if argument != nil {
			arguments = append(arguments, argument)
		}
	}

	return arguments, nil
}

func buildResolver(resolver *parser.Resolver, grootType GrootType) graphql.FieldResolveFn {
	union, isUnion := GetNullable(grootType).(*Union)

	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := makeResolverArgs(resolver, p)
		if err != nil {
			return nil, err
		}

		response := resolver.Func.Call(args)
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
	}
}

func buildSubscriptionResolver(subscriber *parser.Subscriber, grootType GrootType) graphql.FieldResolveFn {
	union, isUnion := GetNullable(grootType).(*Union)

	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := makeResolverArgs(subscriber, p)
		if err != nil {
			return nil, err
		}

		response := subscriber.Func.Call(args)
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
	}
}

func makeResolverArgs(resolver *parser.Resolver, p graphql.ResolveParams) ([]reflect.Value, error) {
	var (
		funcType   = resolver.Func.Type()
		structType = funcType.In(0)
		args       = []reflect.Value{}
	)

	// if it's a map, it's a root query
	if _, isMap := p.Source.(map[string]interface{}); isMap {
		args = append(args, reflect.Indirect(reflect.New(resolver.Type.In(0))))
	} else {
		args = append(args, reflect.ValueOf(p.Source).Convert(structType))
	}

	for _, arg := range resolver.ArgsSignature() {
		switch arg {
		case parser.ResolverArgInput:
			structInterface := reflect.New(funcType.In(1)).Interface()
			jsonBytes, err := json.Marshal(p.Args)
			if err != nil {
				return nil, err
			}

			json.Unmarshal(jsonBytes, &structInterface)
			args = append(args, reflect.Indirect(reflect.ValueOf(structInterface)))
		case parser.ResolverArgContext:
			args = append(args, reflect.ValueOf(p.Context))
		case parser.ResolverArgInfo:
			args = append(args, reflect.ValueOf(p.Info))
		}
	}

	return args, nil
}
