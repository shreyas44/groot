package groot

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

type (
	StringID string
	IntID    int
)

type ScalarType interface {
	json.Marshaler
	json.Unmarshaler
}

type Scalar struct {
	name        string
	description string
	scalar      *graphql.Scalar
	reflectType reflect.Type
}

func astValueToGoValue(valueAST ast.Value) (interface{}, error) {
	var value interface{}

	switch valueAST := valueAST.(type) {
	case *ast.IntValue:
		v, err := strconv.Atoi(valueAST.GetValue().(string))
		if err != nil {
			panic(err)
		}

		value = v

	case *ast.FloatValue:
		v, err := strconv.ParseFloat(valueAST.GetValue().(string), 64)
		if err != nil {
			panic(err)
		}

		value = v

	case *ast.StringValue, *ast.BooleanValue, *ast.EnumValue:
		value = valueAST.GetValue()

	case *ast.ListValue:
		list := []interface{}{}
		for _, listValue := range valueAST.Values {
			goValue, err := astValueToGoValue(listValue)
			if err != nil {
				return nil, err
			}

			list = append(list, goValue)
		}

		value = list

	case *ast.ObjectValue:
		object := map[string]interface{}{}
		for _, objectField := range valueAST.Fields {
			goValue, err := astValueToGoValue(objectField.Value)
			if err != nil {
				return nil, err
			}

			object[objectField.Name.Value] = goValue
		}

		value = object
	}

	return value, nil
}

func astValueToJSON(valueAST ast.Value) ([]byte, error) {
	goValue, err := astValueToGoValue(valueAST)
	if err != nil {
		return nil, err
	}

	return json.Marshal(goValue)
}

func (scalar *Scalar) GraphQLType() graphql.Type {
	if scalar.scalar != nil {
		return scalar.scalar
	}

	scalar.scalar = graphql.NewScalar(graphql.ScalarConfig{
		Name:        scalar.name,
		Description: scalar.description,
		Serialize: func(value interface{}) interface{} {
			var v ScalarType
			if reflect.TypeOf(value).Kind() != reflect.Ptr {
				newValue := reflect.New(reflect.TypeOf(value))
				newValue.Elem().Set(reflect.ValueOf(value))
				v = newValue.Interface().(ScalarType)
			} else {
				v = reflect.ValueOf(value).Interface().(ScalarType)
			}

			return v
		},
		ParseLiteral: func(valueAST ast.Value) interface{} {
			jsonRepr, err := astValueToJSON(valueAST)
			if err != nil {
				panic(err)
			}

			v := reflect.New(scalar.reflectType.Elem()).Interface().(ScalarType)
			err = v.UnmarshalJSON(jsonRepr)
			if err != nil {
				panic(err)
			}

			return v
		},
		ParseValue: func(value interface{}) interface{} {
			jsonRepr, err := json.Marshal(value)
			if err != nil {
				panic(err)
			}

			v := reflect.New(scalar.reflectType.Elem()).Interface().(ScalarType)
			err = v.UnmarshalJSON(jsonRepr)
			if err != nil {
				panic(err)
			}

			return v
		},
	})

	return scalar.scalar
}

func NewScalar(t reflect.Type, builder *SchemaBuilder) *Scalar {
	scalars := map[reflect.Type]*graphql.Scalar{
		reflect.TypeOf(IntID(0)):     graphql.ID,
		reflect.TypeOf(StringID("")): graphql.ID,
		reflect.TypeOf(int(0)):       graphql.Int,
		reflect.TypeOf(int8(0)):      graphql.Int,
		reflect.TypeOf(int16(0)):     graphql.Int,
		reflect.TypeOf(int32(0)):     graphql.Int,
		reflect.TypeOf(uint(0)):      graphql.Int,
		reflect.TypeOf(uint8(0)):     graphql.Int,
		reflect.TypeOf(uint16(0)):    graphql.Int,
		reflect.TypeOf(float32(0.0)): graphql.Float,
		reflect.TypeOf(float64(0.0)): graphql.Float,
		reflect.TypeOf(""):           graphql.String,
		reflect.TypeOf(false):        graphql.Boolean,
	}

	scalar := &Scalar{
		name:        t.Name(),
		reflectType: t,
	}

	if _, ok := scalars[t]; ok {
		scalar.scalar = scalars[t]
	} else if t.Kind() != reflect.Ptr {
		panic("reflect.Type passed to scalar must be a pointer")
	} else {
		scalar.name = t.Elem().Name()
	}

	builder.types[t.Name()] = scalar.GraphQLType()
	builder.grootTypes[t] = scalar
	return scalar
}
