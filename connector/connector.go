package connector

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// ConnType represents the connection's type, i.e. whether we're connecting
// to a OOB management module or to the node's operating system.
type ConnType int

const (
	// DRACConnection is an SSH connection to the node's DRAC
	DRACConnection ConnType = 0
	// HostConnection is an SSH connection to the node's OS
	HostConnection ConnType = 1
)

// ConnectionConfig holds the configuration for a Connection
type ConnectionConfig struct {
	Hostname       string
	Port           int32
	Username       string
	Password       string
	PrivateKeyFile string

	ConnType ConnType
}

// Connector is a provider for Connections.
type Connector interface {
	NewConnection(*ConnectionConfig) (Connection, error)
}

type sshConnector struct {
	dialer dialer
}

func (s *sshConnector) NewConnection(config *ConnectionConfig) (Connection, error) {
	var authMethods []ssh.AuthMethod
	privateBytes, err := ioutil.ReadFile(filepath.Clean(config.PrivateKeyFile))

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

	passwordAuth := ssh.Password(config.Password)
	authMethods = append(authMethods, passwordAuth)

	// TODO: find out how to enable host key verification for M-Lab hosts.
	clientConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	cl, err := s.dialer.Dial("tcp",
		fmt.Sprintf("%s:%d", config.Hostname, config.Port), clientConfig)

	if err != nil {
		return nil, err
	}

	return &sshConnection{
		config: config,
		client: &cl,
	}, nil
}

// NewConnector returns a new Connector based on a default dialer
// implementation.
func NewConnector() Connector {
	return &sshConnector{
		dialer: &sshDialer{},
	}
}

// Connection is any kind of connection over which some commands can be run.
type Connection interface {
	Reboot() (string, error)
	Close() error
}

type sshConnection struct {
	config *ConnectionConfig
	client *client
}

func (c *sshConnection) Exec() (string, error) {
	return "TODO", nil
}

func (c *sshConnection) Reboot() (string, error) {
	return "TODO", nil
}

func (c *sshConnection) Close() error {
	return nil
}
