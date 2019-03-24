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

type mockDialerImpl struct{}

type mockClientImpl struct{}

type mockSessionImpl struct {
	messages map[string]string
}

// Dial is a fake implementation returning an empty ssh.Client
func (*mockDialerImpl) Dial(network, addr string,
	config *ssh.ClientConfig) (client, error) {

	return &mockClientImpl{}, nil
}

// NewSession is a fake implementation returning an empty ssh.Session.
func (*mockClientImpl) NewSession() (session, error) {
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

var (
	mockDialer = &mockDialerImpl{}

	fakeSession = &mockSessionImpl{
		messages: map[string]string{
			"racadm serveraction powercycle": "Server power operation successful",
		},
	}
)

func setupTestConnection(t *testing.T) *Connection {
	conn, err := NewConnection(host, port, username, password, "", mockDialer)
	if err != nil {
		t.Errorf("NewConnection() error = %v", err)
	}

	return conn
}

func TestNewConnection(t *testing.T) {
	t.Run("new-connection-with-password-success", func(t *testing.T) {
		conn := setupTestConnection(t)

		if conn.Host != host || conn.Port != port || conn.dialer != mockDialer {
			t.Errorf("NewConnection() returned an invalid connection")
		}
	})
}

func TestExec(t *testing.T) {
	t.Run("exec-success", func(t *testing.T) {
		conn := setupTestConnection(t)

		out, err := conn.Exec("racadm serveraction powercycle")

		if err != nil {
			t.Errorf("Error while executing command: %v\n", err)
		}

		if out != "Server power operation successful" {
			t.Errorf("Invalid response: %v, expected: Server power operation successful\n", out)
		}

	})
}

func testImplDialer(in dialer) {
	var d DialerImpl
	func(in dialer) {}(&d)
}
