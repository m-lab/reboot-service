package creds

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/api/option"

	"cloud.google.com/go/datastore"
)

type mockConnector struct {
	mustFail bool
}

func (c *mockConnector) NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (client, error) {
	if c.mustFail {
		return nil, errors.New("NewClient() method failed")
	}
	return mc, nil
}

// mockDatastoreClient is a fake DatastoreClient for testing.
type mockClient struct {
	Creds      []*Credentials
	mustFail   bool
	skipAppend bool
}

func (d *mockClient) GetAll(ctx context.Context, q *datastore.Query,
	dst interface{}) ([]*datastore.Key, error) {

	if d.mustFail {
		return nil, errors.New("method GetAll failed")
	}

	if !d.skipAppend {
		creds := dst.(*[]*Credentials)
		*creds = append(*creds, fakeDrac)
	}

	return nil, nil
}

const (
	testHost      = "test"
	testUser      = "user"
	testPass      = "pass"
	testModel     = "drac"
	testAddress   = "addr"
	testNamespace = "test"
)

var fakeDrac = &Credentials{
	Hostname: "host",
	Username: "user",
	Password: "pass",
	Model:    "model",
	Address:  "address",
}

var mc = &mockClient{
	Creds: []*Credentials{
		fakeDrac,
	},
}

func TestNewProvider(t *testing.T) {
	_, err := NewProvider("projectID", "ns", "test")
	if err != nil {
		t.Errorf("NewProvider() unexpected error: %v", err)
	}
}

func Test_datastoreProvider_FindCredentials(t *testing.T) {
	// Inject mockConnector to simulate network failures.
	connector := &mockConnector{}
	provider := &datastoreProvider{
		connector: connector,
		kind:      "test",
		namespace: "ns",
		projectID: "projectID",
	}

	// FindCredentials() should return a Credentials for a known host.
	creds, err := provider.FindCredentials(context.Background(), "testhost")
	if err != nil {
		t.Errorf("FindCredentials() unexpected error")
	}
	if *creds != *fakeDrac {
		t.Errorf("FindCredentials() didn't return the expected Credential")
	}

	// FindCredentials() should fail if the connection fails.
	connector.mustFail = true
	_, err = provider.FindCredentials(context.Background(), "testhost")
	if err == nil {
		t.Errorf("FindCredentials() expected error, got nil.")
	}
	connector.mustFail = false

	// FindCredentials() should fail if the query fails.
	mc.mustFail = true
	_, err = provider.FindCredentials(context.Background(), "testhost")
	if err == nil {
		t.Errorf("FindCredentials() expected error, got nil.")
	}
	mc.mustFail = false

	// FindCredentials() should fail if there is no result for a known host.
	mc.skipAppend = true
	_, err = provider.FindCredentials(context.Background(), "testhost")
	if err == nil {
		t.Errorf("FindCredentials() expected error, got nil.")
	}
	mc.skipAppend = false

}