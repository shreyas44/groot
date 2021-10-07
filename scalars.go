package groot

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

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

func NewScalar(t reflect.Type, builder *SchemaBuilder) (*Scalar, error) {
	parserType, _ := getParserType(t)
	if parserType != ParserScalar && parserType != ParserCustomScalar {
		err := fmt.Errorf(
			"groot: reflect.Type %s passed to NewScalar must have parser type ParserScalar, received %s",
			t.Name(),
			parserType,
		)
		panic(err)
	}

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

	if parserType == ParserScalar {
		scalar.scalar = scalars[t]
	} else if parserType == ParserCustomScalar {
		t = reflect.PtrTo(t)
	}

	builder.grootTypes[t] = scalar
	return scalar, nil
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

			v := reflect.New(scalar.reflectType).Interface().(ScalarType)
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

			v := reflect.New(scalar.reflectType).Interface().(ScalarType)
			err = v.UnmarshalJSON(jsonRepr)
			if err != nil {
				panic(err)
			}

			return v
		},
	})

	return scalar.scalar
}

func (scalar *Scalar) ReflectType() reflect.Type {
	return scalar.reflectType
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
