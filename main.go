package main

import (
	"log"

	"github.com/m-lab/reboot-service/drac"
)

const defaultDRACPort = 806

// DracReboot retrieves credentials from Datastore, logs into a DRAC and sends
// a reboot command to the given hostname.
func DracReboot(host string) {
	username, password, err := FindCredentials(host)

	if err != nil {
		log.Println(err)
		return
	}

	conn, err := drac.NewConnection(host, defaultDRACPort, username, password, "")

	if err != nil {
		log.Println("ERROR: cannot initialize connection: ", err)
		return
	}
	output, err := conn.Reboot()

	if err != nil {
		log.Println(err)
	}

	log.Println(output)
}

func main() {
}
