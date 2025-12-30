package persistence

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

// GeoIPReaders holds the opened GeoIP database readers.
type GeoIPReaders struct {
	City *geoip2.Reader
	ASN  *geoip2.Reader
}

// OpenGeoIPReaders opens GeoIP database readers from the specified data directory.
func OpenGeoIPReaders(dataDir string) (*GeoIPReaders, error) {
	// Validate data directory path
	if dataDir == "" {
		return nil, errors.New("geoip: data directory path is required")
	}

	// Check if directory exists
	info, err := os.Stat(dataDir)
	if err != nil {
		return nil, fmt.Errorf("geoip: failed to access data directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("geoip: %s is not a directory", dataDir)
	}

	// Find database files
	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("geoip: failed to read data directory: %w", err)
	}

	var cityPath, asnPath string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := strings.ToLower(file.Name())

		// Check for city database
		if cityPath == "" && strings.Contains(name, "city") {
			cityPath = filepath.Join(dataDir, file.Name())
		}

		// Check for ASN database
		if asnPath == "" && strings.Contains(name, "asn") {
			asnPath = filepath.Join(dataDir, file.Name())
		}
	}

	// City and ASN databases are required
	if cityPath == "" {
		return nil, errors.New("geoip: city database not found in data directory")
	}
	if asnPath == "" {
		return nil, errors.New("geoip: asn database not found in data directory")
	}

	// Open the databases
	readers := &GeoIPReaders{}

	// Open city database
	cityReader, err := geoip2.Open(cityPath)
	if err != nil {
		return nil, fmt.Errorf("geoip: failed to open city database: %w", err)
	}
	readers.City = cityReader

	// Open ASN database
	asnReader, err := geoip2.Open(asnPath)
	if err != nil {
		readers.Close()
		return nil, fmt.Errorf("geoip: failed to open asn database: %w", err)
	}
	readers.ASN = asnReader

	return readers, nil
}

// Close closes all opened GeoIP database readers.
func (g *GeoIPReaders) Close() error {
	var firstErr error
	if g.City != nil {
		if err := g.City.Close(); err != nil {
			firstErr = err
		}
	}
	if g.ASN != nil {
		if err := g.ASN.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
