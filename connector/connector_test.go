package connector

import (
	"errors"
	"fmt"
	"testing"

	"golang.org/x/crypto/ssh"
)

type mockDialer struct {
	mustFail bool
}

type mockClient struct{}

type mockSession struct {
	messages map[string]string
}

// Dial is a fake implementation returning an empty ssh.Client
func (d *mockDialer) Dial(network, addr string,
	config *ssh.ClientConfig) (client, error) {

	if d.mustFail {
		return nil, errors.New("method Dial() failed")
	}

	return &mockClient{}, nil
}

// NewSession is a fake implementation returning an empty ssh.Session.
func (*mockClient) NewSession() (session, error) {
	return ms, nil
}

func (*mockClient) Close() error {
	return nil
}

// CombinedOutput returns pre-made responses contained in a map.
func (session *mockSession) CombinedOutput(cmd string) ([]byte, error) {
	if val, ok := session.messages[cmd]; ok {
		return []byte(val), nil
	}

	return nil, fmt.Errorf("Undefined message for command: %v", cmd)
}

func (session *mockSession) Close() error {
	return nil
}

var (
	md = &mockDialer{}

	ms = &mockSession{
		messages: map[string]string{
			"racadm serveraction powercycle": "Server power operation successful",
		},
	}
)

func Test_sshConnector_NewConnection(t *testing.T) {
	connector := &sshConnector{
		dialer: md,
	}

	config := &ConnectionConfig{
		Hostname:       "testhost",
		Port:           22,
		Username:       "testuser",
		Password:       "testpass",
		PrivateKeyFile: "",
		ConnType:       DRACConnection,
	}

	_, err := connector.NewConnection(config)
	if err != nil {
		t.Errorf("NewConnection() - unexpected error: %v", err)
	}

	// If dialer.Dial fails, NewConnection should fail.
	md.mustFail = true
	_, err = connector.NewConnection(config)
	if err == nil {
		t.Errorf("NewConnection() - expected err, got nil.")
	}

}

func TestNewConnector(t *testing.T) {
	// Just test that a default Connector is created
	_ = NewConnector()
}
