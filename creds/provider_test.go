package creds

import (
	"context"
	"errors"
	"reflect"
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

func isFakeProvider(t interface{}) bool {
	_, ok := t.(*FakeProvider)
	return ok
}
func TestNewProvider(t *testing.T) {
	provider := NewProvider("projectID", "ns")
	if provider == nil {
		t.Errorf("NewProvider() returned nil.")
	}

	provider = NewProvider("fake", "fake")
	if !isFakeProvider(provider) {
		t.Errorf("NewProvider() didn't return a FakeProvider.")
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

func Test_datastoreProvider_FindCredentials(t *testing.T) {
	type fields struct {
		projectID string
		namespace string
		connector connector
	}
	type args struct {
		ctx  context.Context
		host string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Credentials
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &datastoreProvider{
				projectID: tt.fields.projectID,
				namespace: tt.fields.namespace,
				connector: tt.fields.connector,
			}
			got, err := d.FindCredentials(tt.args.ctx, tt.args.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("datastoreProvider.FindCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("datastoreProvider.FindCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test the FakeProvider implementation.
func TestFakeProvider_AddCredentials(t *testing.T) {
	// Create a FakeProvider and add a Credentials to the map.
	provider := &FakeProvider{
		creds: map[string]*Credentials{},
	}

	fakeDrac := &Credentials{
		Hostname: "host",
		Username: "user",
		Password: "pass",
		Model:    "model",
		Address:  "address",
	}

	provider.AddCredentials("test", fakeDrac)
	if creds, ok := provider.creds["test"]; !ok || creds != fakeDrac {
		t.Errorf("AddCredentials() didn't add the expected Credentials.")
	}
}

func TestFakeProvider_FindCredentials(t *testing.T) {
	fakeDrac := &Credentials{
		Hostname: "host",
		Username: "user",
		Password: "pass",
		Model:    "model",
		Address:  "address",
	}

	provider := &FakeProvider{
		creds: map[string]*Credentials{
			"test": fakeDrac,
		},
	}

	// Retrieve previously added Credentials from the FakeProvider's map.
	creds, err := provider.FindCredentials(context.Background(), "test")
	if err != nil || creds != fakeDrac {
		t.Errorf("FindCredentials() returned an error or wrong Credentials.")
	}

	// Attempt to retrieve Credentials for an unknown hostname.
	creds, err = provider.FindCredentials(context.Background(), "fail")
	if err == nil || creds != nil {
		t.Errorf("FindCredentials() didn't return an error.")
	}
}
