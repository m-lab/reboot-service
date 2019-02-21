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

	"golang.org/x/crypto/ssh"
)

// DRACConnection represents a connection to a DRAC. It includes hostname, port,
// credentials and it wraps a ssh.ClientConfig containing authentication
// settings.
type DRACConnection struct {
	Host string
	Port int32
	Auth *ssh.ClientConfig

	dialer Dialer
}

// Interface to allow mocking of ssh.Dial in unit tests
type Dialer interface {
	Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)
}

type DialerImpl struct{}

// Dial is just a wrapper around ssh.Dial
func (d *DialerImpl) Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	return ssh.Dial(network, addr, config)
}

type Client interface {
	NewSession() (*ssh.Session, error)
}

// NewConnection returns a new Connection configured with the specified
// credentials.
func NewConnection(host string, port int32, username string, password string, privateKeyPath string,
	dialer Dialer) (*DRACConnection, error) {

	var authMethods []ssh.AuthMethod
	privateBytes, err := ioutil.ReadFile(privateKeyPath)

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

	conn := &DRACConnection{
		Host:   host,
		Port:   port,
		Auth:   clientConfig,
		dialer: dialer,
	}

	return conn, nil
}

// connect starts an SSH session on Host:Port, using the provided
// credentials.
func (c *DRACConnection) connect() (*ssh.Client, error) {
	client, err := c.dialer.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), c.Auth)

	if err != nil {
		return nil, err
	}

	return client, err
}

func (c *DRACConnection) getSession(client Client) (*ssh.Session, error) {
	session, err := client.NewSession()

	if err != nil {
		return nil, err
	}

	return session, nil
}

// Exec gets a session and sends a command on this Connection.
func (c *DRACConnection) Exec(cmd string) (string, error) {
	log.Printf("DEBUG: exec %s on %s", cmd, c.Host)
	client, err := c.connect()

	if err != nil {
		return "", err
	}

	session, err := c.getSession(client)

	if err != nil {
		log.Fatalf("Command execution failed: %s\n", err)

	}
	defer session.Close()

	out, _ := session.CombinedOutput(cmd)

	return string(out), nil
}

// Reboot sends a reboot command on this Connection.
func (c *DRACConnection) Reboot() (string, error) {
	log.Printf("DEBUG: reboot %s", c.Host)
	return c.Exec("racadm serveraction powercycle")
}
