# Input Objects

[Input Objects](https://graphql.org/learn/schema/#input-types) can be defined using a regular Go struct simiar to [Object](./object) types.

```go
type NewUserInput struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}
```

For more info on field definitions, see [Field Definitions](./field-definitions).

### Using Input Objects

To use the input object, use it as a regular type in the struct definition used for an argument.

```go
type NewUserArgs struct {
	Input NewUserInput `json:"input"`
}

func (m Mutation) ResolveNewUser(args NewUserArgs) (User, error) {
	// create new user
}
```

This would create the below schema

```graphql
input NewUserInput {
  name: String!
  username: String!
  password: String!
}

type Mutation {
  newUser(input: NewUserInput!): User
}
```
