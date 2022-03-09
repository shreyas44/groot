package groot

import (
	"encoding/json"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/shreyas44/groot/parser"
)

type inputArgsValidator func(v reflect.Value) error
type fieldResolver = graphql.FieldResolveFn
type fieldSubscriber = graphql.FieldResolveFn

func newInputArgsValidator(input *parser.Input) inputArgsValidator {
	validators := []inputArgsValidator{}

	if input == nil {
		return nil
	}

	if validator := input.Validator(); validator != nil {
		validator := func(v reflect.Value) error {
			res := validator.ReflectMethod().Func.Call([]reflect.Value{v})
			resErr := res[0]
			if !resErr.IsNil() {
				return resErr.Interface().(error)
			}

			return nil
		}

		validators = append(validators, validator)
	}

	for _, arg := range input.Arguments() {
		if validator := arg.Validator(); validator != nil {
			validator := func(v reflect.Value) error {
				var (
					field  = v.FieldByName(arg.StructField().Name)
					values = []reflect.Value{v, field}
					res    = validator.ReflectMethod().Func.Call(values)
					resErr = res[0]
				)

				if !resErr.IsNil() {
					return resErr.Interface().(error)
				}

				return nil
			}

			validators = append(validators, validator)
		}

		if input, ok := arg.Type().(*parser.Input); ok {
			validator := newInputArgsValidator(input)
			validator = func(v reflect.Value) error {
				field := v.FieldByName(arg.Type().ReflectType().Name())
				return validator(field)
			}

			validators = append(validators, validator)
		}
	}

	return func(v reflect.Value) error {
		for _, validator := range validators {
			if err := validator(v); err != nil {
				return err
			}
		}

		return nil
	}
}

func newFieldResolver(field *parser.Field) fieldResolver {
	if field.Subscriber() != nil {
		return newSubsriberFieldResolver(field)
	}

	if field.Resolver() == nil {
		return newDefaultFieldResolver(field)
	}

	return newCustomFieldResolver(field.Resolver())
}

func newSubsriberFieldResolver(field *parser.Field) fieldSubscriber {
	return func(p graphql.ResolveParams) (interface{}, error) {
		return p.Source, nil
	}
}

func newDefaultFieldResolver(field *parser.Field) fieldResolver {
	return func(p graphql.ResolveParams) (interface{}, error) {
		value := reflect.ValueOf(p.Source)
		name := field.StructField().Name

		if value.Type().Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if _, ok := field.Type().(*parser.Enum); ok {
			return value.FieldByName(name).Convert(reflect.TypeOf("")).Interface(), nil
		}

		return value.FieldByName(name).Interface(), nil
	}
}

func newCustomFieldResolver(resolver *parser.Resolver) fieldResolver {
	parserReturnType := resolver.Field().Type()
	resolverFunc := resolver.ReflectMethod().Func
	validateInputArgs := newInputArgsValidator(resolver.Field().ArgsInput())

	if !resolver.ReturnsThunk() {
		return func(p graphql.ResolveParams) (interface{}, error) {
			args, err := makeResolverArgs(resolver, validateInputArgs, p)
			if err != nil {
				return nil, err
			}

			response := resolverFunc.Call(args)
			return makeResolverOutput(p, parserReturnType, response)
		}
	}

	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := makeResolverArgs(resolver, validateInputArgs, p)
		if err != nil {
			return nil, err
		}

		response := resolverFunc.Call(args)
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

func newFieldSubscriber(subscriber *parser.Subscriber, parserType parser.Type) fieldResolver {
	subscriberFunc := subscriber.ReflectMethod().Func
	validateInputArgs := newInputArgsValidator(subscriber.Field().ArgsInput())

	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := makeResolverArgs(subscriber, validateInputArgs, p)
		if err != nil {
			return nil, err
		}

		response := subscriberFunc.Call(args)
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

func makeResolverArgs(resolver *parser.Resolver, validateInputArgs inputArgsValidator, p graphql.ResolveParams) ([]reflect.Value, error) {
	var (
		resolverMethod = resolver.ReflectMethod()
		resolverFunc   = resolverMethod.Func
		funcType       = resolverFunc.Type()
		args           = []reflect.Value{}
	)

	// if it's a map, it's a root query
	if _, isMap := p.Source.(map[string]interface{}); isMap {
		args = append(args, reflect.Indirect(reflect.New(resolverMethod.Type.In(0))))
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
			inputArgs := reflect.Indirect(reflect.ValueOf(structInterface))
			if err := validateInputArgs(inputArgs); err != nil {
				return nil, err
			}

			args = append(args, inputArgs)
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
