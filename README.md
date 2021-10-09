# Groot

An easier way to create GraphQL APIs in Go. View the full documentation [here](https://groot.shreyas44.com).

Groot is built on top of [`github.com/graphql-go/graphql`](https://github.com/graphql-go/graphql), which means it should support most existing tooling built for it.

## Motivation

Go already has a couple of implementation of GraphQL, so why another one?

### Type Safety

Go is statically typed, and GraphQL is type safe, which means we don't need to and shouldn't use `interface{}` anywhere. A simple user struct with custom resolvers would look something like this. Although most type checking is done by Go, additional checks like [resolver](./type-definitions/field-resolvers) return types are done by Groot on startup to avoid type errors altogether.

```go
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
```

### Code First

Although the schema first aproach has its advantages, code first is arguably easier to maintain in the long run without having to deal with federated schemas and schema stitching. For more info check out [this blog post](https://blog.logrocket.com/code-first-vs-schema-first-development-graphql/).

When you work with Groot, in a way you're defining your schema first as well since you're defining the structure of your data (struct, interfaces, enums, etc) first.

### No Boilerplate and Code Duplication

The only thing we want to worry about is our types, resolvers, and business logic, nothing more. We also don't want to redeclare our types in Go as well as GraphQL, it can get cumbersome to maintain and keep track of.

### Simple To Use

Seriously, it is.

## Getting Started

Let's see how we can create the below GraphQL schema.

```graphql
type User {
  id: ID!
  name: String!
  email: String!
  posts: [Post!]
}

type Post {
  id: ID!
  title: String!
  body: String!
  author: User!
  timestamp: Int!
}

type Query {
  user(id: ID!): User
  post(id: ID!): Post
}

type Mutation {
  createUser(name: String!, email: String!, password: String!): User
  createPost(title: String!, body: String!): Post
}
```

We can define `User` and `Post` objects as regular structs.

```go
type User struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Email string  `json:"email"`
	// we use a pointer to make the field nullable
	Posts *[]Post `json:"posts"`
}

type Post struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Author    User   `json:"author"`
	Timestamp int64  `json:"timestamp"`
}
```

Yes, that's it!

### Custom Resolvers

The library will automatically generate the resolvers for all the fields. But, you can create custom resolvers for any field by defining a method with the name `Resolve{field-name}`. For example, if we want to define a custom resolver for the `Posts` field on `User` we can write the below method:

```go
func (user User) ResolvePosts() (*[]Post, error) {
	posts, err := db.GetPostsByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	return posts, nil
}
```

If the return type of the resolver is not the same as the return type of the field, Groot will panic on startup. For more details on resolvers and arguments check the [Field Resolvers](./type-definitions/field-resolvers) section.

### Queries

The `Query` type is just another type with a special name, which means we can define it just like we defined the `User` and `Post` types with custom resolvers.

```go
type Query struct {
	// we use pointers to make the field nullable
	User *User `json:"user"`
	Post *Post `json:"post"`
}

type IDArgs struct {
	ID string `json:"id"`
}

func (q Query) ResolveUser(args IDArgs) (*User, error) {
	user, err := db.GetUserByID(args.ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (p Query) ResolvePost(args IDArgs) (*Post, error) {
	post, err := db.GetPostByID(args.ID)
	if err != nil {
		return nil, err
	}
	return post, nil
}
```

Notice how we were able to accept arguments by having the type of first argument of the resolver as a struct.

For a larger schema, you may think we would need to define a lot of fields and methods on the single `Query` type. While you would be right, we can avoid that by just [embedding structs](https://www.geeksforgeeks.org/composition-in-golangj/). You can find more info on embedding and composition in the [Composition](./composition) section.

### Mutations

Similar to the `Query` struct we can define a `Mutation` struct to define our mutations.

```go
type Mutation struct {
	CreateUser *User `json:"createUser"`
	CreatePost *Post `json:"createPost"`
}

type NewUserArgs struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type NewPostArgs struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (m Mutation) ResolveCreateUser(args NewUserArgs) (*User, error) {
	user, err := db.CreateUser(args.Name, args.Email, args.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (m Mutation) ResolveCreatePost(args NewPostArgs) (*Post, error) {
	post, err := db.CreatePost(args.Title, args.Body)
	if err != nil {
		return nil, err
	}
	return post, nil
}
```

### Creating Schema

Finally, to create the schema, we can use the `NewSchema` function. We can also use the [`github.com/graphql-go/handler`](https://github.com/graphql-go/handler) library to create a handler for the schema since `groot.NewSchema` returns a schema of type `graphql.Schema` where the `graphql` package refers to the `github.com/graphql-go/graphql` library.

```go
import (
	"reflect"
	"net/http"
	"github.com/shreyas44/groot"
	"github.com/graphql-go/handler"
)

func main() {
	schema := groot.NewSchema(groot.SchemaConfig{
		Query:    groot.MustParseObject(Query{}),
		Mutation: groot.MustParseObject(Mutation{}),
	})

	h := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
		Playground: true,
	})

	http.Handle("/graphql", h)
	log.Fatal(http.ListenAndServe(":8080", nil)
}
```

---

#### Features Not Supported but Coming Soon:

- [Custom Directives](https://github.com/shreyas44/groot/issues/4)
- [Descriptions for type definitions](https://github.com/shreyas44/groot/issues/2)
- [Enum value description and deprecation](https://github.com/shreyas44/groot/issues/2)

_Note, the library isn't tested yet._
