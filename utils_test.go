package groot

import (
	"reflect"
	"testing"
)

type User struct {
	Id string
	Name string
}

type Post struct {
	Id string
	Body string
	Timestamp int
	Author User
}

func TestGetTypes(t *testing.T) {
	structs := []interface{}{User{}, Post{}}
	types, err := GetTypes(structs...)

	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < len(structs); i++ {
		if structType := reflect.TypeOf(structs[i]); structType != types[i] {
			t.Fatalf("got invalid type %v, expected %v", types[i].Name(), structType.Name())
		}
	}
}

func TestNonStructError(t *testing.T) {
	testTypes := []interface{}{"test", 0123, 0.00}

	for _, testType := range testTypes {
		_, err := GetType(testType)

		if err == nil {
			t.Fatalf("expected error for type %v but did not receive it", reflect.TypeOf(testType))
		}
	}
}