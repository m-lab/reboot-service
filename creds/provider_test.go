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
	if d.mustFail {
		return errors.New("method Close failed")
	}
	return nil
}

func TestNewProvider(t *testing.T) {
	mc := &mockClient{}
	connector := &mockConnector{
		client: mc,
	}
	provider, err := NewProvider(connector, "projectID", "ns")
	if err != nil {
		t.Errorf("NewProvider() returned err: %v", err)
	}
	if provider == nil {
		t.Errorf("NewProvider() returned a nil provider.")
	}

	// Simulate a failure during client initialization.
	connector.mustFail = true
	provider, err = NewProvider(connector, "projectID", "ns")
	if err == nil {
		t.Errorf("NewProvider() expected err, got nil.")
	}
}

func TestListCredentials(t *testing.T) {
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
	provider := &datastoreProvider{
		namespace: "ns",
		projectID: "projectID",
		client:    mc,
	}

	// ListCredentials() should return the known Credentials.
	creds, err := provider.ListCredentials(context.Background())
	if err != nil {
		t.Errorf("ListCredentials() unexpected error")
	}
	if len(creds) != 1 {
		t.Errorf("ListCredentials() returned a slice of the wrong size.")
	}

	if *creds[0] != *fakeDrac {
		t.Errorf("ListCredentials() didn't return the expected Credentials.")
	}

	// ListCredentials should fail if client.GetAll fails.
	mc.mustFail = true
	creds, err = provider.ListCredentials(context.Background())
	if err == nil {
		t.Errorf("ListCredentials() didn't return an error.")
	}
	mc.mustFail = false

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

	provider := &datastoreProvider{
		namespace: "ns",
		projectID: "projectID",
		client:    mc,
	}

	// FindCredentials() should return a Credentials for a known host.
	creds, err := provider.FindCredentials(context.Background(), "testhost")
	if err != nil {
		t.Errorf("FindCredentials() unexpected error")
	}
	if *creds != *fakeDrac {
		t.Errorf("FindCredentials() didn't return the expected Credential")
	}

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
	provider := &datastoreProvider{
		client:    mc,
		namespace: "ns",
		projectID: "projectID",
	}

	err := provider.AddCredentials(context.Background(), "testhost", fakeDrac)
	if err != nil {
		t.Errorf("AddCredentials() unexpected error.")
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
	provider := &datastoreProvider{
		client:    mc,
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

	// DeleteCredentials() should fail if the Delete() fails.
	mc.mustFail = true
	err = provider.DeleteCredentials(context.Background(), "testhost")
	mc.mustFail = false
	if err == nil {
		t.Errorf("DeleteCredentials() expected error, got nil.")
	}
}

func Test_datastoreProvider_Close(t *testing.T) {
	// Create a datastoreProvider with a mock connector and client.
	mc := &mockClient{}
	provider := &datastoreProvider{
		client:    mc,
		namespace: "ns",
		projectID: "projectID",
	}

	err := provider.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	mc.mustFail = true
	err = provider.Close()
	if err == nil {
		t.Errorf("Close() expected error, got nil.")
	}
}
