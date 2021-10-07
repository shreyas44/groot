# Scalars

You can use any of the below Go types to represent the corresponding GraphQL scalar.

1. `Int` - `int, int8, int16, int32, uint, uint8, uint16`
2. `Float` - `float32, float64`
3. `Boolean` - `bool`
4. `String` - `string`
5. `ID` - `graphql.StringID, graphql.IntID`

_Note, `int64, uint32, uint64` are not supported since the built in Int type in GraphQL is a 32 bit integer._

### Custom Scalars

You can create a custom [Scalar](https://graphql.org/learn/schema/#scalar-types) by implementing the `groot.ScalarType` interface on the pointer to that type.

```go
type ScalarType struct {
	json.Marshaler
	json.Unmarshaler
}

// OR

type ScalarType struct {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
}
```

### Example with Time

```go
type Time time.Time

func (t *Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*t).Unix())
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var unix int64
	err := json.Unmarshal(b, &unix)
	if err != nil {
		return err
	}

	// using *t here is necessary since we need to change
	// the value at the address t points to
	*t = Time(time.Unix(unix, 0))
	return nil
}
```

**Keep in mind `*Time` should implement `groot.ScalarType`, not `Time`.**
