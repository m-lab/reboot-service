package creds

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"cloud.google.com/go/datastore"
	"github.com/apex/log"
	"github.com/m-lab/go/rtx"
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

// String marshals a Credentials to a JSON string, disabling HTML escaping
// so that special characters are shown correctly and adding indentation.
func (c *Credentials) String() string {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	rtx.Must(enc.Encode(c), "Error while marshalling JSON")
	return buf.String()
}

// Provider is a Credentials provider.
type Provider interface {
	FindCredentials(context.Context, string) (*Credentials, error)

	// AddCredentials creates a new Credentials entity on this Provider.
	AddCredentials(context.Context, string, *Credentials) error

	// DeleteCredentials removes existing Credentials entities from this
	// provider.
	DeleteCredentials(context.Context, string) error
}

// datastoreProvider is a Provider based on Google Cloud Datastore.
type datastoreProvider struct {
	projectID string
	namespace string

	connector connector
}

// NewProvider returns a Provider based on the default implementation (GCD).
func NewProvider(projectID, namespace string) Provider {
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
	defer client.Close()

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
	defer client.Close()

	log.Debugf("Adding credentials for %v to namespace %v", host, d.namespace)

	// Create entity with key=hostname
	key := datastore.NameKey(kind, host, nil)
	key.Namespace = d.namespace
	_, err = client.Put(ctx, key, creds)
	if err != nil {
		log.WithError(err).Errorf("Cannot add Credentials entity")
		return err
	}
	return nil
}

func (d *datastoreProvider) DeleteCredentials(ctx context.Context,
	host string) error {
	client, err := d.connector.NewClient(ctx, d.projectID)
	if err != nil {
		log.WithError(err).Errorf("Error while creating datastore client")
		return err
	}
	defer client.Close()

	log.Debugf("Deleting credentials for %v from namespace %v", host, d.namespace)

	// Remove entity with key=hostname
	key := datastore.NameKey(kind, host, nil)
	key.Namespace = d.namespace
	err = client.Delete(ctx, key)
	if err != nil {
		log.WithError(err).Errorf("Error deleting entity %s", host)
		return err
	}
	return nil
}
