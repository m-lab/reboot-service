package iface

import (
	"context"

	"cloud.google.com/go/datastore"
)

// DatastoreClient is an interface to make testing possible. The default
// implementation is the actual *datastore.Client as returned by
// datastore.NewClient.
type DatastoreClient interface {
	GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error)
}
