package main

import (
	"github.com/pcornish/domroute/config"
	"github.com/pcornish/domroute/route"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	config.InitLogger()
	if len(os.Args) == 1 {
		printUsage()
	}
	action := os.Args[1]
	var (
		domain  string
		gateway string
	)
	if len(os.Args) == 4 {
		domain = os.Args[2]
		gateway = os.Args[3]
	}

	switch action {
	case "add":
		addEntry(domain, gateway)
		break

	case "cleanup":
		deleteAll()
		break

	case "delete":
		deleteEntry(domain, gateway)
		break

	case "keep":
		keepEntry(domain, gateway)
		break

	default:
		printUsage()
		break
	}
}

func printUsage() {
	log.Fatalln(`usage:

  domroute add <DOMAIN> <GATEWAY_IP>
  domroute add <DOMAIN> <INTERFACE_NAME>

  domroute delete <DOMAIN> <GATEWAY_IP>
  domroute delete <DOMAIN> <INTERFACE_NAME>

  domroute keep <DOMAIN> <GATEWAY_IP>
  domroute keep <DOMAIN> <INTERFACE_NAME>

  domroute cleanup`)
}

func addEntry(domain string, gateway string) {
	ensureExists(domain, gateway)
}

func keepEntry(domain string, gateway string) {
	checkInterval := config.GetCheckInterval()
	log.Debugf("keeping %s routed to %s - checking every %v seconds", domain, gateway, checkInterval.Seconds())

	ticker := time.NewTicker(checkInterval)
	done := make(chan bool)

	ensureExists(domain, gateway)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			ensureExists(domain, gateway)
		}
	}
}

func ensureExists(domain string, gateway string) {
	err := route.EnsureExists(domain, gateway)
	if err != nil {
		log.Fatalf("could not ensure entry exists: %s", err)
	}
}

func deleteAll() {
	err := route.DeleteAllEntries()
	if err != nil {
		log.Fatalf("could not delete all entries: %s", err)
	}
}

func deleteEntry(domain string, gateway string) {
	err := route.DeleteEntry(domain, gateway)
	if err != nil {
		log.Fatalf("could not delete entry: %s", err)
	}
}
