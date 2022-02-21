package main

import (
	"fmt"
	"github.com/pcornish/domroute/route"
	"github.com/pcornish/domroute/state"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
	}

	gateway, err := getGatewayAddr(os.Args[3])
	if err != nil {
		log.Fatalf("could not parse gateway: %s", err)
	} else if nil == gateway {
		log.Fatalln("could not parse gateway: " + os.Args[3])
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

  domroute add <DOMAIN> <GATEWAY_IP>
  domroute add <DOMAIN> <INTERFACE_NAME>

  domroute delete <DOMAIN> <GATEWAY_IP>
  domroute delete <DOMAIN> <INTERFACE_NAME>

  domroute keep <DOMAIN> <GATEWAY_IP>
  domroute keep <DOMAIN> <INTERFACE_NAME>`)
}

func getGatewayAddr(ipOrIface string) (net.IP, error) {
	if matched, _ := regexp.MatchString(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`, ipOrIface); matched {
		return net.ParseIP(ipOrIface), nil
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("could not list network interfaces: %s", err)
	}
	for _, iface := range interfaces {
		if iface.Name != ipOrIface {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("failed to get addresses for interface: %s: %s", iface.Name, err)
		} else if len(addrs) == 0 {
			return nil, fmt.Errorf("no addresses for interface: %s", iface.Name)
		}
		ip := addrs[0].String()
		if strings.Contains(ip, "/") {
			ip = strings.Split(ip, "/")[0]
		}
		log.Printf("resolved gateway interface %s to %s", ipOrIface, ip)
		return net.ParseIP(ip), nil
	}
	return nil, fmt.Errorf("failed to resolve gateway: %s", ipOrIface)
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
