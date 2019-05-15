package creds

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
)

type mockConnector struct {
	client   client
	mustFail bool
}

func (c *mockConnector) NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (client, error) {
	if c.mustFail {
		return nil, errors.New("NewClient() method failed")
	}
	return c.client, nil
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
		for _, cred := range d.Creds {
			*creds = append(*creds, cred)
		}
	}

	return nil, nil
}

func TestNewProvider(t *testing.T) {
	provider := NewProvider("projectID", "ns")
	if provider == nil {
		t.Errorf("NewProvider() returned nil.")
	}
}

func TestFindCredentials(t *testing.T) {
	// Create a mockClient returning fake Credentials.
	fakeDrac := &Credentials{
		Hostname: "host",
		Username: "user",
		Password: "pass",
		Model:    "model",
		Address:  "address",
	}
	mc := &mockClient{
		Creds: []*Credentials{
			fakeDrac,
		},
	}

	// Inject mockConnector to simulate network failures.
	connector := &mockConnector{
		client: mc,
	}
	provider := &datastoreProvider{
		connector: connector,
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
