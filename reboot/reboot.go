// Package reboot provides the business logic to reboot a node via DRAC or
// SSH access.
package reboot

import (
	"context"

	"github.com/m-lab/reboot-service/drac"
	"github.com/m-lab/reboot-service/storage/iface"
	"google.golang.org/api/option"

	"cloud.google.com/go/datastore"
	"github.com/m-lab/reboot-service/storage"
	log "github.com/sirupsen/logrus"
)

var dsNewClient = func(ctx context.Context, projectID string, opts ...option.ClientOption) (iface.DatastoreClient, error) {
	return datastore.NewClient(ctx, projectID, opts...)
}

// Config holds the configuration for retrieving credentials and rebooting
// nodes.
type Config struct {
	ProjectID string
	Namespace string

	SSHPort  int32
	DRACPort int32

	// The Dialer to use for SSH connections - useful for unit testing.
	Dialer drac.Dialer
}

func retrieveDRACCredentials(ctx context.Context, conf *Config, host string) (*storage.Credentials, error) {
	client, err := dsNewClient(ctx, conf.ProjectID)

	if err != nil {
		return nil, err
	}

	return storage.FindCredentials(ctx, storage.Datastore{
		Client:    client,
		Namespace: conf.Namespace,
	}, host)
}

// DRAC reboots a node via its DRAC.
func DRAC(ctx context.Context, conf *Config, host string) (string, error) {
	cred, err := retrieveDRACCredentials(ctx, conf, host)
	if err != nil {
		log.Printf("Cannot retrieve DRAC credentials: %v", err)
		return "", err
	}

	conn, err := drac.NewConnection(host, conf.DRACPort, cred.Username, cred.Password, "", conf.Dialer)
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
