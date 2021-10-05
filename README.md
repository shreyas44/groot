# Groot

A new and easier method to create GraphQL APIs in Go

Currently, the main implementation of GraphQL in Go is https://github.com/graphql-go/graphql. However, creating GraphQL APIs using that package can get quite tedious and messy.

Let's look at an example to create a simple GraphQL Schema

```graphql
enum UserType {
  ADMIN
  USER
}

type User {
  id: String!
  name: String!
  type: UserType!
  oldType: UserType! @deprecated(reason: "Old Field")
}

type Post {
  id: String!
  body: String
  author: User!
  # When the post was posted
  timestamp: Int!
}

type Query {
  user(id: String!) User
  post(id: String!) Post
}
```

What the code looks like with `github.com/graphql-go/graphql`

```go
func main() {
	userTypeEnum := graphql.NewEnum(graphql.EnumConfig{
		Name: "UserType",
		Values: graphql.EnumValueConfigMap{
			"ADMIN": &graphql.EnumValueConfig{
				Value: "ADMIN",
			},
			"USER": &graphql.EnumValueConfig{
				Value: "USER",
			},
		},
	})

	userObject := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "anid", nil
				},
			},
			"name": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "A Name", nil
				},
			},
			"type": &graphql.Field{
				Type: graphql.NewNonNull(UserTypeEnum),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "ADMIN", nil
				},
			},
			"oldType": &graphql.Field{
				Type: graphql.NewNonNull(UserTypeEnum),
				DeprecationReason: "Old Field",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "ADMIN", nil
				}
			},
		},
	})

	postObject := graphql.NewObject(graphql.ObjectConfig{
		Name: "Post",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "anid", nil
				},
			},
			"body": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "a body", nil
				},
			},
			"timestamp": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return 12345, nil
				},
				Description: "When the post was posted",
			},
			"author": &graphql.Field{
				Type: graphql.NewNonNull(userObject),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			},
		},
	})

	queryObject := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"user": &graphql.Field{
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Type: userObject,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return db.User.Get(p.Args["id"])
				},
			},
			"post": &graphql.Field{
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Type: postObject,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return db.Post.Get(p.Args["id"])
				},
			},
		},
	})

	schema := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryObject,
	})
}
```

As you can see, there are quite a few drawbacks to doing this. First, `interface{}`'s are used everywhere; we're not taking advantage of the type system in Go. Second, the code might become unnecessarily lengthy.

In contrast, here's how you'd do it with `Groot`

```go

type UserType string

const (
	UserTypeAdmin UserType = "ADMIN"
	UserTypeUser  UserType = "USER"
)

func (u UserType) Values() []string {
	return []string{string(UserTypeAdmin), string(UserTypeUser)}
}

type User struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Type UserType `json:"type"`
	OldType UserType `json:"oldType" deprecate:"Old Field"`
}

type Post struct {
	ID        string  `json:"id"`
	// you can use a pointer to mark the field as nullable
	Body      *string `json:"body"`
	Timestamp int     `json:"timestamp" description:"When the post was posted"`
	Author    User    `json:"author"`
}

type Query struct {
	User *User `json:"user"`
	Post *Post `json:"post"`
}

type IdArgs struct {
	ID string `json:"id"`
}

func (query Query) ResolveUser(args IdArgs) (User, error) {
	return db.User.Get(args.ID), nil
}

// you can either include or omit the arguments, context, and info based on the needs of your resolver.
func (query Query) ResolvePost(args IdArgs, context context.Context, info graphql.ResolveInfo) (Post, error) {
	return db.Post.Get(args.ID), nil
}

func main() {
	schema := groot.NewSchema(groot.SchemaConfig{
		Query: reflect.TypeOf(Query{})
	})
}
```

As you can see, the code is much more readable, and takes full advantage of the Go type system. Groot also comes with default resolvers, and type checks the resolvers on startup. It's also intercompatible with `gtihub.com/graphql-go/graphql` since it uses `github.com/graphql-go/graphql` under the hood, and just provides a really nice abstraction on top of it. This means almost all extensions and libraries meant to be used with `github.com/graphql-go/graphql` can continue to be used with Groot.

No generated code, no boilerplate, composable, and type safe!

---

### Features Not Yet Supported

- [ ] Subscriptions
- [ ] Custom Scalars
- [ ] Descriptions for type definitions
- [ ] Enum value description and deprecation

### Note

- The only case where there's no type safety or type checking either by Go or Groot is for GraphQL interfaces
- Although almost all use cases are covered (except the ones mentioned above), the library isn't teststed yet. If there's anything else missing please open an issue!
