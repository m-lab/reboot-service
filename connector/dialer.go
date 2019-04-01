package connector

import "golang.org/x/crypto/ssh"

// dialer is an interface to allow mocking of ssh.Dial in unit tests.
type dialer interface {
	Dial(network, addr string, config *ssh.ClientConfig) (client, error)
}

// client is an interface to allow mocking of ssh.Client in unit tests.
type client interface {
	NewSession() (session, error)
	Close() error
}

type session interface {
	CombinedOutput(cmd string) ([]byte, error)
	Close() error
}

type sshDialer struct{}

type sshClient struct {
	client *ssh.Client
}

func (cw sshClient) NewSession() (session, error) { return cw.client.NewSession() }
func (cw sshClient) Close() error                 { return cw.client.Close() }

// Dial is just a wrapper around ssh.Dial
func (d *sshDialer) Dial(network, addr string,
	config *ssh.ClientConfig) (client, error) {

	cl, err := ssh.Dial(network, addr, config)

	if err != nil {
		return nil, err
	}

	return sshClient{cl}, nil
}
