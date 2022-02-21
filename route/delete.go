package route

import (
	"fmt"
	"github.com/pcornish/domroute/state"
	log "github.com/sirupsen/logrus"
	"net"
	"os/exec"
)

func DeleteEntry(domain string, gateway string) error {
	gw, err := getGatewayAddr(gateway)
	if err != nil {
		return fmt.Errorf("could not parse gateway: %s", err)
	} else if "" == gateway {
		return fmt.Errorf("no gateway provided")
	}

	routes, err := state.ReadRoutesForDomain(domain, gw)
	if err != nil {
		return fmt.Errorf("failed to load existing entries from state: %s", err)
	} else {
		log.Debugf("found %d existing entries in state", len(routes))
		for _, e := range routes {
			deleteGatewayEntry(domain, gw, e.Ip)
		}
	}

	ips, err := net.LookupIP(domain)
	if err != nil {
		return fmt.Errorf("failed to resolve domain: " + domain)
	} else {
		log.Infof("resolved %s to %s", domain, ips)
		for _, ip := range ips {
			deleteGatewayEntry(domain, gw, ip.String())
		}
	}

	log.Infof("ensured route: %s->%s does not exist", domain, gateway)
	return nil
}

func DeleteAllEntries() error {
	routes, err := state.ReadAllRoutes()
	if err != nil {
		return fmt.Errorf("failed to load existing entries from state: %s", err)
	}
	log.Infof("deleting %d existing entries found in state", len(routes))
	for _, e := range routes {
		gw, err := getGatewayAddr(e.Gateway)
		if err != nil {
			return fmt.Errorf("could not parse gateway: %s", err)
		}
		deleteGatewayEntry(e.Domain, gw, e.Ip)
	}
	err = state.RemoveAllRoutes()
	if err != nil {
		return err
	}
	log.Infof("removed %v entries", len(routes))
	return nil
}

func deleteGatewayEntry(domain string, gateway net.IP, ip string) {
	err := deleteIfExists(domain, ip, gateway.String())
	if err != nil {
		log.Warnf("failed to delete route: %s->%s: %s", ip, gateway, err)
	}
}

func deleteIfExists(domain string, ip string, gateway string) error {
	exists, err := routeExists(ip, gateway)
	if err != nil {
		return err
	}
	if exists {
		return deleteRoute(domain, ip, gateway)
	} else {
		log.Debugf("route %s->%s does not exist", ip, gateway)
	}
	return nil
}

func deleteRoute(domain string, ip string, gateway string) error {
	log.Debugf("deleting route %v->%v", ip, gateway)
	cmd := exec.Command("route", "-n", "delete", "-net", ip+"/32", gateway)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete route %s->%s %s: %s", ip, gateway, cmd.Path, err)
	}
	log.Debugf("deleted route %v->%v", ip, gateway)

	err := state.RemoveRecordedRoute(domain, ip, gateway)
	if err != nil {
		return fmt.Errorf("unable to delete recorded route %s->%s from state file: %s", ip, gateway, err)
	}
	return nil
}
