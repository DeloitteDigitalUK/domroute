package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

func ensureRoute(domain string, gateway net.IP) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to resolve domain: " + domain))
	}
	log.Printf("resolved %s to %s", domain, ips)

	for _, ip := range ips {
		exists, err := routeExists(ip.String(), gateway.String())
		if err != nil {
			log.Fatalln(err)
		}
		if exists {
			log.Printf("route %s->%s already exists", ip, gateway)
		} else {
			err = createRoute(domain, ip.String(), gateway.String())
			if err != nil {
				log.Printf("error creating route %s->%s: %s", ip, gateway, err)
				continue
			}
		}
	}

	deprecated, err := findDeprecated(domain, gateway, ips)
	log.Printf("found %d deprecated routes to remove", len(deprecated))
	for _, deprecated := range deprecated {
		err := deleteRouteIfExists(domain, deprecated, gateway.String())
		if err != nil {
			log.Printf("failed to delete route: %s->%s", deprecated, gateway)
			continue
		}
	}
}

func findDeprecated(domain string, gateway net.IP, ips []net.IP) ([]string, error) {
	routes, err := readRoutes()
	if err != nil {
		return nil, err
	}
	var deprecated []string
	for _, existing := range routes {
		if existing.Domain == domain && existing.Gateway == gateway.String() {
			found := false
			for _, ip := range ips {
				if existing.Ip == ip.String() {
					found = true
					break
				}
			}
			if !found {
				deprecated = append(deprecated, existing.Ip)
			}
		}
	}
	return deprecated, nil
}

func routeExists(ip string, gateway string) (bool, error) {
	cmd := exec.Command("netstat", "-rn")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to run %v: %s\n", cmd.Path, err)
	}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, ip) && strings.Contains(line, gateway) {
			return true, nil
		}
	}
	return false, nil
}

func createRoute(domain string, ip string, gateway string) error {
	log.Printf("creating route %v->%v", ip, gateway)
	cmd := exec.Command("route", "-n", "add", "-net", ip+"/32", gateway)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add route %s->%s %v: %s", ip, gateway, cmd.Path, err)
	}
	log.Printf("created route %v->%v", ip, gateway)

	err := recordRoute(domain, ip, gateway)
	if err != nil {
		return fmt.Errorf("unable to record route %s->%s in state file: %s", ip, gateway, err)
	}
	return nil
}

func deleteRouteIfExists(domain string, ip string, gateway string) error {
	exists, err := routeExists(ip, gateway)
	if err != nil {
		return err
	}
	if exists {
		return deleteRoute(domain, ip, gateway)
	} else {
		log.Printf("route %s->%s does not exist", ip, gateway)
	}
	return nil
}

func deleteRoute(domain string, ip string, gateway string) error {
	log.Printf("deleting route %v->%v", ip, gateway)
	cmd := exec.Command("route", "-n", "delete", "-net", ip+"/32", gateway)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete route %s->%s %s: %s", ip, gateway, cmd.Path, err)
	}
	log.Printf("deleted route %v->%v", ip, gateway)

	err := removeRecordedRoute(domain, ip, gateway)
	if err != nil {
		return fmt.Errorf("unable to delete recorded route %s->%s from state file: %s", ip, gateway, err)
	}
	return nil
}
