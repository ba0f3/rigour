package scan

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

	onOpen := func(r naabu.Result) {
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

	return naabu.Run(ctx, ipRange, naabu.Options{
		ScanType: cfg.ScanType,
		Ports:    cfg.Ports,
		TopPorts: cfg.TopPorts,
		Retries:  cfg.Retries,
		Rate:     cfg.Rate,
	}, onOpen)
}

func ScanTargetWithDiscovery(ipRange string, cfg discovery.DiscoveryConfig, scanCfg fingerprint.FingerprintConfig) ([]plugins.Service, error) {
	ctx := context.Background()

	if strings.TrimSpace(ipRange) == "" {
		return nil, fmt.Errorf("target is empty")
	}

	results := make([]plugins.Service, 0)

	onOpen := func(r naabu.Result) {
		addr := netip.AddrPortFrom(netip.MustParseAddr(r.Host), uint16(r.Port))
		t := plugins.Target{Address: addr}

		svc, err := scanCfg.ScanTarget(t)
		if err == nil && svc != nil {
			results = append(results, *svc)
		}
	}

	err := naabu.Run(ctx, ipRange, naabu.Options{
		ScanType: cfg.ScanType,
		Ports:    cfg.Ports,
		TopPorts: cfg.TopPorts,
		Retries:  cfg.Retries,
		Rate:     cfg.Rate,
	}, onOpen)
	if err != nil {
		return nil, err
	}

	return results, nil
}
