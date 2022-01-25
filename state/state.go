package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
)

type route struct {
	Domain  string `json:"domain"`
	Ip      string `json:"ip"`
	Gateway string `json:"gateway"`
}

func readAllRoutes() ([]route, error) {
	stateFile, err := getStateFilePath()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("no existing routes in state")
			return []route{}, nil
		} else {
			return nil, fmt.Errorf("failed to open state file: %s: %s", stateFile, err)
		}
	}
	j, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %s: %s", stateFile, err)
	}
	var routes []route
	err = json.Unmarshal(j, &routes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall state file: %s: %s", stateFile, err)
	}
	log.Printf("loaded %d existing routes from state", len(routes))
	return routes, nil
}

func ReadRoutesForDomain(domain string, gateway net.IP) ([]route, error) {
	routes, err := readAllRoutes()
	if err != nil {
		return nil, err
	}
	var matched []route
	for _, r := range routes {
		if r.Domain == domain && r.Gateway == gateway.String() {
			matched = append(matched, r)
		}
	}
	return matched, nil
}

func RecordRoute(domain string, ip string, gateway string) error {
	entry := route{Domain: domain, Ip: ip, Gateway: gateway}
	routes, err := readAllRoutes()
	if err != nil {
		return err
	}
	routes = append(routes, entry)
	return writeStateFile(routes)
}

func RemoveRecordedRoute(domain string, ip string, gateway string) error {
	routes, err := readAllRoutes()
	if err != nil {
		return err
	}

	// write an empty JSON array instead of null if no entries
	r := []route{}
	for _, existing := range routes {
		if existing.Ip != ip || existing.Domain != domain || existing.Gateway != gateway {
			r = append(r, existing)
		}
	}

	return writeStateFile(r)
}

func getStateFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine user home: %s", err)
	}
	stateFile := filepath.Join(homeDir, ".domroute")
	return stateFile, nil
}

func writeStateFile(routes []route) error {
	j, err := json.Marshal(routes)
	if err != nil {
		return fmt.Errorf("failed to marshall routes: %v: %s", routes, err)
	}
	stateFile, err := getStateFilePath()
	if err != nil {
		return err
	}
	file, err := os.Create(stateFile)
	if err != nil {
		return fmt.Errorf("failed to open state file: %s: %s", stateFile, err)
	}
	_, err = file.Write(j)
	if err != nil {
		return fmt.Errorf("failed to write to state file: %s: %s", stateFile, err)
	}
	return nil
}
