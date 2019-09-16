package connector

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// ConnType represents the connection's type, i.e. whether we're connecting
// to a OOB management module or to the node's operating system.
type ConnType int

const (
	// UnspecifiedConnection is a connection with no type defined.
	// This value should not be used.
	UnspecifiedConnection ConnType = 0
	// BMCConnection is an SSH connection to the node's BMC
	BMCConnection ConnType = 1
	// HostConnection is an SSH connection to the node's OS
	HostConnection ConnType = 2
)

// ConnectionConfig holds the configuration for a Connection
type ConnectionConfig struct {
	Hostname       string
	Port           int32
	Username       string
	Password       string
	PrivateKeyFile string

	ConnType ConnType
	Timeout  time.Duration
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
		// If a private key cannot be read, we still want to try logging in
		// with a username/password pair, thus this is just a warning.
		// This also allows to skip private key auth by passing an empty
		// string.
		log.Println("Cannot read private key: ", err)
	} else {

		privateKey, err := ssh.ParsePrivateKey(privateBytes)

		if err != nil {
			// If a private key file is provided and can be read but the
			// content is a not a valid private key,
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
		Timeout:         config.Timeout,
	}

	cl, err := s.dialer.Dial("tcp",
		fmt.Sprintf("%s:%d", config.Hostname, config.Port), clientConfig)

	if err != nil {
		return nil, err
	}

	return &sshConnection{
		config: config,
		client: cl,
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
	ExecDRACShell(string) (string, error)
	Reboot() (string, error)
	Close() error
}

type sshConnection struct {
	config *ConnectionConfig
	client client
}

// exec runs a command over the connection. It's meant to be used internally
// inside wrappers such as Reboot().
func (c *sshConnection) exec(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		log.Printf("Error executing command \"%v\": %v", cmd, err)
	}

	return string(output), err
}

// ExecDRACShell runs a command in shell mode on a DRAC's SSH server.
//
// Due to $reasons (a bug?) the server does not seem to send exit codes
// correctly after a command's execution - thus, session.Wait() hangs
// indefinitely. However, the exit code is sent after the session is closed,
// which can be forced 1. with stdin.Close() 2. by writing "exit" on stdin.
func (c *sshConnection) ExecDRACShell(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return "", err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", err
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return "", err
	}

	err = session.Shell()
	if err != nil {
		return "", err
	}

	_, err = fmt.Fprintf(stdin, "%s\n", cmd)
	if err != nil {
		return "", err
	}

	_, err = fmt.Fprintf(stdin, "exit\n")
	if err != nil {
		return "", err
	}

	log.Println("Waiting...")
	session.Wait()

	readers := io.MultiReader(stdout, stderr)
	out, err := ioutil.ReadAll(readers)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// Reboot reboots the node via this Connection. The method to perform the
// reboot is chosen depending on sshConnection.ConnType.
func (c *sshConnection) Reboot() (string, error) {
	var output string
	var err error

	if c.config.ConnType == HostConnection {
		// To reboot a host a "reboot-api" user is created, and the only way
		// to authenticate is via a private key. Logging in with such user will
		// automatically trigger a "systemctl reboot" command.
		// To actually start an SSH session (and thus trigger a reboot) a
		// command must be executed.
		output, err = c.exec("")
		if err != nil {
			return "", err
		}
	} else if c.config.ConnType == BMCConnection {
		output, err = c.exec("racadm serveraction powercycle")
		if err != nil {
			return "", err
		}
	} else {
		return "", errors.New("unable to reboot: unspecified connection")
	}
	return output, nil
}

func (c *sshConnection) Close() error {
	err := c.client.Close()
	if err != nil {
		log.Printf("Error while closing connection: %v", err)
	}

	return err
}
