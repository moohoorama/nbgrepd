package env

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func GetDomainName() (string, error) {
	envHostName := os.Getenv("HOSTNAME")
	if len(envHostName) > 4 {
		return envHostName, nil
	}
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return hostname, err
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip, err := ipv4.MarshalText()
			if err != nil {
				fmt.Println("ipv4 error", hostname, err)
				return hostname, err
			}
			hosts, err := net.LookupAddr(string(ip))
			if err != nil && len(hosts) == 0 {
				return hostname, err
			}
			fqdn := hosts[0]
			return strings.TrimSuffix(fqdn, "."), nil
		}
	}
	return hostname, nil
}
