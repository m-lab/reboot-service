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

	dialer Dialer
}

// Dialer is an interface to allow mocking of ssh.Dial in unit tests.
type Dialer interface {
	Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)
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
	NewSession() (*ssh.Session, error)
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

// connect starts an SSH session on Host:Port, using the provided
// credentials.
func (c *Connection) connect() (*ssh.Client, error) {
	client, err := c.dialer.Dial("tcp",
		fmt.Sprintf("%s:%d", c.Host, c.Port), c.Auth)

	if err != nil {
		return nil, err
	}

	return client, err
}

// Exec gets a session and sends a command on this Connection.
func (c *Connection) Exec(cmd string) (string, error) {
	log.Printf("DEBUG: exec %s on %s", cmd, c.Host)
	client, err := c.connect()

	if err != nil {
		return "", err
	}

	session, err := getSession(client)

	if err != nil {
		log.Printf("SSH session creation failed: %s\n", err)
		return "", err
	}
	defer func() {
		if session.Close() != nil {
			log.Printf("Cannot close SSH session: %s\n", err)
		}
	}()

	out, err := session.CombinedOutput(cmd)
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

// getSession gets an SSH session from a client.
func getSession(client Client) (*ssh.Session, error) {
	session, err := client.NewSession()

	if err != nil {
		return nil, err
	}

	return session, nil
}
