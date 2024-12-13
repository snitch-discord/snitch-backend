// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	AddServerToGroup(ctx context.Context, arg AddServerToGroupParams) error
	FindGroupIDByServerID(ctx context.Context, serverID int) (uuid.UUID, error)
	InsertGroup(ctx context.Context, arg InsertGroupParams) error
}

var _ Querier = (*Queries)(nil)
