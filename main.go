// Command reboot-api starts a HTTP server providing endpoints that allow
// the user to interact with nodes on M-Lab's infrastructure.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

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

func handleDRACReboot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	host := r.URL.Query().Get("host")
	if len(host) == 0 {
		w.Write([]byte("URL parameter 'host' is missing"))
		log.Info("URL parameter 'host' is missing")
		return
	}

	output, err := reboot.DRAC(context.Background(), rebootConfig, host)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Reboot failed: %v", err)))
		log.WithError(err).Warn("Reboot failed")
		return
	}

	w.Write([]byte(output))
}

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

	http.HandleFunc("/v1/reboot", handleDRACReboot)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
