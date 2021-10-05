package groot

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

type fieldResolverArgType int

type CustomName interface {
	GraphQLName() string
}

const (
	fieldResolverArgInputArg fieldResolverArgType = iota
	fieldResolverArgContext
	fieldResolverArgInfo
)

type Field struct {
	name              string
	description       string
	deprecationReason string
	type_             GrootType
	arguments         []*Argument
	resolve           func(p graphql.ResolveParams) (interface{}, error)
	field             *graphql.Field
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

	return field.field
}

func getArguments(t reflect.Type, builder *SchemaBuilder) []*Argument {
	arguments := []*Argument{}
	if t.Kind() != reflect.Struct {
		panic("argument type must be a struct")
	}

	fieldCount := t.NumField()
	for i := 0; i < fieldCount; i++ {
		fieldType := t.Field(i)
		if argument := NewArgument(fieldType, builder); argument != nil {
			arguments = append(arguments, argument)
		}
	}

	return arguments
}

func isTypeInterface(t reflect.Type) bool {
	interfaceType := reflect.TypeOf((*InterfaceType)(nil)).Elem()

	if !t.Implements(interfaceType) {
		return false
	}

	for i := 0; i < t.NumField(); i++ {
		if field := t.Field(i); field.Anonymous && field.Type.Implements(interfaceType) {
			return !isTypeInterface(field.Type)
		}
	}

	return true
}

func isTypeUnion(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous && field.Type == reflect.TypeOf(UnionType{}) {
			return true
		}
	}

	return false
}

func parseFieldType(t reflect.Type, builder *SchemaBuilder) GrootType {
	if grootType, ok := builder.grootTypes[t]; ok {
		return grootType
	}

	switch t.Kind() {
	case reflect.Ptr:
		return GetNullable(parseFieldType(t.Elem(), builder))

	case reflect.Slice:
		return NewNonNull(NewArray(parseFieldType(t.Elem(), builder)))

	case reflect.Struct:
		if isTypeInterface(t) {
			return NewNonNull(NewInterface(t, builder))
		}

		if isTypeUnion(t) {
			return NewNonNull(NewUnion(t, builder))
		}

		return NewNonNull(NewObject(t, builder))

	case reflect.Int, reflect.Bool, reflect.Float32:
		return NewNonNull(NewScalar(t, builder))

	case reflect.String:
		if t.Name() == "string" {
			return NewNonNull(NewScalar(t, builder))
		}

		return NewNonNull(NewEnum(t, builder))
	}

	panic("invalid struct field type")
}

