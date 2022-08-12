package main

import (
	"github.com/DeloitteDigitalUK/domroute/config"
	"github.com/DeloitteDigitalUK/domroute/route"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	config.InitLogger()
	if len(os.Args) == 1 {
		printUsage(0)
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

	case "delete":
		deleteEntry(domain, gateway)
		break

	case "keep":
		keepEntry(domain, gateway)
		break

	case "purge":
		deleteAll()
		break

	default:
		printUsage(1)
		break
	}
}

func printUsage(code int) {
	println(`DOMROUTE
Usage:

add:
  Route traffic for a given domain via a destination gateway.

  domroute add <DOMAIN> <GATEWAY_IP>
  domroute add <DOMAIN> <GATEWAY_CIDR>
  domroute add <DOMAIN> <INTERFACE_NAME>

delete:
  Delete a domain route added by this tool.

  domroute delete <DOMAIN> <GATEWAY_IP>
  domroute delete <DOMAIN> <GATEWAY_CIDR>
  domroute delete <DOMAIN> <INTERFACE_NAME>

keep:
  Periodically update resolution for domain and gateway,
  routing traffic for a given domain via a destination gateway.

  domroute keep <DOMAIN> <GATEWAY_IP>
  domroute keep <DOMAIN> <GATEWAY_CIDR>
  domroute keep <DOMAIN> <INTERFACE_NAME>

purge:
  Remove all routes added by this tool.

  domroute purge`)

	os.Exit(code)
}

func addEntry(domain string, gateway string) {
	if err := route.EnsureExists(domain, gateway); err != nil {
		log.Fatalf("failed to add entry %s for gateway %s: %s", domain, gateway, err)
	}
}

func keepEntry(domain string, gateway string) {
	checkInterval := config.GetCheckInterval()
	log.Debugf("keeping %s routed to %s - checking every %v seconds", domain, gateway, checkInterval.Seconds())

	ticker := time.NewTicker(checkInterval)
	done := make(chan bool)
	due := make(chan bool)

	go func() {
		for {
			due <- true
			<-ticker.C
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-due:
			if err := route.EnsureExists(domain, gateway); err != nil {
				log.Warnf("failed to ensure entry %s for gateway %s: %s - will retry", domain, gateway, err)
			}
		}
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
