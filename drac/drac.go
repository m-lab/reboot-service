// Package drac contains the functions to establish an SSH connection with
// a DRAC, using a username/password pair, and implements some utility
// functions to:
// - Reboot the node
// - Disable/Enable IP block for remote access
package drac

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// Connection represents a connection to a DRAC. It includes hostname, port,
// credentials and it wraps a ssh.ClientConfig containing authentication
// settings.
type Connection struct {
	Host string
	Port int32
	Auth *ssh.ClientConfig

	dialer  Dialer
	client  Client
	session Session
}

// Dialer is an interface to allow mocking of ssh.Dial in unit tests.
type Dialer interface {
	Dial(network, addr string, config *ssh.ClientConfig) (Client, error)
}

// DialerImpl is a default implementation of Dialer.
type DialerImpl struct{}

// Dial is just a wrapper around ssh.Dial
func (d *DialerImpl) Dial(network, addr string,
	config *ssh.ClientConfig) (*ssh.Client, error) {
	return ssh.Dial(network, addr, config)
}

// Client is an interface to allow mocking of ssh.Client in unit tests.
type Client interface {
	NewSession() (Session, error)
	Close() error
}

type Session interface {
	CombinedOutput(cmd string) ([]byte, error)
	Close() error
}

type SessionWrapper struct {
	Session
}

// NewConnection returns a new Connection configured with the specified
// credentials.
func NewConnection(host string, port int32, username string, password string,
	privateKeyPath string, dialer Dialer) (*Connection, error) {

	var authMethods []ssh.AuthMethod
	privateBytes, err := ioutil.ReadFile(filepath.Clean(privateKeyPath))

	if err != nil {
		log.Println("Cannot read private key: ", err)
	} else {

		privateKey, err := ssh.ParsePrivateKey(privateBytes)

		if err != nil {
			// If a private key exists but it's not parseable, the connection
			// is not created.
			log.Println("Cannot parse private key: ", err)
			return nil, err
		}

		privateKeyAuth := ssh.PublicKeys(privateKey)
		authMethods = append(authMethods, privateKeyAuth)
	}

	passwordAuth := ssh.Password(password)
	authMethods = append(authMethods, passwordAuth)

	// TODO: find out how to enable host key verification for M-Lab hosts.
	clientConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn := &Connection{
		Host:   host,
		Port:   port,
		Auth:   clientConfig,
		dialer: dialer,
	}

	return conn, nil
}

// Connect initializes the SSH client connecting to Host:Port.
// This method is idempotent and won't create a Client when
// one is available already. To create a new Client, call Close()
// first.
func (c *Connection) connect() error {
	if c.client == nil {
		client, err := c.dialer.Dial("tcp",
			fmt.Sprintf("%s:%d", c.Host, c.Port), c.Auth)

		if err != nil {
			return err
		}

		c.client = client
	}

	return nil
}

// CreateSession creates a new session for the current Client. This method is
// idempotent and won't create a Session when one is available already. To
// create a new Session, call Close() first, then Connect() again.
func (c *Connection) createSession() error {
	if c.session == nil {
		session, err := c.client.NewSession()
		if err != nil {
			log.Printf("Error while initializing SSH session: %s", err)
		}

		c.session = session
	}

	return nil
}

// Close closes the underlying Client and Session and sets the corresponding
// pointers to nil.
func (c *Connection) close() error {
	err := c.session.Close()
	if err != nil {
		return err
	}

	c.session = nil

	err = c.client.Close()
	if err != nil {
		return err
	}

	c.client = nil
	return nil

}

// Exec executes a command on this Connection.
func (c *Connection) Exec(cmd string) (string, error) {
	c.connect()
	c.createSession()
	defer c.close()

	out, err := c.session.CombinedOutput(cmd)
	if err != nil {
		log.Printf("Command execution failed: %s\n", err)
		return "", err
	}

	return string(out), nil
}

// Reboot sends a reboot command on this Connection.
func (c *Connection) Reboot() (string, error) {
	log.Printf("DEBUG: reboot %s", c.Host)
	return c.Exec("racadm serveraction powercycle")
}
