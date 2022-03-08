package groot

import (
	"encoding/json"
	"math/big"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/shreyas44/groot/parser"
)

type ScalarType = parser.ScalarType

type ID string

var builtinScalars = map[reflect.Type]*graphql.Scalar{
	reflect.TypeOf(ID("")):       graphql.ID,
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

func NewScalar(parserScalar *parser.Scalar, builder *SchemaBuilder) *graphql.Scalar {
	if graphqlScalar, ok := builtinScalars[parserScalar.ReflectType()]; ok {
		return graphqlScalar
	}

	// TODO: description
	scalar := graphql.NewScalar(graphql.ScalarConfig{
		Name: parserScalar.ReflectType().Name(),
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

			v := reflect.New(parserScalar.ReflectType()).Interface().(ScalarType)
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

			v := reflect.New(parserScalar.ReflectType()).Interface().(ScalarType)
			err = v.UnmarshalJSON(jsonRepr)
			if err != nil {
				panic(err)
			}

			return v
		},
	})

	builder.addType(parserScalar, scalar)
	return scalar
}

func astValueToGoValue(valueAST ast.Value) (interface{}, error) {
	var value interface{}

	// TODO: determine performance implications of using big.Int and big.Float

	switch valueAST := valueAST.(type) {
	case *ast.IntValue:
		num := new(big.Int)
		num.SetString(valueAST.Value, 10)
		value = num

	case *ast.FloatValue:
		num := new(big.Float)
		num.SetString(valueAST.Value)
		value = num

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
