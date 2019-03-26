package reboot

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"cloud.google.com/go/datastore"

	"github.com/m-lab/reboot-service/drac"
	"github.com/m-lab/reboot-service/storage"
	"github.com/m-lab/reboot-service/storage/iface"

	"google.golang.org/api/option"

	"golang.org/x/crypto/ssh"
)

// Mock SSH client
type mockDialerImpl struct{}

type mockClientImpl struct{}

type mockSessionImpl struct {
	messages map[string]string
}

// Dial is a fake implementation returning an empty ssh.Client
func (*mockDialerImpl) Dial(network, addr string,
	config *ssh.ClientConfig) (drac.Client, error) {

	return &mockClientImpl{}, nil
}

// NewSession is a fake implementation returning an empty ssh.Session.
func (*mockClientImpl) NewSession() (drac.Session, error) {
	return fakeSession, nil
}

func (*mockClientImpl) Close() error {
	return nil
}

// CombinedOutput returns pre-made responses contained in a map.
func (session *mockSessionImpl) CombinedOutput(cmd string) ([]byte, error) {
	if val, ok := session.messages[cmd]; ok {
		return []byte(val), nil
	}

	return nil, fmt.Errorf("Undefined message for command: %v", cmd)
}

func (session *mockSessionImpl) Close() error {
	return nil
}

// Mock storage

// mockDatastoreClient is a fake DatastoreClient for testing.
type mockDatastoreClient struct {
	Creds      []*storage.Credentials
	mustFail   bool
	skipAppend bool
}

func (d mockDatastoreClient) GetAll(ctx context.Context, q *datastore.Query,
	dst interface{}) ([]*datastore.Key, error) {

	if d.mustFail {
		return nil, errors.New("method GetAll failed")
	}

	if !d.skipAppend {
		creds := dst.(*[]*storage.Credentials)
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

var fakeDrac = &storage.Credentials{
	Hostname: testHost,
	Username: testUser,
	Password: testPass,
	Model:    testModel,
	Address:  testAddress,
}

var (
	mockDialer = &mockDialerImpl{}

	fakeSession = &mockSessionImpl{
		messages: map[string]string{
			"racadm serveraction powercycle": "Server power operation successful",
		},
	}
)

func TestDRAC(t *testing.T) {
	conf := &Config{
		DRACPort:  806,
		Namespace: "test",
		ProjectID: "testproj",
		SSHPort:   22,
		Dialer:    mockDialer,
	}

	dsClient := &mockDatastoreClient{
		Creds: []*storage.Credentials{
			fakeDrac,
		},
	}

	failingDSClient := &mockDatastoreClient{
		mustFail: true,
	}

	// Inject a mock DatastoreClient that knows just one host.
	dsNewClient = func(ctx context.Context, projectID string, opts ...option.ClientOption) (iface.DatastoreClient, error) {
		return dsClient, nil
	}

	out, err := DRAC(context.Background(), conf, "test")
	if err != nil {
		t.Errorf("DRAC() expected err = nil, got %v", err)
	}
	if out != "Server power operation successful" {
		t.Errorf("DRAC() returned an unexpected output: %v", out)
	}

	// If credentials' retrieval from datastore fails, DRAC should fail too.
	// Inject a mock DatastoreClient that fails no matter what.
	dsNewClient = func(ctx context.Context, projectID string, opts ...option.ClientOption) (iface.DatastoreClient, error) {
		return failingDSClient, nil
	}

	_, err = DRAC(context.Background(), conf, "test")
	if err == nil {
		t.Errorf("DRAC() expected err, got nil")
	}

}
