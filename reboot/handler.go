package reboot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/apex/log"
	"github.com/m-lab/reboot-service/connector"
	"github.com/m-lab/reboot-service/creds"
)

// Config holds the configuration for the reboot handler
type Config struct {
	ProjectID string
	Namespace string

	SSHPort  int32
	DRACPort int32

	PrivateKeyPath string
}

func NewHandler(config *Config, credsProvider creds.Provider, connector connector.Connector) *Handler {
	return &Handler{
		config:        config,
		credsProvider: credsProvider,
		connector:     connector,
	}
}

// Handler is the HTTP handler for /reboot
type Handler struct {
	config *Config

	credsProvider creds.Provider
	connector     connector.Connector
}

func (h *Handler) rebootDRAC(ctx context.Context, host string) (string, error) {
	// Retrieve credentials from the credentials provider.
	creds, err := h.credsProvider.FindCredentials(ctx, host)
	if err != nil {
		log.WithError(err).Errorf("Cannot retrieve credentials for host: %v", host)
		return "", err
	}

	// Make a connection to the host
	connectionConfig := &connector.ConnectionConfig{
		Hostname:       creds.Address,
		Username:       creds.Username,
		Password:       creds.Password,
		Port:           h.config.DRACPort,
		PrivateKeyFile: h.config.PrivateKeyPath,
		ConnType:       connector.DRACConnection,
	}

	conn, err := h.connector.NewConnection(connectionConfig)
	if err != nil {
		log.WithError(err).
			Errorf("Cannot connect to host: %s:%d with username %s",
				connectionConfig.Hostname, connectionConfig.Port, connectionConfig.Username)
		return "", err
	}

	output, err := conn.Reboot()
	if err != nil {
		log.WithError(err).Errorf("Cannot issue reboot command")
		return "", err
	}

	return output, nil
}

// ServeHTTP handles POST requests to the /reboot endpoint
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	host := r.URL.Query().Get("host")
	if len(host) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("URL parameter 'host' is missing"))
		log.Info("URL parameter 'host' is missing")
		return
	}

	output, err := h.rebootDRAC(context.Background(), host)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Reboot failed: %v", err)))
		log.WithError(err).Warn("Reboot failed")
		return
	}

	w.Write([]byte(output))

}
