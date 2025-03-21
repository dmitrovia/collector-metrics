package ip

import (
	"fmt"
	"net"
	"net/netip"
)

func GetLocalIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("GetLocalAddresse->GetIP: %w", err)
	}

	ladrrs := make([]string, 0)

	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)

		if !ok {
			continue
		}

		if ipnet.IP.IsLoopback() {
			continue
		}

		ladrrs = append(ladrrs, ipnet.IP.String())
	}

	return ladrrs, nil
}

func ContainsIPInSubnet(ipin string,
	subnet string,
) (bool, error) {
	network, err := netip.ParsePrefix(subnet)
	if err != nil {
		return false, fmt.Errorf("IpContainsInSubne->Pr: %w", err)
	}

	ipt, err := netip.ParseAddr(ipin)
	if err != nil {
		return false, fmt.Errorf("IpContainsInSubne->Ad: %w", err)
	}

	return network.Contains(ipt), nil
}
