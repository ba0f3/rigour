package persistence

import (
	"errors"
	"time"

	"github.com/ctrlsam/rigour/internal"
)

type Config struct {
	KafkaBrokers string
	KafkaGroupID string
	Topic        string

	DbURI        string
	DbName       string
	DbCollection string
	DbTimeout    time.Duration

	GeoIPDataDir string
}

func (c Config) withDefaults() Config {
	out := c
	if out.Topic == "" {
		out.Topic = internal.KafkaTopicScannedServices
	}
	if out.DbName == "" {
		out.DbName = internal.DatabaseName
	}
	if out.DbCollection == "" {
		out.DbCollection = internal.HostsRepositoryName
	}
	if out.DbTimeout <= 0 {
		out.DbTimeout = 10 * time.Second
	}
	return out
}

func (c Config) Validate() error {
	c = c.withDefaults()
	if c.KafkaBrokers == "" {
		return errors.New("persistence: kafka brokers is required")
	}
	if c.KafkaGroupID == "" {
		return errors.New("persistence: kafka group id is required")
	}
	if c.DbURI == "" {
		return errors.New("persistence: db uri is required")
	}
	if c.GeoIPDataDir == "" {
		return errors.New("persistence: geoip data directory is required")
	}
	return nil
}
