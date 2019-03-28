// Command reboot-api starts a HTTP server providing endpoints that allow
// the user to interact with nodes on M-Lab's infrastructure.
package main

import (
	"flag"
	"net/http"

	"github.com/m-lab/reboot-service/connector"

	"github.com/m-lab/reboot-service/creds"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/rtx"
	"github.com/m-lab/reboot-service/reboot"

	"cloud.google.com/go/datastore"
	log "github.com/sirupsen/logrus"
)

var (
	listenAddr  string
	projectID   string
	namespace   string
	sshPort     int
	dracPort    int
	dsNewClient = datastore.NewClient

	rebootConfig *reboot.Config
)

const (
	defaultProjID    = "mlab-sandbox"
	defaultNamespace = "reboot-api"
	defaultSSHPort   = 22
	defaultDRACPort  = 806
)

func init() {
	log.SetLevel(log.DebugLevel)
	flag.StringVar(&listenAddr, "listenaddr", ":8080", "Address to listen on")
	flag.StringVar(&projectID, "project", defaultProjID, "GCD project ID")
	flag.StringVar(&namespace, "namespace", defaultNamespace, "GCD namespace")
	flag.IntVar(&sshPort, "sshport", defaultSSHPort, "SSH port to use")
	flag.IntVar(&dracPort, "dracport", defaultDRACPort, "DRAC port to use")
}

func createRebootConfig() *reboot.Config {
	// Initialize configuration based on passed flags.
	return &reboot.Config{
		Namespace: namespace,
		ProjectID: projectID,
		SSHPort:   int32(sshPort),
		DRACPort:  int32(dracPort),
	}
}

func main() {
	flag.Parse()
	rtx.Must(flagx.ArgsFromEnv(flag.CommandLine), "Cannot parse env args")

	rebootConfig = createRebootConfig()

	credentials := creds.NewProvider(projectID, namespace)
	connector := connector.NewConnector()
	rebootHandler := reboot.NewHandler(rebootConfig, credentials, connector)

	http.Handle("/v1/reboot", rebootHandler)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
