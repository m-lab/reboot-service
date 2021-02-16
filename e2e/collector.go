package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/apex/log"
	"github.com/m-lab/reboot-service/connector"
	"github.com/m-lab/reboot-service/creds"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	statusOK               = "ok"
	reasonSuccess          = "success"
	reasonCredsNotFound    = "credentials_not_found"
	reasonConnectionFailed = "connection_failed"

	// Timeout for the e2e test must be shorter than Prometheus' timeout.
	connectionTimeout = 45 * time.Second
)

type collectorConfig struct {
	bmcPort   int32
	provider  creds.Provider
	connector connector.Connector
}

type e2eTestCollector struct {
	target       string
	config       *collectorConfig
	resultMetric *prometheus.Desc
}

func newE2ETestCollector(target string, config *collectorConfig) *e2eTestCollector {
	return &e2eTestCollector{
		target: target,
		config: config,
		resultMetric: prometheus.NewDesc("reboot_e2e_result",
			"E2E test result for this target", []string{"target", "status", "reason"},
			nil),
	}
}

func (c *e2eTestCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.resultMetric
}

func (c *e2eTestCollector) Collect(ch chan<- prometheus.Metric) {
	// Get credentials for this BMC using the configured provider.
	creds, err := c.getCredentials(c.target)
	if err != nil {
		log.Errorf("Error while getting credentials for %s: %v", c.target, err)
		ch <- prometheus.MustNewConstMetric(c.resultMetric,
			prometheus.GaugeValue, 0, c.target, statusOK, reasonCredsNotFound)
		return
	}

	// We've got credentials, let's try to SSH.
	config := &connector.ConnectionConfig{
		ConnType: connector.BMCConnection,
		Hostname: c.target,
		Port:     c.config.bmcPort,
		Username: creds.Username,
		Password: creds.Password,
		Timeout:  connectionTimeout,
	}
	conn, err := c.config.connector.NewConnection(config)
	if err != nil {
		// TODO: here we should be able to distinguish different errors.
		log.Errorf("Error while creating connection to %s: %v", c.target, err)
		ch <- prometheus.MustNewConstMetric(c.resultMetric,
			prometheus.GaugeValue, 0, c.target, statusOK, reasonConnectionFailed)
		return
	}

	// TODO: execute a no-op command?
	conn.Close()

	ch <- prometheus.MustNewConstMetric(c.resultMetric, prometheus.GaugeValue,
		1, c.target, statusOK, reasonSuccess)
}

func (c *e2eTestCollector) getCredentials(hostname string) (*creds.Credentials, error) {
	creds, err := c.config.provider.FindCredentials(context.Background(), hostname)
	if err != nil {
		return nil, fmt.Errorf("Cannot retrieve credentials: %v", err)
	}

	return creds, nil
}
