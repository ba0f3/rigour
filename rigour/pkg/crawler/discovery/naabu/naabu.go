package naabu

import (
	"context"
	"fmt"

	"github.com/ctrlsam/rigour/pkg/crawler/discovery"
	naabuResult "github.com/projectdiscovery/naabu/v2/pkg/result"
	naabuRunner "github.com/projectdiscovery/naabu/v2/pkg/runner"
)

// Run executes Naabu discovery for multiple input targets and invokes onResult
// for each open port found.
func Run(ctx context.Context, ipRanges []string, opts discovery.DiscoveryConfig, onResult func(discovery.Result)) error {
	if len(ipRanges) == 0 {
		return fmt.Errorf("naabu discovery input is empty")
	}

	naabuOpts := &naabuRunner.Options{
		Host: ipRanges,
		// caller-configurable
		ScanType: opts.ScanType,
		Ports:    opts.Ports,
		TopPorts: opts.TopPorts,
		Rate:     opts.Rate,
		Retries:  opts.Retries,
		Threads:  100,
		//Silent:   true,
	}

	naabuOpts.OnReceive = func(hr *naabuResult.HostResult) {
		for _, p := range hr.Ports {
			//fmt.Println("[DISCOVERY] Open port found:", hr.IP, p.Port)
			onResult(discovery.Result{
				Host:     hr.IP,
				Port:     p.Port,
				Protocol: p.Protocol.String(),
			})
		}
	}

	r, err := naabuRunner.NewRunner(naabuOpts)
	if err != nil {
		return fmt.Errorf("naabu.NewRunner failed: %w", err)
	}
	defer r.Close()

	// Naabu runner is not fully context-aware; honour ctx by stopping early if canceled.
	done := make(chan error, 1)
	go func() {
		done <- r.RunEnumeration(ctx)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("naabu enumeration failed: %w", err)
		}
		return nil
	}
}
