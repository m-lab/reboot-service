package connector

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"golang.org/x/crypto/ssh"
)

type mockDialer struct {
	mustFail bool
}

type mockClient struct {
	mustFail bool
}

type mockSession struct {
	messages map[string]string
}

// Dial is a fake implementation returning an empty ssh.Client
func (d *mockDialer) Dial(network, addr string,
	config *ssh.ClientConfig) (client, error) {

	if d.mustFail {
		return nil, errors.New("method Dial() failed")
	}

	return mc, nil
}

// NewSession is a fake implementation returning an empty ssh.Session.
func (c *mockClient) NewSession() (session, error) {
	if c.mustFail {
		return nil, errors.New("method NewSession() failed")
	}
	return ms, nil
}

func (c *mockClient) Close() error {
	if c.mustFail {
		return errors.New("method Close() failed")
	}
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

func (session *mockSession) Shell() error {
	return nil
}

func (session *mockSession) StdinPipe() (io.WriteCloser, error) {
	return nil, nil
}

func (session *mockSession) StdoutPipe() (io.Reader, error) {
	return nil, nil
}

func (session *mockSession) StderrPipe() (io.Reader, error) {
	return nil, nil
}

func (session *mockSession) Wait() error {
	return nil
}

var (
	md = &mockDialer{}
	mc = &mockClient{}
	ms = &mockSession{
		messages: map[string]string{
			"racadm serveraction powercycle": "Server power operation successful",

			// empty command -> empty response allows to test HostConnection.
			"": "",
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
		ConnType:       BMCConnection,
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
	md.mustFail = false
}

func TestNewConnector(t *testing.T) {
	// Just test that a default Connector is created
	if NewConnector() == nil {
		t.Errorf("NewConnector() returned nil.")
	}
}

func Test_sshConnection_Reboot(t *testing.T) {
	connector := &sshConnector{
		dialer: md,
	}

	config := &ConnectionConfig{
		Hostname:       "testhost",
		Port:           22,
		Username:       "testuser",
		Password:       "testpass",
		PrivateKeyFile: "",
		ConnType:       BMCConnection,
	}

	bmcConn, err := connector.NewConnection(config)
	if err != nil {
		t.Errorf("NewConnection() - unexpected error: %v", err)
	}

	hostConfig := &ConnectionConfig{
		Hostname:       "testhost",
		Port:           22,
		Username:       "testuser",
		Password:       "testpass",
		PrivateKeyFile: "",
		ConnType:       HostConnection,
	}

	hostConn, err := connector.NewConnection(hostConfig)
	if err != nil {
		t.Errorf("NewConnection() - unexpected error: %v", err)
	}

	// Reboot() should return a known output.
	output, err := bmcConn.Reboot()
	if err != nil {
		t.Errorf("Reboot() unexpected error: %v", err)
	}
	if output != "Server power operation successful" {
		t.Errorf("Reboot() returned an unexpected output: %v", output)
	}

	// Reboot() should fail if the command execution fails.
	messages := ms.messages
	ms.messages = make(map[string]string)
	output, err = bmcConn.Reboot()
	if err == nil {
		t.Errorf("Reboot() expected error, got nil.")
	}
	ms.messages = messages

	// Reboot() should fail if a session can't be created.
	mc.mustFail = true
	output, err = bmcConn.Reboot()
	if err == nil {
		t.Errorf("Reboot() expected error, got nil.")
	}
	mc.mustFail = false

	// Tests for HostConnection
	// Reboot() should return a known output.
	output, err = hostConn.Reboot()
	if err != nil {
		t.Errorf("Reboot() unexpected error: %v", err)
	}

	// Reboot() should fail if the command execution fails.
	messages = ms.messages
	ms.messages = make(map[string]string)
	output, err = hostConn.Reboot()
	if err == nil {
		t.Errorf("Reboot() expected error, got nil.")
	}
	ms.messages = messages

	// Reboot() should fail if a session can't be created.
	mc.mustFail = true
	output, err = hostConn.Reboot()
	if err == nil {
		t.Errorf("Reboot() expected error, got nil.")
	}
	mc.mustFail = false
}

func Test_sshConnection_Close(t *testing.T) {
	connector := &sshConnector{
		dialer: md,
	}

	config := &ConnectionConfig{
		Hostname:       "testhost",
		Port:           22,
		Username:       "testuser",
		Password:       "testpass",
		PrivateKeyFile: "",
		ConnType:       BMCConnection,
	}

	conn, err := connector.NewConnection(config)
	if err != nil {
		t.Errorf("NewConnection() - unexpected error: %v", err)
	}

	// Close() shouldn't return an error if it succeeds.
	err = conn.Close()
	if err != nil {
		t.Errorf("Close() error: %v", err)
	}

	// Close() should fail if the underlying client.Close() fails.
	mc.mustFail = true
	err = conn.Close()
	if err == nil {
		t.Errorf("Close() unexpected error: %v", err)
	}
}
