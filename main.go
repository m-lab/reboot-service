package main

import (
	"log"

	"github.com/evfirerob/reboot-service/drac"
)

const defaultDRACPort = 806

func DracReboot(host string) {
	username, password, err := FindCredentials(host)

	if err != nil {
		log.Println(err)
		return
	}

	conn := drac.NewConnection(host, defaultDRACPort, username, password, "")
	output, err := conn.Reboot()

	if err != nil {
		log.Println(err)
	}

	log.Println(output)
}

func main() {
	DracReboot("mlab4d.lga0t.measurement-lab.org")
}
