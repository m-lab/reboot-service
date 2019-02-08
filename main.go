package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/m-lab/reboot-service/drac"
)

const defaultDRACPort = 806

var (
	fPrivateKeyPath string
	fListenPort     int
)

func init() {
	log.SetFlags(log.LUTC)
	flag.IntVar(&fListenPort, "port", 4040, "Port to listen on")
	flag.StringVar(&fPrivateKeyPath, "private-key-path", "", "Private key path")
}

// reboot retrieves credentials from Datastore, logs into a DRAC and sends
// a reboot command to the given hostname.
func reboot(host string) (string, error) {
	username, password, err := FindCredentials(host)

	if err != nil {
		log.Println("ERROR: cannot fetch credentials:", err)
		return "", err
	}

	conn, err := drac.NewConnection(host, defaultDRACPort, username, password, fPrivateKeyPath)

	if err != nil {
		log.Println("ERROR: cannot initialize connection:", err)
		return "", err
	}
	output, err := conn.Reboot()

	if err != nil {
		log.Println(err)
	}

	return output, err
}

// RebootHandler is the GET handler for /reboot
func RebootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	hosts, ok := r.URL.Query()["host"]

	if !ok || len(hosts[0]) < 1 {
		log.Println("URL parameter 'host' is missing")
		return
	}

	host := hosts[0]
	output, err := reboot(host)

	if err != nil {
		log.Println("ERROR: RebootHandler:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write([]byte(output))
}

func main() {
	http.HandleFunc("/reboot", RebootHandler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(fListenPort), nil))
}
