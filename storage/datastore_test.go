package storage

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/datastore"
)

// mockDatastoreClient is a fake DatastoreClient for testing.
type mockDatastoreClient struct {
	Creds      []*Credentials
	mustFail   bool
	skipAppend bool
}

func (d mockDatastoreClient) GetAll(ctx context.Context, q *datastore.Query,
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
	Hostname: testHost,
	Username: testUser,
	Password: testPass,
	Model:    testModel,
	Address:  testAddress,
}

func TestFindCredentials(t *testing.T) {
	ctx := context.Background()
	var mockClient = &mockDatastoreClient{
		Creds: []*Credentials{
			fakeDrac,
		},
	}

	var datastore = Datastore{
		Client:    mockClient,
		Namespace: testNamespace,
	}

	// FindCredentials must return valid credentials for a known hostname.
	creds, err := FindCredentials(ctx, datastore, testHost)
	if err != nil {
		t.Errorf("FindCredentials() error = %v", err)
		return
	}
	if *creds != *fakeDrac {
		t.Errorf("FindCredentials() didn't return a valid Credentials.")
	}

	// FindCredentials must fail if there is an error while retrieving
	// credentials from Datastore.
	mockClient.mustFail = true
	_, err = FindCredentials(ctx, datastore, testHost)
	if err == nil {
		t.Errorf("FindCredentials() didn't return an error as expected.")
	}

	// FindCredentials must return an error if there is no result for the
	// requested hostname.
	mockClient.mustFail = false
	mockClient.skipAppend = true
	_, err = FindCredentials(ctx, datastore, testHost)
	if err == nil {
		t.Errorf("FindCredentials() didn't return an error as expected.")
	}
}
