package e2e

import (
	"github.com/prometheus/client_golang/prometheus"
)

type bmcE2ECollector struct {
	target       string
	resultMetric *prometheus.Desc
}

func newBMCE2ECollector(target string) *bmcE2ECollector {
	return &bmcE2ECollector{
		target: target,
		resultMetric: prometheus.NewDesc("reboot_e2e_result",
			"E2E test result for this target", []string{"target"}, nil),
	}
}

func (c *bmcE2ECollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.resultMetric
}

func (c *bmcE2ECollector) Collect(ch chan<- prometheus.Metric) {
	// TODO: actual collect metrics from the BMC.
	ch <- prometheus.MustNewConstMetric(c.resultMetric, prometheus.GaugeValue, 1, c.target)
}
