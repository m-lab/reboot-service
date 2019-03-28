package creds

import (
	"context"
	"errors"

	"cloud.google.com/go/datastore"
	"github.com/apex/log"
)

// Credentials is a struct holding the credentials for a given hostname,
// plus some additional metadata such as the IP address and the model (DRAC
// or otherwise).
type Credentials struct {
	Hostname string `datastore:"hostname"`
	Username string `datastore:"username"`
	Password string `datastore:"password"`
	Model    string `datastore:"model"`
	Address  string `datastore:"address"`
}

// Provider is a Credentials provider.
type Provider interface {
	FindCredentials(context.Context, string) (*Credentials, error)
}

// datastoreProvider is a Provider based on Google Cloud Datastore.
type datastoreProvider struct {
	projectID string
	namespace string
	kind      string

	connector connector
}

// NewProvider returns a Provider based on the default implementation (GCD).
func NewProvider(projectID, namespace, kind string) (Provider, error) {
	return &datastoreProvider{
		kind:      kind,
		projectID: projectID,
		namespace: namespace,

		connector: &datastoreConnector{},
	}, nil
}

func (d *datastoreProvider) FindCredentials(ctx context.Context, host string) (*Credentials, error) {
	client, err := d.connector.NewClient(ctx, d.projectID)
	if err != nil {
		log.WithError(err).Errorf("Error while creating datastore client")
		return nil, err
	}

	log.Debugf("Retrieving credentials for %v from namespace %v", host, d.namespace)

	query := datastore.NewQuery(d.kind).Namespace(d.namespace)
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
