package groot

import "github.com/graphql-go/graphql"

type Interface struct {
	Name             string
	Description      string
	fields           []*Field
	privateInterface *graphql.Interface
}

func (i *Interface) GraphQLType() *graphql.Interface {
	if i.privateInterface != nil {
		return i.privateInterface
	}

	graphqlFields := FieldsFromFields(i.fields)
	i.privateInterface = graphql.NewInterface(graphql.InterfaceConfig{
		Name:        i.Name,
		Description: i.Description,
		Fields:      graphqlFields,
		// ResolveType: ,
	})

	return i.privateInterface
}

func InterfacesFromInterfaces(interfaces []*Interface) []*graphql.Interface {
	graphqlInterfaces := []*graphql.Interface{}

	for _, i := range interfaces {
		graphqlInterface := i.GraphQLType()
		graphqlInterfaces = append(graphqlInterfaces, graphqlInterface)
	}

	return graphqlInterfaces
}
