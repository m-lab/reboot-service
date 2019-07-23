package e2e

import (
	"github.com/m-lab/reboot-service/connector"
	"github.com/m-lab/reboot-service/creds"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	statusOK = "ok"
)

type collectorConfig struct {
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
	ch <- prometheus.MustNewConstMetric(c.resultMetric, prometheus.GaugeValue,
		1, c.target, statusOK)
}
