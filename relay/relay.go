package relay

import (
	"errors"

	"github.com/shreyas44/groot"
)

type NodeDefinition struct {
	groot.InterfaceType
	ID groot.StringID `json:"id"`
}

func NewNodeDefinition(id string) NodeDefinition {
	return NodeDefinition{
		ID: groot.StringID(id),
	}
}

type Node interface {
	ImplementsNode() NodeDefinition
}

func (d NodeDefinition) ImplementsNode() NodeDefinition {
	return d
}

type PageInfo struct {
	HasPreviousPage bool   `json:"hasPreviousPage"`
	HasNextPage     bool   `json:"hasNextPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

type PaginationArgs struct {
	First  *int    `json:"first"`
	Last   *int    `json:"last"`
	After  *string `json:"after"`
	Before *string `json:"before"`
}

func (args PaginationArgs) Validate() error {
	if args.First == nil && args.Last == nil {
		return errors.New("first or last must be specified")
	}

	if args.First != nil && args.Last != nil {
		return errors.New("first and last cannot be set at the same time")
	}

	if *args.First < 0 || *args.Last < 0 {
		return errors.New("first and last must be greater than 0")
	}

	if args.After != nil && args.Before != nil {
		return errors.New("after and before cannot be set at the same time")
	}

	if args.First != nil && args.Before != nil {
		return errors.New("first and before cannot be set at the same time")
	}

	if args.Last != nil && args.After != nil {
		return errors.New("last and after cannot be set at the same time")
	}

	return nil
}

// We can add these once we have generics in Go

// type ConnectionEdge struct {
// 	Cursor string `json:"cursor"`
// 	Node   Node   `json:"node"`
// }

// type Connection struct {
// 	PageInfo PageInfo         `json:"pageInfo"`
// 	Edges    []ConnectionEdge `json:"edges"`
// }
