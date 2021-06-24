package groot

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

type Field struct {
	Name              string
	Description       string
	DeprecationReason string
	Type              graphql.Type
	Arguments         []*Argument
	Resolve           func(p graphql.ResolveParams) (interface{}, error)
	field             *graphql.Field
}

type resolverArguments struct {
	arguments bool
	context   bool
	info      bool
}

func (field *Field) GraphQLType() *graphql.Field {
	if field.field != nil {
		return field.field
	}

	args := graphql.FieldConfigArgument{}

	for _, argument := range field.Arguments {
		args[argument.Name] = argument.GraphQLType()
	}

	field.field = &graphql.Field{
		Name:              field.Name,
		Type:              field.Type,
		Description:       field.Description,
		Resolve:           field.Resolve,
		DeprecationReason: field.DeprecationReason,
		Args:              args,
	}

	return field.field
}

func GetArguments(t reflect.Type) []*Argument {
	arguments := []*Argument{}
	if t.Kind() != reflect.Struct {
		panic("argument type must be a struct")
	}

	fieldCount := t.NumField()
	for i := 0; i < fieldCount; i++ {
		fieldType := t.Field(i)
		argument := NewArgument(fieldType)
		arguments = append(arguments, argument)
	}

	return arguments
}

func NewField(structField reflect.StructField, structType reflect.Type) *Field {
	var name string
	var description string
	var depractionReason string
	var arguments []*Argument
	var resolver func(p graphql.ResolveParams) (interface{}, error)
	resolverArguments := [3]bool{}

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
		depractionReason = deprecate
	}

	graphqlType, ok := graphqlTypes[structField.Type]
	if structFieldType := structField.Type; !ok && structFieldType.Kind() == reflect.Struct {
		graphqlTypes[structField.Type] = nil
		object := NewObject(structFieldType)
		graphqlType = object.GraphQLType()
	}

	// default resolver
	resolver = func(p graphql.ResolveParams) (interface{}, error) {
		return p.Source.(reflect.Value).FieldByName(structField.Name), nil
	}

	// custom resolver
	if method, exists := structType.MethodByName(fmt.Sprintf("Resolve%s", structField.Name)); exists {
		methodType := method.Func.Type()

		// type check resolver
		outCount := methodType.NumOut()
		inCount := methodType.NumIn()

		if outCount != 2 {
			panic(
				fmt.Sprintf("return type of (%s, error) was expected for resolver %s", structType.Name(), method.Name),
			)
		}

		// credits - https://stackoverflow.com/questions/30688514/go-reflect-how-to-check-whether-reflect-type-is-an-error-type/30688564
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if methodType.Out(0) != structField.Type || !methodType.Out(1).Implements(errorInterface) {
			message := fmt.Sprintf(
				"return type of (%s, error) was expected for resolver %s, got (%s, %s)",
				structField.Type.Name(),
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

		// ignore first input as that's the struct the method is acting on
		// TODO: fix logic to decide arguments of resolver
		for i := 1; i < inCount; i++ {
			if methodType.In(i) == reflect.TypeOf(graphql.ResolveInfo{}) {
				resolverArguments[2] = true
			} else if methodType.In(i).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
				resolverArguments[1] = true
			} else if methodType.In(i).Kind() == reflect.Struct {
				resolverArguments[0] = true
			}
		}

		if resolverArguments[0] {
			arguments = GetArguments(methodType.In(1))
		}

		resolver = func(p graphql.ResolveParams) (interface{}, error) {
			var source reflect.Value
			args := []reflect.Value{}

			// if it's a map, it's a root query
			if _, isMap := p.Source.(map[string]interface{}); isMap {
				source = reflect.Indirect(reflect.New(structType))
			} else {
				source = p.Source.(reflect.Value).Convert(structType)
			}

			args = append(args, source)

			if resolverArguments[0] {
				structInterface := reflect.New(methodType.In(1)).Interface()
				jsonBytes, _ := json.Marshal(p.Args)
				json.Unmarshal(jsonBytes, &structInterface)

				args = append(args, reflect.Indirect(reflect.ValueOf(structInterface)))
			}

			if resolverArguments[1] {
				args = append(args, reflect.ValueOf(p.Context))
			}

			if resolverArguments[2] {
				args = append(args, reflect.ValueOf(p.Info))
			}

			response := method.Func.Call(args)
			return response[0], nil
		}
	}

	field := &Field{
		Name:              name,
		Description:       description,
		Type:              graphqlType,
		Resolve:           resolver,
		DeprecationReason: depractionReason,
		Arguments:         arguments,
	}

	// hydrate field.field with *graphql.Field
	field.GraphQLType()

	return field
}

func FieldsFromFields(fields []*Field) []*graphql.Field {
	graphqlFields := []*graphql.Field{}

	for _, field := range fields {
		graphqlField := field.GraphQLType()
		graphqlFields = append(graphqlFields, graphqlField)
	}

	return graphqlFields
}
