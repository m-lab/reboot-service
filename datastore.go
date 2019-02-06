package main

import (
	"context"
	"log"

	"cloud.google.com/go/datastore"
)

const projectID = "mlab-sandbox"
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

// FindCredentials retrieves a username/password pair from Google Cloud
// Datastore for a given hostname.
func FindCredentials(host string) (string, string, error) {
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}

	query := datastore.NewQuery("Credentials")
	query = query.Filter("hostname = ", host)

	var creds []*Credentials
	_, err = client.GetAll(ctx, query, &creds)

	if err != nil {
		return "", "", err
	}

	if len(creds) == 0 {
		return "", "", err
	}

	cred := creds[0]
	return cred.Username, cred.Password, nil

}
