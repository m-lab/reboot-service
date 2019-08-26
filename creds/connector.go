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
	Delete(context.Context, *datastore.Key) error
	Close() error
}

// Connector is an interface to abstract a new Client creation.
type Connector interface {
	NewClient(context.Context, string, ...option.ClientOption) (client, error)
}

// DatastoreConnector is the default implementation of a Connector. It allows to
// create a new datastore Client.
type DatastoreConnector struct{}

// NewClient returns a datastore Client with the provided configuration.
func (d *DatastoreConnector) NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (client, error) {
	return datastore.NewClient(ctx, projectID, opts...)
}
