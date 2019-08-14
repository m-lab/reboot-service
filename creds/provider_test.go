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

func (d *mockClient) Put(context.Context, *datastore.Key,
	interface{}) (*datastore.Key, error) {
	if d.mustFail {
		return nil, errors.New("method Put failed")
	}

	return nil, nil
}

func (d *mockClient) Delete(ctx context.Context, key *datastore.Key) error {
	if d.mustFail {
		return errors.New("method Delete failed")
	}
	return nil
}

func (d *mockClient) Close() error {
	return nil
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

func TestAddCredentials(t *testing.T) {
	fakeDrac := &Credentials{
		Hostname: "host",
		Username: "user",
		Password: "pass",
		Model:    "model",
		Address:  "address",
	}

	// Create a datastoreProvider with a mock connector and client.
	mc := &mockClient{}
	connector := &mockConnector{
		client: mc,
	}
	provider := &datastoreProvider{
		connector: connector,
		namespace: "ns",
		projectID: "projectID",
	}

	err := provider.AddCredentials(context.Background(), "testhost", fakeDrac)
	if err != nil {
		t.Errorf("AddCredentials() unexpected error.")
	}

	// AddCredentials() should fail if the connection fails.
	connector.mustFail = true
	err = provider.AddCredentials(context.Background(), "testhost", fakeDrac)
	connector.mustFail = false
	if err == nil {
		t.Errorf("AddCredentials() expected error, got nil.")
	}

	// AddCredentials() should fail if the Put() fails.
	mc.mustFail = true
	err = provider.AddCredentials(context.Background(), "testhost", fakeDrac)
	mc.mustFail = false
	if err == nil {
		t.Errorf("AddCredentials() expected error, got nil.")
	}
}

func TestCredentials_String(t *testing.T) {
	creds := &Credentials{
		Address:  "127.0.0.1",
		Username: "username",
		Password: "!\"£$%^&*()_+-=",
		Model:    "DRAC",
		Hostname: "mlab1d.lga0t.measurement-lab.org",
	}

	expected := `{
  "hostname": "mlab1d.lga0t.measurement-lab.org",
  "username": "username",
  "password": "!\"£$%^&*()_+-=",
  "model": "DRAC",
  "address": "127.0.0.1"
}
`

	if creds.String() != expected {
		t.Errorf("Credentials.String() didn't return the expected output.")
	}

}

func Test_datastoreProvider_deleteCredentials(t *testing.T) {
	fakeDrac := &Credentials{
		Hostname: "testhost",
		Username: "user",
		Password: "pass",
		Model:    "model",
		Address:  "address",
	}

	// Create a datastoreProvider with a mock connector and client.
	mc := &mockClient{}
	connector := &mockConnector{
		client: mc,
	}
	provider := &datastoreProvider{
		connector: connector,
		namespace: "ns",
		projectID: "projectID",
	}

	err := provider.AddCredentials(context.Background(), "testhost", fakeDrac)
	if err != nil {
		t.Errorf("AddCredentials() unexpected error.")
	}

	err = provider.DeleteCredentials(context.Background(), "testhost")
	if err != nil {
		t.Errorf("DeleteCredentials() returned error: %v", err)
	}

	// DeleteCredentials() should fail if the connection fails.
	connector.mustFail = true
	err = provider.DeleteCredentials(context.Background(), "testhost")
	connector.mustFail = false
	if err == nil {
		t.Errorf("DeleteCredentials() expected error, got nil.")
	}

	// DeleteCredentials() should fail if the Delete() fails.
	mc.mustFail = true
	err = provider.DeleteCredentials(context.Background(), "testhost")
	mc.mustFail = false
	if err == nil {
		t.Errorf("DeleteCredentials() expected error, got nil.")
	}
}
