package creds

import (
	"context"
	"errors"

	"cloud.google.com/go/datastore"
	"github.com/apex/log"
)

const kind = "Credentials"

// Credentials is a struct holding the credentials for a given hostname,
// plus some additional metadata such as the IP address and the model (DRAC
// or otherwise).
type Credentials struct {
	Hostname string `datastore:"hostname" json:"hostname"`
	Username string `datastore:"username" json:"username"`
	Password string `datastore:"password" json:"password"`
	Model    string `datastore:"model" json:"model"`
	Address  string `datastore:"address" json:"address"`
}

// Provider is a Credentials provider.
type Provider interface {
	FindCredentials(context.Context, string) (*Credentials, error)

	// AddCredentials creates a new Credentials entity on this Provider.
	AddCredentials(context.Context, string, *Credentials) error
}

// datastoreProvider is a Provider based on Google Cloud Datastore.
type datastoreProvider struct {
	projectID string
	namespace string

	connector connector
}

// FakeProvider is a fake provider to use for testing. It holds a map of
// hostname -> *Credentials that can be populated as needed when testing.
type FakeProvider struct {
	creds map[string]*Credentials
}

// FindCredentials returns a Credentials from the creds map or an error.
func (p *FakeProvider) FindCredentials(ctx context.Context,
	host string) (*Credentials, error) {
	if cred, ok := p.creds[host]; ok {
		return cred, nil
	}

	return nil, errors.New("hostname not found")
}

// AddCredentials adds a Credentials to the map.
func (p *FakeProvider) AddCredentials(ctx context.Context, host string,
	cred *Credentials) error {
	p.creds[host] = cred
	return nil
}

// NewProvider returns a Provider based on the default implementation (GCD).
// If the projectID and namespace are both "fake", it returns a fake
// implementation that's useful for testing.
func NewProvider(projectID, namespace string) Provider {
	if projectID == "fake" && namespace == "fake" {
		return &FakeProvider{}
	}

	return &datastoreProvider{
		projectID: projectID,
		namespace: namespace,

		connector: &datastoreConnector{},
	}
}

func (d *datastoreProvider) FindCredentials(ctx context.Context, host string) (*Credentials, error) {
	client, err := d.connector.NewClient(ctx, d.projectID)
	if err != nil {
		log.WithError(err).Errorf("Error while creating datastore client")
		return nil, err
	}

	log.Debugf("Retrieving credentials for %v from namespace %v", host, d.namespace)

	query := datastore.NewQuery(kind).Namespace(d.namespace)
	query = query.Filter("hostname = ", host)

	var creds []*Credentials
	_, err = client.GetAll(ctx, query, &creds)

	if err != nil {
		return nil, err
	}

	if len(creds) == 0 {
		return nil, errors.New("Hostname not found in Datastore")
	}

	cred := creds[0]
	return cred, nil
}

// AddCredentials creates a new Credentials entity on GCD.
func (d *datastoreProvider) AddCredentials(ctx context.Context,
	host string, creds *Credentials) error {
	client, err := d.connector.NewClient(ctx, d.projectID)
	if err != nil {
		log.WithError(err).Errorf("Error while creating datastore client")
		return err
	}

	log.Debugf("Adding credentials for %v to namespace %v", host, d.namespace)

	// Create entity with key=hostname
	key := datastore.NameKey(kind, host, nil)
	key.Namespace = d.namespace
	_, err = client.Put(ctx, key, creds)
	if err != nil {
		log.WithError(err).Errorf("Cannot add Credentials entity")
	}
	return nil
}
