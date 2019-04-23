// Command reboot-api starts a HTTP server providing endpoints that allow
// the user to interact with nodes on M-Lab's infrastructure.
package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/goji/httpauth"

	"github.com/apex/log"
	"github.com/m-lab/reboot-service/connector"

	"github.com/m-lab/reboot-service/creds"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/httpx"
	"github.com/m-lab/go/prometheusx"
	"github.com/m-lab/go/rtx"
	"github.com/m-lab/reboot-service/reboot"
)

var (
	// Command line flags.
	listenAddr = flag.String("listenaddr", defaultListenAddr, "Address to listen on")
	projectID  = flag.String("datastore.project", defaultProjID, "GCD project ID")
	namespace  = flag.String("datastore.namespace", defaultNamespace, "GCD namespace")
	rebootUser = flag.String("reboot.user", defaultRebootUser, "User for rebooting CoreOS hosts")
	keyPath    = flag.String("reboot.key", "", "SSH private key path")

	sshPort = flag.Int("reboot.sshport", defaultSSHPort, "SSH port to use")
	bmcPort = flag.Int("reboot.bmcport", defaultBMCPort, "DRAC port to use")

	username = flag.String("auth.username", "", "Username for HTTP basic auth")
	password = flag.String("auth.password", "", "Password for HTTP basic auth")

	// Context for the whole program.
	ctx, cancel = context.WithCancel(context.Background())
)

const (
	defaultListenAddr = ":8080"
	defaultPromPort   = ":9600"
	defaultProjID     = "mlab-sandbox"
	defaultNamespace  = "reboot-api"
	defaultSSHPort    = 22
	defaultBMCPort    = 806
	defaultRebootUser = "reboot-api"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func createRebootConfig() *reboot.Config {
	// Initialize configuration based on passed flags.
	return &reboot.Config{
		Namespace: *namespace,
		ProjectID: *projectID,
		SSHPort:   int32(*sshPort),
		BMCPort:   int32(*bmcPort),

		RebootUser:     *rebootUser,
		PrivateKeyPath: *keyPath,
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

	var rebootHandler http.Handler
	rebootHandler = reboot.NewHandler(rebootConfig, credentials, connector)

	// Initialize HTTP server.
	// TODO(roberto): add promhttp instruments for handlers.
	if *username != "" && *password != "" {
		authOpts := httpauth.AuthOptions{
			Realm:    "reboot-api",
			User:     *username,
			Password: *password,
		}
		rebootHandler = httpauth.BasicAuth(authOpts)(rebootHandler)
	} else {
		log.Warn("Username and password have not been specified!")
		log.Warn("Make sure you add -auth.username and -auth.password before " +
			"running in production.")
	}

	rebootMux := http.NewServeMux()
	rebootMux.Handle("/v1/reboot", rebootHandler)

	s := &http.Server{
		Addr:    *listenAddr,
		Handler: rebootMux,
	}
	rtx.Must(httpx.ListenAndServeAsync(s), "Could not start HTTP server")
	defer s.Close()

	// Initialize Prometheus server for monitoring.
	promServer := prometheusx.MustServeMetrics()
	defer promServer.Close()

	// Keep serving until the context is canceled.
	<-ctx.Done()
}
