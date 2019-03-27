package connector

import "golang.org/x/crypto/ssh"

// Dialer is an interface to allow mocking of ssh.Dial in unit tests.
type dialer interface {
	Dial(network, addr string, config *ssh.ClientConfig) (client, error)
}

// Client is an interface to allow mocking of ssh.Client in unit tests.
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
	client *ssh.Client
}

func (cw clientImpl) NewSession() (session, error) { return cw.client.NewSession() }
func (cw clientImpl) Close() error                 { return cw.client.Close() }

// Dial is just a wrapper around ssh.Dial
func (d *dialerImpl) Dial(network, addr string,
	config *ssh.ClientConfig) (client, error) {

	cl, err := ssh.Dial(network, addr, config)

	if err != nil {
		return nil, err
	}

	return clientImpl{cl}, nil
}
