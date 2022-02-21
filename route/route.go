package route

import (
	"fmt"
	"github.com/pcornish/domroute/state"
	log "github.com/sirupsen/logrus"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

// EnsureExists resolves the current IPs for domain and gateway,
// then updates the route table if required. As well as adding
// new routes, those previously added for this domain and gateway
// combination, that no longer match the resolved IPs are removed,
// from the route table.
func EnsureExists(domain string, gateway string) error {
	gw, err := getGatewayAddr(gateway)
	if err != nil {
		return fmt.Errorf("could not parse gateway: %s", err)
	} else if "" == gateway {
		return fmt.Errorf("no gateway provided")
	}

	ips, err := getDomainAddr(domain, err)
	if err != nil {
		return err
	}

	for _, ip := range ips {
		exists, err := routeExists(ip.String(), gw.String())
		if err != nil {
			return fmt.Errorf("failed to determine if route %v->%v exists: %s", ip, gw, err)
		}
		if exists {
			log.Debugf("route %s->%s already exists", ip, gateway)
		} else {
			err = createRoute(domain, ip.String(), gw.String())
			if err != nil {
				log.Debugf("error creating route %s->%s: %s", ip, gateway, err)
				continue
			}
		}
	}

	deprecated, err := findDeprecated(domain, gw, ips)
	log.Debugf("found %d deprecated routes to remove", len(deprecated))
	for _, deprecated := range deprecated {
		err := deleteIfExists(domain, deprecated, gw.String())
		if err != nil {
			log.Warnf("failed to delete route: %s->%s: %s", deprecated, gateway, err)
			continue
		}
	}

	log.Infof("ensured route: %s->%s exists", domain, gateway)
	return nil
}

func getGatewayAddr(ipOrIface string) (net.IP, error) {
	if matched, _ := regexp.MatchString(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`, ipOrIface); matched {
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
		log.Infof("resolved gateway interface %s to %s", ipOrIface, ip)
		return net.ParseIP(ip), nil
	}
	return nil, fmt.Errorf("failed to resolve gateway: %s", ipOrIface)
}

func getDomainAddr(domain string, err error) ([]net.IP, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve domain: " + domain)
	}
	log.Infof("resolved %s to %s", domain, ips)
	return ips, nil
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
	log.Debugf("creating route %v->%v", ip, gateway)
	cmd := exec.Command("route", "-n", "add", "-net", ip+"/32", gateway)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add route %s->%s %v: %s", ip, gateway, cmd.Path, err)
	}
	log.Infof("created route %v->%v", ip, gateway)

	err := state.RecordRoute(domain, ip, gateway)
	if err != nil {
		return fmt.Errorf("unable to record route %s->%s in state file: %s", ip, gateway, err)
	}
	return nil
}

func findDeprecated(domain string, gateway net.IP, ips []net.IP) ([]string, error) {
	routes, err := state.ReadRoutesForDomain(domain, gateway)
	if err != nil {
		return nil, err
	}
	var deprecated []string
	for _, existing := range routes {
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
	return deprecated, nil
}
