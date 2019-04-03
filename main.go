// Command reboot-api starts a HTTP server providing endpoints that allow
// the user to interact with nodes on M-Lab's infrastructure.
package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/m-lab/reboot-service/connector"

	"github.com/m-lab/reboot-service/creds"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/httpx"
	"github.com/m-lab/go/prometheusx"
	"github.com/m-lab/go/rtx"
	"github.com/m-lab/reboot-service/reboot"

	log "github.com/sirupsen/logrus"
)

var (
	// Command line flags.
	listenAddr = flag.String("listenaddr", defaultListenAddr, "Address to listen on")
	promAddr   = flag.String("promaddr", defaultPromPort,
		"Address to listen on for Prometheus metrics")
	projectID = flag.String("project", defaultProjID, "GCD project ID")
	namespace = flag.String("namespace", defaultNamespace, "GCD namespace")
	sshPort   = flag.Int("sshport", defaultSSHPort, "SSH port to use")
	dracPort  = flag.Int("dracport", defaultDRACPort, "DRAC port to use")

	// Context for the whole program.
	ctx, cancel = context.WithCancel(context.Background())
)

const (
	defaultListenAddr = ":8080"
	defaultPromPort   = ":8081"
	defaultProjID     = "mlab-sandbox"
	defaultNamespace  = "reboot-api"
	defaultSSHPort    = 22
	defaultDRACPort   = 806
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func createRebootConfig() *reboot.Config {
	// Initialize configuration based on passed flags.
	return &reboot.Config{
		Namespace: *namespace,
		ProjectID: *projectID,
		SSHPort:   int32(*sshPort),
		DRACPort:  int32(*dracPort),
	}
}

func main() {
	// TODO(roberto): create end-to-end test that calls main() verifies that
	// the "wiring" does not cause any crashes.
	flag.Parse()
	rtx.Must(flagx.ArgsFromEnv(flag.CommandLine), "Cannot parse env args")

	// Initialize configuration, credentials provider and connector.
	rebootConfig := createRebootConfig()
	credentials := creds.NewProvider(*projectID, *namespace)
	connector := connector.NewConnector()

	rebootHandler := reboot.NewHandler(rebootConfig, credentials, connector)

	// Initialize HTTP server.
	rebootMux := http.NewServeMux()
	rebootMux.Handle("/v1/reboot", rebootHandler)

	s := &http.Server{
		Addr:    *listenAddr,
		Handler: rebootMux,
	}
	rtx.Must(httpx.ListenAndServeAsync(s), "Could not start HTTP server")
	defer s.Close()

	// Initialize Prometheus server for monitoring.
	prometheusx.MustStartPrometheus(*promAddr)

	// Keep serving until the context is canceled.
	<-ctx.Done()
}
