package main

import (
	"github.com/pcornish/domroute/route"
	"github.com/pcornish/domroute/state"
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
	case "add":
		addEntry(domain, gateway)
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

  domroute add <DOMAIN> <GATEWAY>
  domroute delete <DOMAIN> <GATEWAY>
  domroute keep <DOMAIN> <GATEWAY>`)
}

func addEntry(domain string, gateway net.IP) {
	route.EnsureExists(domain, gateway)
}

func keepEntry(domain string, gateway net.IP) {
	checkInterval := getCheckInterval()
	log.Printf("keeping %s routed to %s - checking every %v seconds", domain, gateway.String(), checkInterval.Seconds())

	ticker := time.NewTicker(checkInterval)
	done := make(chan bool)

	route.EnsureExists(domain, gateway)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			route.EnsureExists(domain, gateway)
		}
	}
}

func deleteEntry(domain string, gateway net.IP) {
	routes, err := state.ReadRoutesForDomain(domain, gateway)
	if err != nil {
		log.Printf("failed to load existing entries from state: %s", err)
	} else {
		log.Printf("found %d existing entries in state", len(routes))
		for _, e := range routes {
			deleteGatewayEntry(domain, gateway, e.Ip)
		}
	}

	ips, err := net.LookupIP(domain)
	if err != nil {
		log.Println("failed to resolve domain: " + domain)
	} else {
		log.Printf("resolved %s to %s", domain, ips)
		for _, ip := range ips {
			deleteGatewayEntry(domain, gateway, ip.String())
		}
	}
}

func deleteGatewayEntry(domain string, gateway net.IP, ip string) {
	err := route.DeleteIfExists(domain, ip, gateway.String())
	if err != nil {
		log.Printf("failed to delete route: %s->%s", ip, gateway)
	}
}
