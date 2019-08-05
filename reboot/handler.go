package reboot

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/apex/log"
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

func (h *Handler) rebootHost(ctx context.Context, node string, site string) (string, error) {
	host := makeHostname(node, site)

	// Connect to the host
	connectionConfig := &connector.ConnectionConfig{
		Hostname:       host,
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
		metricHostReboots.WithLabelValues(site, node, "error-connect").Inc()
		return "", err
	}
	defer conn.Close()

	_, err = conn.Reboot()
	if err != nil {
		log.WithError(err).Errorf("Cannot issue reboot command (type: %v)", connectionConfig.ConnType)
		metricHostReboots.WithLabelValues(site, node, "error-reboot").Inc()
		return "", err
	}

	metricHostReboots.WithLabelValues(site, node, "ok").Inc()
	return "System reboot successful", nil
}

func (h *Handler) rebootBMC(ctx context.Context, node string, site string) (string, error) {
	// There are different ways a BMC hostname can be provided:
	// - mlab1.lga0t
	// - mlab1d.lga0t
	// - mlab1.lga0t.measurement-lab.org
	// - mlab1d.lga0t.measurement-lab.org
	// To make sure this is handled in a flexible way, the site and host parts
	// are provided separately and re-assembled here.
	host := makeBMCHostname(node, site)

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
		metricBMCReboots.WithLabelValues(site, node, "error-connect").Inc()
		return "", err
	}
	defer conn.Close()

	start := time.Now()
	output, err := conn.Reboot()
	if err != nil {
		log.WithError(err).Errorf("Cannot issue reboot command")
		metricBMCReboots.WithLabelValues(site, node, "error-reboot").Inc()
		return "", err
	}

	metricBMCReboots.WithLabelValues(site, node, "ok").Inc()
	metricBMCRebootTimeHist.Observe(time.Since(start).Seconds())
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

	// Split hostname into site/node. If site and node cannot be extracted,
	// we are reasonably sure this is not a valid M-Lab node's BMC.
	node, site, err := parseNodeSite(host)
	if err != nil {
		errStr := fmt.Sprintf(
			"The specified hostname is not a valid M-Lab node: %s", host)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errStr))
		log.Errorf(errStr)
		return
	}

	method := r.URL.Query().Get("method")
	var output string
	if method == "host" {
		output, err = h.rebootHost(context.Background(), node, site)
	} else { // default method is DRAC
		output, err = h.rebootBMC(context.Background(), node, site)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Reboot failed: %v", err)))
		log.WithError(err).Error("Reboot failed")
		return
	}

	log.WithField("output", output).Infof("%v rebooted successfully.", host)
	w.Write([]byte(output))
}

// parseNodeSite extracts node and site from a full hostname.
func parseNodeSite(hostname string) (string, string, error) {
	regex := regexp.MustCompile("(mlab[1-4]d?)\\.([a-zA-Z]{3}[0-9t]{2}).*")
	result := regex.FindStringSubmatch(hostname)
	if len(result) != 3 {
		return "", "",
			fmt.Errorf("The specified hostname is not a valid M-Lab node: %s", hostname)
	}

	return result[1], result[2], nil
}

// makeBMCHostname returns a full BMC hostname made from the specified node
// and site (node + 'd' + site + "measurement-lab.org").
func makeBMCHostname(node string, site string) string {
	if node[len(node)-1] != 'd' {
		node = node + "d"
	}

	return fmt.Sprintf("%s.%s.measurement-lab.org", node, site)
}

func makeHostname(node string, site string) string {
	return fmt.Sprintf("%s.%s.measurement-lab.org", node, site)
}
