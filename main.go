package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
	}

	gateway := net.ParseIP(os.Args[3])
	if nil == gateway {
		log.Fatalln("could not parse gateway IP: " + os.Args[2])
	}

	action := os.Args[1]
	domain := os.Args[2]

	switch action {
	case "keep":
		keepEntry(domain, gateway)
		break

	case "delete":
		deleteEntry(domain, gateway)
		break

	default:
		printUsage()
		break
	}
}

func printUsage() {
	log.Fatalln(`usage:

  domroute keep <DOMAIN> <GATEWAY>
  domroute delete <DOMAIN> <GATEWAY>`)
}

func keepEntry(domain string, gateway net.IP) {
	checkInterval := getCheckInterval()
	log.Printf("keeping %s routed to %s - checking every %v seconds", domain, gateway.String(), checkInterval.Seconds())

	ticker := time.NewTicker(checkInterval)
	done := make(chan bool)

	ensureRoute(domain, gateway)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			ensureRoute(domain, gateway)
		}
	}
}

func deleteEntry(domain string, gateway net.IP) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to resolve domain: " + domain))
	}
	log.Printf("resolved %s to %s", domain, ips)

	for _, ip := range ips {
		err := deleteRoute(domain, ip.String(), gateway.String())
		if err != nil {
			log.Printf("failed to delete route: %s->%s", ip.String(), gateway)
			continue
		}
	}
}
