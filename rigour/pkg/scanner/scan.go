package scanner

import (
	"context"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/ctrlsam/rigour/pkg/discovery"
	"github.com/ctrlsam/rigour/pkg/discovery/naabu"
	"github.com/ctrlsam/rigour/pkg/fingerprint"
	"github.com/ctrlsam/rigour/pkg/fingerprint/plugins"
)

// ScanTargetWithDiscoveryStream runs discovery and fingerprinting and invokes onEvent
// as soon as a service is identified.
//
// Note: today this runs fingerprinting inline inside naabu's callback; if you later
// parallelize scanning, ensure onEvent is concurrency-safe.
func ScanTargetWithDiscoveryStream(
	ipRange string,
	cfg discovery.DiscoveryConfig,
	scanCfg fingerprint.FingerprintConfig,
	onEvent func(ScanEvent),
) error {
	ctx := context.Background()

	if strings.TrimSpace(ipRange) == "" {
		return fmt.Errorf("target is empty")
	}
	if onEvent == nil {
		return fmt.Errorf("onEvent callback is nil")
	}

	onOpen := func(r discovery.Result) {
		fmt.Println("Discovered open port:", r.Host, r.Port)
		addr := netip.AddrPortFrom(netip.MustParseAddr(r.Host), uint16(r.Port))
		t := plugins.Target{Address: addr}

		svc, err := scanCfg.ScanTarget(t)
		if err == nil && svc != nil {
			onEvent(ScanEvent{
				Timestamp: time.Now(),
				IP:        r.Host,
				Port:      r.Port,
				Protocol:  "tcp",
				TLS:       svc.TLS,
				Transport: "tcp",
				Metadata:  svc.Raw,
			})
		}
	}

	return naabu.Run(ctx, ipRange, discovery.DiscoveryConfig{
		ScanType: cfg.ScanType,
		Ports:    cfg.Ports,
		TopPorts: cfg.TopPorts,
		Retries:  cfg.Retries,
		Rate:     cfg.Rate,
	}, onOpen)
}