func NewField(structType reflect.Type, structField reflect.StructField, builder *SchemaBuilder) *Field {
	var (
		name              string
		description       string
		depracationReason string
		arguments         []*Argument
		resolver          graphql.FieldResolveFn
		grootType         = parseFieldType(structField.Type, builder)
	)

	// find out how to avoid using a second argument
	if structType.Kind() != reflect.Struct {
		panic("type of second argument in NewField must be a struct")
	}

	if ignoreTag := structField.Tag.Get("groot_ignore"); ignoreTag == "true" {
		return nil
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
		// reflect.ValueOf(p.Source).Convert(structType)
		value := reflect.ValueOf(p.Source).FieldByName(structField.Name)
		return value.Interface(), nil
	}

	// custom resolver
	if method, exists := structType.MethodByName(fmt.Sprintf("Resolve%s", structField.Name)); exists {
		var (
			methodType = method.Func.Type()
			returnType = structField.Type

			contextInterface = reflect.TypeOf((*context.Context)(nil)).Elem()
			interfaceType    = reflect.TypeOf((*InterfaceType)(nil)).Elem()
			errorInterface   = reflect.TypeOf((*error)(nil)).Elem()
			resolverInfoType = reflect.TypeOf(graphql.ResolveInfo{})

			outCount = methodType.NumOut()
			inCount  = methodType.NumIn()

			resolverArgs        = []fieldResolverArgType{}
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
			panic(
				fmt.Sprintf("return type of (%s, error) was expected for resolver %s", returnType.Name(), method.Name),
			)
		}

		_, ok := GetNullable(grootType).(*Interface)
		if ok && (methodType.Out(0) != interfaceType || !methodType.Out(1).Implements(errorInterface)) {
			message := fmt.Sprintf(
				"return type of (%s, error) was expected for resolver %s, got (%s, %s)",
				interfaceType.Name(),
				method.Name,
				methodType.Out(0).Name(),
				methodType.Out(1).Name(),
			)

			panic(message)
		} else if !ok && (methodType.Out(0) != returnType || !methodType.Out(1).Implements(errorInterface)) {
			message := fmt.Sprintf(
				"return type of (%s, error) was expected for resolver %s, got (%s, %s)",
				returnType.Name(),
				method.Name,
				methodType.Out(0).Name(),
				methodType.Out(1).Name(),
			)

			panic(message)
		}

		if inCount > 4 {
			panic(
				fmt.Sprintf(
					"resolver %s can accept only up to 3 arguments of type (Args, context.Context, graphql.ResolveInfo)",
					method.Name,
				),
			)
		}

		for i := 1; i < inCount; i++ {
			arg := method.Type.In(i)

			switch {
			case arg.Implements(contextInterface):
				resolverArgs = append(resolverArgs, fieldResolverArgContext)
			case arg == resolverInfoType:
				resolverArgs = append(resolverArgs, fieldResolverArgInfo)
			case arg.Kind() == reflect.Struct:
				resolverArgs = append(resolverArgs, fieldResolverArgInputArg)
			}
		}

		isValid := false
		for _, permutation := range validArgPermuations {
			if reflect.DeepEqual(resolverArgs, permutation) {
				isValid = true

				for i, arg := range resolverArgs {
					if arg == fieldResolverArgInputArg {
						arguments = getArguments(method.Type.In(i+1), builder)
					}
				}

				break
			}
		}

		if !isValid {
			panic("invalid resolver args order")
		}

		union, isUnion := GetNullable(grootType).(*Union)
		resolver = func(p graphql.ResolveParams) (interface{}, error) {
			var source reflect.Value
			// if it's a map, it's a root query
			if _, isMap := p.Source.(map[string]interface{}); isMap {
				source = reflect.Indirect(reflect.New(method.Type.In(0)))
			} else {
				source = reflect.ValueOf(p.Source).Convert(structType)
			}

			args := []reflect.Value{source}

			for _, arg := range resolverArgs {
				switch arg {
				case fieldResolverArgInputArg:
					structInterface := reflect.New(methodType.In(1)).Interface()
					jsonBytes, _ := json.Marshal(p.Args)
					json.Unmarshal(jsonBytes, &structInterface)
					args = append(args, reflect.Indirect(reflect.ValueOf(structInterface)))
				case fieldResolverArgContext:
					args = append(args, reflect.ValueOf(p.Context))
				case fieldResolverArgInfo:
					args = append(args, reflect.ValueOf(p.Info))
				}
			}

			response := method.Func.Call(args)
			value, err := response[0], response[1]

			if isUnion {
				p := graphql.ResolveTypeParams{
					Value:   value.Interface(),
					Info:    p.Info,
					Context: p.Context,
				}

				value = union.resolveValue(p)
			}

			if err.IsNil() {
				return value.Interface(), nil
			}

			return value.Interface(), err.Interface().(error)
		}
	} else if _, ok := GetNullable(grootType).(*Interface); ok {
		msg := fmt.Sprintf(
			"field %s on struct %s of interface type must have a resolver function",
			structField.Name,
			structType.Name(),
		)
		panic(msg)
	}

	field := &Field{
		name:              name,
		description:       description,
		type_:             grootType,
		resolve:           resolver,
		deprecationReason: depracationReason,
		arguments:         arguments,
	}

	field.field = field.GraphQLField()
	return field
}
