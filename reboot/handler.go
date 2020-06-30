package reboot

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/m-lab/go/host"
	"github.com/m-lab/reboot-service/connector"
	"github.com/m-lab/reboot-service/creds"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const bmcTimeout = 60 * time.Second

var (
	metricBMCReboots = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "reboot_bmc_total",
			Help: "Total number of successful BMC reboots",
		},
		[]string{
			"site",
			"machine",
			"status",
		},
	)
	metricBMCRebootTimeHist = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "reboot_bmc_duration_seconds",
		Help:    "Duration histogram for successful BMC reboots, in seconds",
		Buckets: []float64{15, 30, 45, 60},
	})
	metricHostReboots = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "reboot_host_total",
			Help: "Total number of successful host reboots",
		},
		[]string{
			"site",
			"machine",
			"status",
		},
	)
)

// Config holds the configuration for the reboot handler
type Config struct {
	ProjectID string
	Namespace string

	SSHPort int32
	BMCPort int32

	RebootUser     string
	PrivateKeyPath string
}

// NewHandler creates a new Handler for the /v1/reboot endpoint.
// Configuration, credential provider and connector need to be passed as
// arguments.
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

func (h *Handler) rebootHost(ctx context.Context, node host.Name) (string, error) {

	// Connect to the host
	connectionConfig := &connector.ConnectionConfig{
		Hostname:       node.String(),
		Username:       h.config.RebootUser,
		Port:           h.config.SSHPort,
		PrivateKeyFile: h.config.PrivateKeyPath,
		ConnType:       connector.HostConnection,
	}

	conn, err := h.connector.NewConnection(connectionConfig)
	if err != nil {
		log.WithError(err).
			Errorf("Cannot connect to host: %s:%d with username %s",
				connectionConfig.Hostname, connectionConfig.Port, connectionConfig.Username)
		metricHostReboots.WithLabelValues(node.Site, node.Machine, "error-connect").Inc()
		return "", err
	}
	defer conn.Close()

	_, err = conn.Reboot()
	if err != nil {
		log.WithError(err).Errorf("Cannot issue reboot command (type: %v)", connectionConfig.ConnType)
		metricHostReboots.WithLabelValues(node.Site, node.Machine, "error-reboot").Inc()
		return "", err
	}

	metricHostReboots.WithLabelValues(node.Site, node.Machine, "ok").Inc()
	return "System reboot successful", nil
}

func (h *Handler) rebootBMC(ctx context.Context, node host.Name) (string, error) {
	// Retrieve credentials from the credentials provider.
	creds, err := h.credsProvider.FindCredentials(ctx, node.String())
	if err != nil {
		log.WithError(err).Errorf("Cannot retrieve credentials for host: %v", node.String())
		return "", err
	}

	// Make a connection to the host
	connectionConfig := &connector.ConnectionConfig{
		Hostname:       creds.Address,
		Username:       creds.Username,
		Password:       creds.Password,
		Port:           h.config.BMCPort,
		PrivateKeyFile: h.config.PrivateKeyPath,
		ConnType:       connector.BMCConnection,
		Timeout:        bmcTimeout,
	}

	conn, err := h.connector.NewConnection(connectionConfig)
	if err != nil {
		log.WithError(err).
			Errorf("Cannot connect to DRAC: %s:%d with username %s",
				connectionConfig.Hostname, connectionConfig.Port, connectionConfig.Username)
		metricBMCReboots.WithLabelValues(node.Site, node.Machine, "error-connect").Inc()
		return "", err
	}
	defer conn.Close()

	start := time.Now()
	output, err := conn.Reboot()
	if err != nil {
		log.WithError(err).Errorf("Cannot issue reboot command")
		metricBMCReboots.WithLabelValues(node.Site, node.Machine, "error-reboot").Inc()
		return "", err
	}

	metricBMCReboots.WithLabelValues(node.Site, node.Machine, "ok").Inc()
	metricBMCRebootTimeHist.Observe(time.Since(start).Seconds())
	return output, nil
}

// ServeHTTP handles POST requests to the /reboot endpoint
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	target := r.URL.Query().Get("host")
	if len(target) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("URL parameter 'host' is missing"))
		log.Info("URL parameter 'host' is missing")
		return
	}

	// Split hostname into site/node. If site and node cannot be extracted,
	// we are reasonably sure this is not a valid M-Lab node's BMC.
	node, err := host.Parse(target)
	if err != nil {
		errStr := fmt.Sprintf(
			"The specified hostname is not a valid M-Lab node: %s", target)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errStr))
		log.Errorf(errStr)
		return
	}

	method := r.URL.Query().Get("method")
	var output string
	if method == "host" {
		output, err = h.rebootHost(context.Background(), node)
	} else { // default method is DRAC
		output, err = h.rebootBMC(context.Background(), node)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Reboot failed: %v", err)))
		log.WithError(err).Error("Reboot failed")
		return
	}

	log.WithField("output", output).Infof("%v rebooted successfully.",
		node.String())
	w.Write([]byte(output))
}
