package groot

import (
	"encoding/json"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type InputValidator interface {
	Validate() error
}

func buildResolver(resolver *parser.Resolver) graphql.FieldResolveFn {
	parserReturnType := resolver.Field().Type()

	if !resolver.ReturnsThunk() {
		return func(p graphql.ResolveParams) (interface{}, error) {
			args, err := makeResolverArgs(resolver, p)
			if err != nil {
				return nil, err
			}

			response := resolver.Func.Call(args)
			return makeResolverOutput(p, parserReturnType, response)
		}
	}

	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := makeResolverArgs(resolver, p)
		if err != nil {
			return nil, err
		}

		response := resolver.Func.Call(args)
		thunk, resErr := response[0], response[1]
		if !resErr.IsNil() {
			return nil, resErr.Interface().(error)
		}

		return func() (interface{}, error) {
			output := thunk.Call([]reflect.Value{})
			return makeResolverOutput(p, parserReturnType, output)
		}, nil
	}
}

func buildSubscriptionResolver(subscriber *parser.Subscriber, parserType parser.Type) graphql.FieldResolveFn {
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
				response := []reflect.Value{value, reflect.ValueOf((*error)(nil))}
				output, _ := makeResolverOutput(p, parserType, response)
				ch <- output
			}

			close(ch)
		}()

		return ch, nil
	}
}

func makeResolverArgs(resolver *parser.Resolver, p graphql.ResolveParams) ([]reflect.Value, error) {
	var (
		funcType = resolver.Func.Type()
		args     = []reflect.Value{}
	)

	// if it's a map, it's a root query
	if _, isMap := p.Source.(map[string]interface{}); isMap {
		args = append(args, reflect.Indirect(reflect.New(resolver.Type.In(0))))
	} else {
		value := reflect.ValueOf(p.Source)
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		args = append(args, value)
	}

	for _, arg := range resolver.ArgsSignature() {
		switch arg {
		case parser.ResolverArgInput:
			// TODO: figure out a better way to do this instead of marshalling and unmarshalling

			structInterface := reflect.New(funcType.In(1)).Interface()
			jsonBytes, err := json.Marshal(p.Args)
			if err != nil {
				return nil, err
			}

			json.Unmarshal(jsonBytes, &structInterface)
			if i, ok := structInterface.(InputValidator); ok {
				if err := i.Validate(); err != nil {
					return nil, err
				}
			}

			args = append(args, reflect.Indirect(reflect.ValueOf(structInterface)))
		case parser.ResolverArgContext:
			args = append(args, reflect.ValueOf(p.Context))
		case parser.ResolverArgInfo:
			args = append(args, reflect.ValueOf(p.Info))
		}
	}

	return args, nil
}

func makeResolverOutput(p graphql.ResolveParams, parserType parser.Type, response []reflect.Value) (interface{}, error) {
	var union *parser.Union
	var isUnion bool
	value, resErr := response[0], response[1]

	if nullable, isNullable := parserType.(*parser.Nullable); isNullable {
		union, isUnion = nullable.Element().(*parser.Union)
	} else {
		union, isUnion = parserType.(*parser.Union)
	}

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
