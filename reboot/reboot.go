// Package reboot provides the business logic to reboot a node via DRAC or
// SSH access.
package reboot

import (
	"context"
	"log"

	"github.com/m-lab/reboot-service/drac"

	"cloud.google.com/go/datastore"
	"github.com/m-lab/reboot-service/storage"
)

var dsNewClient = datastore.NewClient

// TODO(roberto): these should be specified by the caller.
const (
	projectID        = "mlab-sandbox"
	defaultNamespace = "reboot-api"
)

func retrieveDRACCredentials(ctx context.Context, host string) (*storage.Credentials, error) {
	client, err := dsNewClient(ctx, projectID)

	if err != nil {
		return nil, err
	}

	return storage.FindCredentials(ctx, storage.Datastore{
		Client:    client,
		Namespace: defaultNamespace,
	}, host)
}

// DRAC reboots a node via its DRAC.
func DRAC(ctx context.Context, host string, port int32) (string, error) {
	cred, err := retrieveDRACCredentials(ctx, host)
	if err != nil {
		log.Printf("Cannot retrieve DRAC credentials: %v", err)
		return "", err
	}

	conn, err := drac.NewConnection(host, port, cred.Username, cred.Password, "", &drac.DialerImpl{})
	if err != nil {
		log.Printf("Cannot initialize DRAC connection %v", err)
		return "", err
	}

	output, err := conn.Reboot()
	if err != nil {
		log.Printf("Cannot send reboot command: %v", err)
		return "", err
	}

	return output, nil
}
