package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"strings"

	"github.com/ctrlsam/rigour/internal/messaging/kafka"
	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/internal/storage/mongodb"
	"github.com/ctrlsam/rigour/pkg/notifications/telegram"
	"github.com/ctrlsam/rigour/pkg/types"
)

// App wires Kafka consumer + Mongo repository + enricher.
type App struct {
	cfg Config

	consumer *kafka.TypedConsumer[types.Service]
	repo     storage.HostRepository
	enricher *Enricher

	mongoClient *mongodb.Client
}

func NewApp(ctx context.Context, cfg Config) (*App, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	consumer, err := kafka.NewTypedConsumer[types.Service](kafka.ConsumerConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.Topic,
		GroupID: cfg.KafkaGroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("persistence: kafka consumer: %w", err)
	}

	mongoClient, err := mongodb.NewClient(ctx, cfg.DbURI, cfg.DbTimeout)
	if err != nil {
		_ = consumer.Close()
		return nil, fmt.Errorf("persistence: mongodb client: %w", err)
	}

	repo, err := mongoClient.NewHostsRepository(ctx, storage.RepositoryConfig{
		URI:        cfg.DbURI,
		Database:   cfg.DbName,
		Collection: cfg.DbCollection,
		Timeout:    int(cfg.DbTimeout.Seconds()),
	})
	if err != nil {
		_ = consumer.Close()
		_ = mongoClient.Close(ctx)
		return nil, fmt.Errorf("persistence: hosts repository: %w", err)
	}

	// Open GeoIP databases from data directory
	geoipReaders, err := OpenGeoIPReaders(cfg.GeoIPDataDir)
	if err != nil {
		_ = consumer.Close()
		_ = mongoClient.Close(ctx)
		return nil, fmt.Errorf("persistence: geoip: %w", err)
	}

	enricher := NewEnricher(geoipReaders)

	return &App{
		cfg:         cfg,
		consumer:    consumer,
		repo:        repo,
		enricher:    enricher,
		mongoClient: mongoClient,
	}, nil
}

// Close closes underlying resources.
func (app *App) Close(ctx context.Context) error {
	var firstErr error
	if app == nil {
		return nil
	}
	if app.consumer != nil {
		if err := app.consumer.Close(); err != nil {
			firstErr = err
		}
	}
	if app.mongoClient != nil {
		if err := app.mongoClient.Close(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if app.enricher != nil {
		app.enricher.Close()
	}
	return firstErr
}

// Run blocks consuming messages until ctx is canceled.
func (app *App) Run(ctx context.Context) error {
	if app == nil {
		return errors.New("persistence: app is nil")
	}
	fmt.Println("persistence: started consuming messages...")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := app.consumer.Fetch(ctx)
		if err != nil {
			return err
		}
		if msg == nil {
			continue
		}

		if err := app.handleService(ctx, msg.Value); err != nil {
			return err
		}
	}
}

func (app *App) handleService(ctx context.Context, svc types.Service) error {
	fmt.Println("persistence: processing service:", svc.IP, svc.Port, svc.Protocol)

	now := time.Now()
	if svc.LastScan.IsZero() {
		svc.LastScan = now
	}

	// 1. Ensure host exists (first time only).
	if err := app.repo.EnsureHost(ctx, svc.IP, now); err != nil {
		return err
	}

	// 2. Enrich host with GeoIP/ASN data.
	host := &types.Host{IP: svc.IP, LastSeen: now}
	host, err := app.enricher.EnrichHost(ctx, host)
	if err != nil {
		return err
	}

	// 3. Update host with enrichment data (ASN, Location, Labels).
	if err := app.repo.UpdateHost(ctx, *host); err != nil {
		return err
	}

	// 4. Upsert service under the enriched host.
	result, err := app.repo.UpsertService(ctx, svc)
	if err != nil {
		return err
	}

	// 5. Notify if new or significant update
	if app.cfg.TelegramToken != "" && app.cfg.TelegramChatID != 0 {
		var msg string
		if result == storage.UpsertResultNewService {
			msg = fmt.Sprintf("🚀 *New Service Discovered*\n\n*IP:* `%s`\n*Port:* `%d`\n*Protocol:* `%s`\n*TLS:* `%v`\n*Transport:* `%s`",
				svc.IP, svc.Port, svc.Protocol, svc.TLS, svc.Transport)
		} else if result == storage.UpsertResultUpdatedService {
			// Do not notify for updated services to reduce noise
			// msg = fmt.Sprintf("🔄 *Service Updated*\n\n*IP:* `%s`\n*Port:* `%d`\n*Protocol:* `%s`\n*TLS:* `%v`\n*Transport:* `%s`",
			// 	svc.IP, svc.Port, svc.Protocol, svc.TLS, svc.Transport)
		}

		if msg != "" {
			bot := telegram.NewBot(app.cfg.TelegramToken, app.cfg.TelegramChatID)
			if svc.HTTP != nil {
				msg += fmt.Sprintf("\n*Status:* `%s`", svc.HTTP.Status)
			} else if svc.HTTPS != nil {
				msg += fmt.Sprintf("\n*Status:* `%s`", svc.HTTPS.Status)
			} else if svc.SSH != nil && svc.SSH.Banner != "" {
				banner := strings.TrimSpace(svc.SSH.Banner)
				if len(banner) > 100 {
					banner = banner[:100] + "..."
				}
				msg += fmt.Sprintf("\n*Banner:* `%s`", banner)
			}
			_ = bot.Notify(msg)
		}
	}

	return nil
}
