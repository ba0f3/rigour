package fingerprint

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/ctrlsam/rigour/pkg/fingerprint/plugins"
)

func ParseTarget(inputTarget string) (plugins.Target, error) {
	scanTarget := plugins.Target{}
	target := strings.Split(strings.TrimSpace(inputTarget), ":")
	if len(target) != 2 {
		return plugins.Target{}, fmt.Errorf("invalid target: %s", inputTarget)
	}

	hostStr, portStr := target[0], target[1]

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return plugins.Target{}, fmt.Errorf("invalid port specified")
	}

	ip := net.ParseIP(hostStr)
	var isHostname = false
	if ip == nil {
		var addrs []net.IP
		addrs, err = net.LookupIP(hostStr)
		if err != nil {
			return plugins.Target{}, err
		}
		isHostname = true
		ip = addrs[0]
	}

	// use IPv4 representation if possible
	ipv4 := ip.To4()
	if ipv4 != nil {
		ip = ipv4
	}

	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return plugins.Target{}, fmt.Errorf("invalid ip address specified %s", err)
	}
	targetAddr := netip.AddrPortFrom(addr, uint16(port))
	scanTarget.Address = targetAddr

	if isHostname {
		scanTarget.Host = hostStr
	}

	return scanTarget, nil
}
