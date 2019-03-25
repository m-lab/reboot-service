// Package storage contains types and functions to retrieve data from Google
// Cloud Datastore.
package storage

import (
	"context"
	"errors"

	"cloud.google.com/go/datastore"
	"github.com/m-lab/reboot-service/storage/iface"
	log "github.com/sirupsen/logrus"
)

const datastoreKind = "Credentials"

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

// Datastore is a struct holding the DatastoreClient and the namespace to use.
type Datastore struct {
	Client    iface.DatastoreClient
	Namespace string
}

// FindCredentials retrieves a username/password pair from a DatastoreClient
// for a given hostname.
func FindCredentials(ctx context.Context, ds Datastore,
	host string) (*Credentials, error) {
	log.Debugf("Retrieving credentials for %v from namespace %v", host, ds.Namespace)

	query := datastore.NewQuery(datastoreKind).Namespace(ds.Namespace)
	query = query.Filter("hostname = ", host)

	var creds []*Credentials
	_, err := ds.Client.GetAll(ctx, query, &creds)

	if err != nil {
		return nil, err
	}

	if len(creds) == 0 {
		return nil, errors.New("Hostname not found in Datastore")
	}

	cred := creds[0]
	return cred, nil
}
