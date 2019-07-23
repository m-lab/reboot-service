package e2e

import (
	"context"
	"fmt"

	"github.com/apex/log"
	"github.com/m-lab/reboot-service/connector"
	"github.com/m-lab/reboot-service/creds"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	statusOK               = "ok"
	statusCredsNotFound    = "credentials_not_found"
	statusConnectionFailed = "connection_failed"
)

type collectorConfig struct {
	bmcPort   int32
	provider  creds.Provider
	connector connector.Connector
}

type bmcE2ECollector struct {
	target       string
	config       *collectorConfig
	resultMetric *prometheus.Desc
}

func newBMCE2ECollector(target string, config *collectorConfig) *bmcE2ECollector {
	return &bmcE2ECollector{
		target: target,
		config: config,
		resultMetric: prometheus.NewDesc("reboot_e2e_result",
			"E2E test result for this target", []string{"target", "status"},
			nil),
	}
}

func (c *bmcE2ECollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.resultMetric
}

func (c *bmcE2ECollector) Collect(ch chan<- prometheus.Metric) {
	// TODO: actually try connecting to the BMC and report result.

	// Get credentials for this BMC using the configured provider.
	creds, err := c.getCredentials(c.target)
	if err != nil {
		log.Errorf("Error while getting credentials for %s: %v", c.target, err)
		ch <- prometheus.MustNewConstMetric(c.resultMetric,
			prometheus.GaugeValue, 0, c.target, statusCredsNotFound)
		return
	}

	// We've got credentials, let's try to SSH.
	config := &connector.ConnectionConfig{
		ConnType: connector.BMCConnection,
		Hostname: c.target,
		Port:     c.config.bmcPort,
		Username: creds.Username,
		Password: creds.Password,
	}
	conn, err := c.config.connector.NewConnection(config)
	if err != nil {
		// TODO: here we should be able to distinguish different errors.
		log.Errorf("Error while creating connection to %s: %v", c.target, err)
		ch <- prometheus.MustNewConstMetric(c.resultMetric,
			prometheus.GaugeValue, 0, c.target, statusConnectionFailed)
		return
	}

	// TODO: execute a no-op command?
	conn.Close()

	ch <- prometheus.MustNewConstMetric(c.resultMetric, prometheus.GaugeValue,
		1, c.target, statusOK)
}

func (c *bmcE2ECollector) getCredentials(hostname string) (*creds.Credentials, error) {
	creds, err := c.config.provider.FindCredentials(context.Background(), hostname)
	if err != nil {
		return nil, fmt.Errorf("Cannot retrieve credentials: %v", err)
	}

	return creds, nil
}
