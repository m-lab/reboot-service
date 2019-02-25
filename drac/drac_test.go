package drac

import (
	"fmt"
	"testing"

	"golang.org/x/crypto/ssh"
)

const (
	host     = "localhost"
	port     = 22
	username = "test"
	password = "test"
)

type mockDialer struct{}
type mockClient struct{}
type mockSession struct {
	messages map[string]string
}

// Dial is a fake implementation returning an empty ssh.Client
func (*mockDialer) Dial(network, addr string,
	config *ssh.ClientConfig) (Client, error) {

	return &mockClient{}, nil
}

// NewSession is a fake implementation returning an empty ssh.Session.
func (*mockClient) NewSession() (Session, error) {
	return fakeSession, nil
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
	dialer = &mockDialer{}
	client = &mockClient{}

	fakeSession = &mockSession{
		messages: map[string]string{
			"racadm serveraction powercycle": "Server power operation successful",
		},
	}
)

func TestNewConnection(t *testing.T) {
	t.Run("new-connection-with-password-success", func(t *testing.T) {
		conn, err := NewConnection(host, port, username, password, "", dialer)
		if err != nil {
			t.Errorf("NewConnection() error = %v", err)
			return
		}

		if conn.Host != host || conn.Port != port || conn.dialer != dialer {
			t.Errorf("NewConnection() returned an invalid connection")
		}
	})
}

func TestConnection_Exec(t *testing.T) {
	t.Run("exec-success", func(t *testing.T) {
		c, err := NewConnection(host, port, username, password, "", dialer)
		if err != nil {
			t.Errorf("Error while creating connection: %v", err)
		}

		out, err := c.Exec("racadm serveraction powercycle")

		if err != nil {
			t.Errorf("Error while executing command: %v\n", err)
		}

		if out != "Server power operation successful" {
			t.Errorf("Invalid response: %v, expected: Server power operation successful\n", out)
		}

	})
}
