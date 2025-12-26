package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/ctrlsam/rigour/pkg/discovery"
	"github.com/ctrlsam/rigour/pkg/fingerprint"
	"github.com/ctrlsam/rigour/pkg/scanner"
	"github.com/spf13/cobra"
)

type cliConfig struct {
	fastMode bool
	timeout  int
	useUDP   bool
	verbose  bool
	stream   bool

	// Discovery Naabu settings
	scanType string
	ports    string
	topPorts string
	retries  int
	rate     int
}

var (
	config  cliConfig
	rootCmd = &cobra.Command{
		Use: "rigour [flags]\nTARGET SPECIFICATION:\n\tRequires an ip address or CIDR range\n" +
			"EXAMPLES:\n\trigour 192.168.1.0/24\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			configErr := checkConfig(config)
			if configErr != nil {
				return configErr
			}

			cidrRange := args[0]

			// Quick estimate of number of IPs in the range.
			_, ipnet, _ := net.ParseCIDR(cidrRange)
			ones, bits := ipnet.Mask.Size()
			numIPs := 1 << (bits - ones)
			fmt.Printf("[+] Scanning %d IPs in range %s\n", numIPs, cidrRange)

			onEvent := func(ev scanner.ScanEvent) {
				b, err := json.MarshalIndent(ev, "", "  ")
				if err != nil {
					// Streaming should never abort the whole scan due to a single marshal failure.
					fmt.Fprintf(os.Stderr, "failed to marshal event: %v\n", err)
					return
				}
				// Print one pretty JSON object at a time.
				_, _ = os.Stdout.Write(append(b, '\n'))
			}

			err := scanner.ScanTargetWithDiscoveryStream(cidrRange, createDiscoveryConfig(config), createScanConfig(config), onEvent)
			if err != nil {
				return fmt.Errorf("Failed running discovery+scan stream (%w)", err)
			}
			return nil
		},
	}
)

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	rootCmd.PersistentFlags().BoolVarP(&config.fastMode, "fast", "f", false, "fast mode")
	rootCmd.PersistentFlags().
		BoolVarP(&config.useUDP, "udp", "U", false, "run UDP plugins")

	rootCmd.PersistentFlags().BoolVarP(&config.verbose, "verbose", "v", false, "verbose mode")
	rootCmd.PersistentFlags().BoolVar(&config.stream, "stream", true, "stream results as NDJSON as services are identified")
	rootCmd.PersistentFlags().
		IntVarP(&config.timeout, "timeout", "w", 2000, "timeout (milliseconds)")

	// Discovery flags (Naabu). These control how rigour discovers open ports.
	rootCmd.PersistentFlags().StringVar(&config.scanType, "scan-type", "c", "discovery scan type (naabu; e.g. c=connect)")
	rootCmd.PersistentFlags().StringVar(&config.ports, "ports", "", "ports list (e.g. 80,443). If set, overrides top ports")
	rootCmd.PersistentFlags().StringVar(&config.topPorts, "top-ports", "100", "top ports (e.g. 100, 1000, full)") // full
	rootCmd.PersistentFlags().IntVar(&config.retries, "retries", 3, "discovery retries")
	rootCmd.PersistentFlags().IntVar(&config.rate, "rate", 50_000, "discovery rate (packets per second)")
}

func checkConfig(config cliConfig) error {
	if config.useUDP && config.verbose {
		user, err := user.Current()
		if err != nil {
			return fmt.Errorf("Failed to retrieve current user (error: %w)", err)
		}
		if !((runtime.GOOS == "linux" || runtime.GOOS == "darwin") && user.Uid == "0") {
			fmt.Fprintln(os.Stderr, "Note: UDP Scan may require root privileges")
		}
	}

	return nil
}

func createScanConfig(config cliConfig) fingerprint.FingerprintConfig {
	return fingerprint.FingerprintConfig{
		DefaultTimeout: time.Duration(config.timeout) * time.Millisecond,
		FastMode:       config.fastMode,
		UDP:            config.useUDP,
		Verbose:        config.verbose,
	}
}

func createDiscoveryConfig(config cliConfig) discovery.DiscoveryConfig {
	return discovery.DiscoveryConfig{
		ScanType: config.scanType,
		Ports:    config.ports,
		TopPorts: config.topPorts,
		Retries:  config.retries,
		Rate:     config.rate,
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
