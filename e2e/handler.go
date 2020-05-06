// Package e2e contains the handler to perform an end-to-end connectivity test
// on a given BMC module on the M-Lab infrastructure.
package e2e

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/apex/log"
	"github.com/m-lab/go/host"
	"github.com/m-lab/reboot-service/connector"
	"github.com/m-lab/reboot-service/creds"
)

// Handler is the HTTP handler for /e2e
type Handler struct {
	bmcPort int32

	connector connector.Connector
	provider  creds.Provider
}

// NewHandler returns a Handler with the specified configuration.
func NewHandler(bmcPort int32, prov creds.Provider, connector connector.Connector) *Handler {
	return &Handler{
		bmcPort:   bmcPort,
		connector: connector,
		provider:  prov,
	}
}

// ServeHTTP handles GET requests to the /e2e endpoint, parsing the target
// parameter and delegating writing the actual response to promhttp.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	target := r.URL.Query().Get("target")
	if len(target) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("URL parameter 'target' is missing"))
		log.Info("URL parameter 'target' is missing")
		return
	}

	// Parses the target parameter. If a valid BMC hostname cannot be extracted
	// we are reasonably sure this is not a valid M-Lab node's BMC.
	bmcName, err := host.Parse(target)
	if err != nil {
		errStr := fmt.Sprintf(target)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errStr))
		log.Errorf(errStr)
		return
	}

	collectorConfig := &collectorConfig{
		bmcPort:   h.bmcPort,
		connector: h.connector,
		provider:  h.provider,
	}

	registry := prometheus.NewRegistry()
	collector := newE2ETestCollector(bmcName.String(), collectorConfig)
	registry.MustRegister(collector)
	promHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	promHandler.ServeHTTP(w, r)
}
