package storage

import (
	"context"
	"testing"

	"cloud.google.com/go/datastore"
)

type DatastoreClientMock struct {
	Creds []*Credentials
}

func (DatastoreClientMock) GetAll(ctx context.Context, q *datastore.Query,
	dst interface{}) ([]*datastore.Key, error) {

	creds := dst.(*[]*Credentials)
	*creds = append(*creds, fakeDrac)
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
	creds, err := FindCredentials(context.Background(), mockClient, testHost)
	if err != nil {
		t.Errorf("FindCredentials() error = %v", err)
		return
	}

	if creds.Hostname != fakeDrac.Hostname || creds.Username != fakeDrac.Username || creds.Password != fakeDrac.Password {
		t.Errorf("FindCredentials() didn't return a valid Credentials")
	}
}
