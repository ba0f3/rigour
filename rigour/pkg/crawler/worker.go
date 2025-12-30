package crawler

import (
	"context"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/discovery"
	"github.com/ctrlsam/rigour/pkg/crawler/discovery/naabu"
	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint"
	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	"github.com/ctrlsam/rigour/pkg/types"
)

// ScanTargetWithDiscoveryStream runs discovery and fingerprinting and invokes onEvent
// as soon as a service is identified.
//
// Fingerprinting runs concurrently with a configurable worker pool to avoid blocking
// port discovery. The onEvent callback must be concurrency-safe.
func ScanTargetWithDiscoveryStream(
	ipRange string,
	cfg discovery.DiscoveryConfig,
	scanCfg fingerprint.FingerprintConfig,
	onEvent func(types.Service),
) error {
	ctx := context.Background()

	if strings.TrimSpace(ipRange) == "" {
		return fmt.Errorf("target is empty")
	}
	if onEvent == nil {
		return fmt.Errorf("onEvent callback is nil")
	}

	// Channel for discovered ports
	portQueue := make(chan discovery.Result, 100)

	// WaitGroup to track fingerprinting workers
	var wg sync.WaitGroup

	// Number of concurrent fingerprinting workers
	// Adjust based on your network/target constraints
	numWorkers := 20

	// Start fingerprinting workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range portQueue {
				addr := netip.AddrPortFrom(netip.MustParseAddr(r.Host), uint16(r.Port))
				t := plugins.Target{Address: addr}

				svc, err := scanCfg.FingerprintTarget(t)
				if err == nil && svc != nil {
					out := types.FromPluginService(svc, time.Now())
					onEvent(*out)
				}
			}
		}()
	}

	onOpen := func(r discovery.Result) {
		// discard ipv6 result - not sure why naabu returns these
		if strings.Contains(r.Host, ":") {
			return
		}

		fmt.Println("Discovered open port:", r.Host, r.Port)

		// Non-blocking send to worker pool
		select {
		case portQueue <- r:
		default:
			// Queue full, fingerprint inline to avoid dropping
			addr := netip.AddrPortFrom(netip.MustParseAddr(r.Host), uint16(r.Port))
			t := plugins.Target{Address: addr}
			svc, err := scanCfg.FingerprintTarget(t)
			if err == nil && svc != nil {
				out := types.FromPluginService(svc, time.Now())
				onEvent(*out)
			}
		}
	}

	// Run discovery
	err := naabu.Run(ctx, ipRange, discovery.DiscoveryConfig{
		ScanType: cfg.ScanType,
		Ports:    cfg.Ports,
		TopPorts: cfg.TopPorts,
		Retries:  cfg.Retries,
		Rate:     cfg.Rate,
	}, onOpen)

	// Close queue and wait for workers to finish
	close(portQueue)
	wg.Wait()

	return err
}
