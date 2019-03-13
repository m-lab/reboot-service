package storage

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/datastore"
)

type DatastoreClientMock struct {
	Creds      []*Credentials
	mustFail   bool
	skipAppend bool
}

func (d DatastoreClientMock) GetAll(ctx context.Context, q *datastore.Query,
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
	testHost    = "test"
	testUser    = "user"
	testPass    = "pass"
	testModel   = "drac"
	testAddress = "addr"
)

var fakeDrac = &Credentials{
	Hostname: testHost,
	Username: testUser,
	Password: testPass,
	Model:    testModel,
	Address:  testAddress,
}

var mockClient = DatastoreClientMock{
	Creds: []*Credentials{
		fakeDrac,
	},
}

func TestFindCredentials(t *testing.T) {
	ctx := context.Background()

	creds, err := FindCredentials(ctx, mockClient, testHost)
	if err != nil {
		t.Errorf("FindCredentials() error = %v", err)
		return
	}

	if creds.Hostname != fakeDrac.Hostname || creds.Username != fakeDrac.Username || creds.Password != fakeDrac.Password {
		t.Errorf("FindCredentials() didn't return a valid Credentials.")
	}

	mockClient.mustFail = true
	_, err = FindCredentials(ctx, mockClient, testHost)

	if err == nil {
		t.Errorf("FindCredentials() didn't return an error as expected.")
	}

	mockClient.mustFail = false
	mockClient.skipAppend = true

	_, err = FindCredentials(ctx, mockClient, testHost)

	if err == nil {
		t.Errorf("FindCredentials() didn't return an error as expected.")
	}
}
