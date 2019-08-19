package credstest

import (
	"context"
	"errors"

	"github.com/m-lab/reboot-service/creds"
)

// FakeProvider is a fake provider to use for testing. It holds a map of
// hostname -> *Credentials that can be populated as needed when testing.
type FakeProvider struct {
	creds map[string]*creds.Credentials
}

// NewProvider returns a FakeProvider.
func NewProvider() *FakeProvider {
	return &FakeProvider{
		creds: make(map[string]*creds.Credentials),
	}
}

// FindCredentials returns a Credentials from the creds map or an error.
func (p *FakeProvider) FindCredentials(ctx context.Context,
	host string) (*creds.Credentials, error) {
	if cred, ok := p.creds[host]; ok {
		return cred, nil
	}

	return nil, errors.New("hostname not found")
}

// AddCredentials adds a Credentials to the map.
func (p *FakeProvider) AddCredentials(ctx context.Context, host string,
	cred *creds.Credentials) error {
	p.creds[host] = cred
	return nil
}

// DeleteCredentials removes a Credentials from the map.
// Note: GCD does not return an error if the key does not exist. However, to
// make this easier to test, FakeProvider.DeleteCredentials will fail in that
// case.
func (p *FakeProvider) DeleteCredentials(ctx context.Context, host string) error {
	if _, ok := p.creds[host]; ok {
		delete(p.creds, host)
		return nil
	}
	return errors.New("hostname not found")
}
