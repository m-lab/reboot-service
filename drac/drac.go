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

	dialer  dialer
	client  client
	session session
}

// dialer is an interface to allow mocking of ssh.Dial in unit tests.
type dialer interface {
	Dial(network, addr string, config *ssh.ClientConfig) (client, error)
}

// client is an interface to allow mocking of ssh.client in unit tests.
type client interface {
	NewSession() (session, error)
	Close() error
}

type session interface {
	CombinedOutput(cmd string) ([]byte, error)
	Close() error
}

type dialerImpl struct{}

type clientImpl struct {
	*ssh.Client
}

func (cw clientImpl) NewSession() (session, error) { return cw.NewSession() }
func (cw clientImpl) Close() error                 { return cw.Close() }

// Dial is just a wrapper around ssh.Dial
func (d *dialerImpl) Dial(network, addr string,
	config *ssh.ClientConfig) (client, error) {

	cl, err := ssh.Dial(network, addr, config)

	if err != nil {
		return nil, err
	}

	return clientImpl{cl}, nil
}

// NewConnection returns a new Connection configured with the specified
// credentials.
func NewConnection(host string, port int32, username string, password string,
	privateKeyPath string, d dialer) (*Connection, error) {

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
		dialer: d,
	}

	return conn, nil
}

// connect initializes the SSH client connecting to Host:Port.
// This method is idempotent and won't create a Client when
// one is available already. To create a new Client, call Close()
// first.
func (c *Connection) connect() error {
	if c.client == nil {
		cl, err := c.dialer.Dial("tcp",
			fmt.Sprintf("%s:%d", c.Host, c.Port), c.Auth)

		if err != nil {
			return err
		}

		c.client = cl
	}

	return nil
}

// createSession creates a new session for the current Client. This method is
// idempotent and won't create a Session when one is available already. To
// create a new Session, call Close() first, then Connect() again.
func (c *Connection) createSession() error {
	if c.session == nil {
		sess, err := c.client.NewSession()
		if err != nil {
			return err
		}

		c.session = sess
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
	err := c.connect()
	if err != nil {
		log.Printf("Error while initializing SSH session: %s", err)
		return "", err
	}

	err = c.createSession()
	if err != nil {
		log.Printf("Error while initializing session: %s", err)
		return "", err
	}
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
