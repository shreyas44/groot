package gql

import (
	"github.com/graphql-go/handler"
)

func NewHandler(config handler.Config) *handler.Handler {
	handler := handler.New(&config)
	return handler
}