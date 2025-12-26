package naabu

import (
	"context"
	"fmt"
	"strings"

	"github.com/projectdiscovery/goflags"
	naabuResult "github.com/projectdiscovery/naabu/v2/pkg/result"
	naabuRunner "github.com/projectdiscovery/naabu/v2/pkg/runner"
)

type Options struct {
	ScanType  string
	Ports     string
	TopPorts  string
	Interface string
	Retries   int
	Rate      int
	Stream    bool
}

type OSFingerprint struct {
	Target     string
	DeviceType string
	Running    string
	OSCPE      string
	OSDetails  string
}

type Result struct {
	Host          string
	Port          int
	Protocol      string
	Confidence    int
	OSFingerprint *OSFingerprint
	MacAddress    string
}

// Run executes Naabu discovery for a single input target and invokes onResult
// for each open port found.
func Run(ctx context.Context, ipRange string, opts Options, onResult func(Result)) error {
	if strings.TrimSpace(ipRange) == "" {
		return fmt.Errorf("naabu discovery input is empty")
	}

	naabuOpts := &naabuRunner.Options{
		Host: goflags.StringSlice{ipRange},
		// caller-configurable
		ScanType: opts.ScanType,
		Ports:    opts.Ports,
		TopPorts: opts.TopPorts,
		Rate:     opts.Rate,
		Retries:  opts.Retries,
		//Silent:            true,
	}

	naabuOpts.OnReceive = func(hr *naabuResult.HostResult) {
		for _, p := range hr.Ports {
			//fmt.Println("[DISCOVERY] Open port found:", hr.IP, p.Port)
			onResult(Result{
				Host:          hr.IP,
				Port:          p.Port,
				Protocol:      "tcp",
				Confidence:    int(hr.Confidence),
				OSFingerprint: (*OSFingerprint)(hr.OS),
				MacAddress:    hr.MacAddress,
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
