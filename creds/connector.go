package creds

import (
	"context"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
)

// Client is an interface to make testing possible. The default
// implementation is the actual *datastore.Client as returned by
// datastore.NewClient.
type client interface {
	GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error)
	Put(context.Context, *datastore.Key, interface{}) (*datastore.Key, error)
	Close() error
}

type connector interface {
	NewClient(context.Context, string, ...option.ClientOption) (client, error)
}

type datastoreConnector struct{}

func (d *datastoreConnector) NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (client, error) {
	return datastore.NewClient(ctx, projectID, opts...)
}
