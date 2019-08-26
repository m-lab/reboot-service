package creds

import (
	"context"
	"testing"
)

func Test_datastoreConnector_NewClient(t *testing.T) {
	connector := DatastoreConnector{}
	connector.NewClient(context.Background(), "testproject")
}
