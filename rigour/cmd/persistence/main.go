package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	internalconst "github.com/ctrlsam/rigour/internal"
	"github.com/ctrlsam/rigour/pkg/persistence"
	"github.com/spf13/cobra"
)

type cliConfig struct {
	brokers       string
	groupID       string
	topic         string
	dbURI         string
	dbName        string
	dbCollection  string
	geoipDataPath string
}

func main() {
	var cfg cliConfig

	root := &cobra.Command{
		Use:   "rigour-persistence",
		Short: "Consume crawler service events and persist/enrich hosts in MongoDB",
		RunE: func(cmd *cobra.Command, args []string) error {
			appCfg := persistence.Config{
				KafkaBrokers: cfg.brokers,
				KafkaGroupID: cfg.groupID,
				Topic:        cfg.topic,
				DbURI:        cfg.dbURI,
				DbName:       cfg.dbName,
				DbCollection: cfg.dbCollection,
				DbTimeout:    10 * time.Second,
				GeoIPDataDir: cfg.geoipDataPath,
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Shutdown on SIGINT/SIGTERM
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigCh
				cancel()
			}()

			app, err := persistence.NewApp(ctx, appCfg)
			if err != nil {
				return err
			}
			defer func() { _ = app.Close(context.Background()) }()

			err = app.Run(ctx)
			if err == context.Canceled {
				return nil
			}
			return err
		},
	}

	root.Flags().StringVar(&cfg.brokers, "brokers", "localhost:29092", "Kafka brokers (comma-separated)")
	root.Flags().StringVar(&cfg.groupID, "group", "rigour-persistence", "Kafka consumer group id")
	root.Flags().StringVar(&cfg.topic, "topic", internalconst.KafkaTopicScannedServices, "Kafka topic to consume")

	root.Flags().StringVar(&cfg.dbURI, "mongo-uri", "mongodb://localhost:27017", "MongoDB connection URI")
	root.Flags().StringVar(&cfg.dbName, "mongo-db", internalconst.DatabaseName, "MongoDB database name")
	root.Flags().StringVar(&cfg.dbCollection, "mongo-coll", internalconst.HostsRepositoryName, "MongoDB hosts collection name")

	root.Flags().StringVar(&cfg.geoipDataPath, "geoip-path", "", "Path to GeoIP data directory containing GeoLite2 database files")

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
