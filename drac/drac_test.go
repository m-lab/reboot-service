// Package drac contains the functions to establish an SSH connection with
// a DRAC, using a username/password pair, and implements some utility
// functions to:
// - Reboot the node
// - Disable/Enable IP block for remote access
package drac

import (
	"os"
	"reflect"
	"testing"

	"golang.org/x/crypto/ssh"
)

const (
	host           = "localhost"
	port           = 22
	username       = "test"
	password       = "test"
	privateKeyPath = "private.key"
)

type MockDialer struct{}
type MockClient struct{}

// Dial is a fake implementation returning an empty ssh.Client
func (*MockDialer) Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	return fakeSSHClient, nil
}

// NewSession is a fake implementation returning an empty ssh.Session
func (*MockClient) NewSession() (*ssh.Session, error) {
	return fakeSession, nil
}

var (
	mockDialer    = &MockDialer{}
	mockClient    = &MockClient{}
	fakeSSHClient = &ssh.Client{}
	fakeSession   = &ssh.Session{}
)

func TestNewConnection(t *testing.T) {
	_, err := os.Create(privateKeyPath)
	if err != nil {
		t.Errorf("Cannot create the fake private key file: %v", err)
		return
	}
	defer os.Remove(privateKeyPath)

	t.Run("new-connection-with-password-success", func(t *testing.T) {
		conn, err := NewConnection(host, port, username, password, "", mockDialer)
		if err != nil {
			t.Errorf("NewConnection() error = %v", err)
			return
		}

		if conn.Host != host || conn.Port != port || conn.dialer != mockDialer {
			t.Errorf("NewConnection() returned an invalid connection")
		}
	})
}

func TestDRACConnection_connect(t *testing.T) {
	t.Run("connect-success", func(t *testing.T) {
		c, err := NewConnection(host, port, username, password, "", mockDialer)
		if err != nil {
			t.Errorf("Error while creating connection: %v", err)
		}

		client, err := c.connect()
		if err != nil {
			t.Errorf("DRACConnection.connect() error = %v", err)
			return
		}

		if !reflect.DeepEqual(client, fakeSSHClient) {
			t.Errorf("DRACConnection.connect() = %v, want %v", client, fakeSSHClient)
		}
	})
}

func TestDRACConnection_getSession(t *testing.T) {
	t.Run("get-session-success", func(t *testing.T) {
		c, err := NewConnection(host, port, username, password, "", mockDialer)
		if err != nil {
			t.Errorf("Error while creating connection: %v", err)
		}

		got, err := c.getSession(mockClient)
		if err != nil {
			t.Errorf("DRACConnection.getSession() error = %v", err)
			return
		}

		if !reflect.DeepEqual(got, fakeSession) {
			t.Errorf("DRACConnection.getSession() = %v, want %v", got, fakeSession)
		}
	})

}