# Enums

To create an [enum](https://graphql.org/learn/schema/#enumeration-types) you need to create a new type of type `string` which implements the `groot.EnumType` interface where the `Values` method should return all the values of the enum.

```go
type EnumType interface {
	Values() []string
}
```

To define an enum `UserType` with values `ADMIN` and `USER`, we can use the below code:

```go
type UserType string

const (
	UserTypeAdmin UserType = "ADMIN"
	UserTypeUser  UserType = "USER"
)

func (u UserType) Values() []string {
	return []string{string(UserTypeAdmin), string(UserTypeUser)}
}
```

You can then use `UserType` as a regluar type on any field.

_Note, if you don't implement the `EnumType` interface, Groot will treat it as a regular string instead of an enum._
