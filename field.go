package groot

import (
	"encoding/json"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

func NewField(parserField *parser.Field, builder *SchemaBuilder) *graphql.Field {
	var resolver graphql.FieldResolveFn
	var subscribe graphql.FieldResolveFn

	graphqlType := getOrCreateType(parserField.Type(), builder)

	// default resolver
	resolver = func(p graphql.ResolveParams) (interface{}, error) {
		value := reflect.ValueOf(p.Source).FieldByName(parserField.Name)
		return value.Interface(), nil
	}

	if parserField.Subscriber() != nil {
		// subscription resolver
		subscribe = buildSubscriptionResolver(parserField.Subscriber(), parserField.Type())
		resolver = func(p graphql.ResolveParams) (interface{}, error) {
			return p.Source, nil
		}

	} else if parserField.Resolver() != nil {
		// custom resolver
		resolver = buildResolver(parserField.Resolver(), parserField.Type())
	}

	args := graphql.FieldConfigArgument{}
	for _, parserArgs := range parserField.Arguments() {
		args[parserArgs.JSONName()] = NewArgument(parserArgs, builder)
	}

	field := &graphql.Field{
		Name:              parserField.JSONName(),
		Type:              graphqlType,
		Description:       parserField.Description(),
		Resolve:           resolver,
		DeprecationReason: parserField.DeprecationReason(),
		Args:              args,
	}

	if subscribe != nil {
		field.Subscribe = subscribe
	}

	return field
}

func getOrCreateType(t parser.Type, builder *SchemaBuilder) graphql.Type {
	if graphqlType, ok := builder.getType(t); ok {
		return NewNonNull(graphqlType)
	}

	switch t := t.(type) {
	case *parser.Scalar:
		return NewNonNull(NewScalar(t, builder))
	case *parser.Enum:
		return NewNonNull(NewEnum(t, builder))
	case *parser.Object:
		return NewNonNull(NewObject(t, builder))
	case *parser.Interface:
		return NewNonNull(NewInterface(t, builder))
	case *parser.Union:
		return NewNonNull(NewUnion(t, builder))
	case *parser.Input:
		return NewNonNull(NewInputObject(t, builder))
	case *parser.Array:
		return NewNonNull(NewArray(getOrCreateType(t.Element(), builder)))
	case *parser.Nullable:
		return GetNullable(getOrCreateType(t.Element(), builder)).(graphql.Type)
	}

	panic("groot: unexpected error occurred")
}

func buildResolver(resolver *parser.Resolver, parserType parser.Type) graphql.FieldResolveFn {
	var union *parser.Union
	var isUnion bool

	if nullable, isNullable := parserType.(*parser.Nullable); isNullable {
		union, isUnion = nullable.Element().(*parser.Union)
	} else {
		union, isUnion = parserType.(*parser.Union)
	}

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

			value = resolveUnionValue(union, p)
		}

		if resErr.IsNil() {
			return value.Interface(), nil
		}

		return value.Interface(), resErr.Interface().(error)
	}
}

func buildSubscriptionResolver(subscriber *parser.Subscriber, parserType parser.Type) graphql.FieldResolveFn {
	var union *parser.Union
	var isUnion bool

	if nullable, isNullable := parserType.(*parser.Nullable); isNullable {
		union, isUnion = nullable.Element().(*parser.Union)
	} else {
		union, isUnion = parserType.(*parser.Union)
	}

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

					value = resolveUnionValue(union, p)
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
