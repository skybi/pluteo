package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/skybi/pluteo/internal/api"
	"github.com/skybi/pluteo/internal/apikey/quota"
	"github.com/skybi/pluteo/internal/config"
	"github.com/skybi/pluteo/internal/storage/cache"
	"github.com/skybi/pluteo/internal/storage/postgres"
	"github.com/skybi/pluteo/internal/task"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Set up zerolog to use pretty printing
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})
	log.Info().Msg("starting up...")

	// Load the application configuration
	log.Info().Msg("loading configuration...")
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatal().Err(err).Msg("could not load the configuration")
	}
	if cfg.IsEnvProduction() {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Debug().Str("config", fmt.Sprintf("%+v", cfg)).Msg("")

	// Initialize the PostgreSQL storage driver
	log.Info().Msg("initializing database connection...")
	pgStorage := postgres.New(cfg.PostgresDSN)
	if err := pgStorage.Initialize(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("could not initialize the database connection")
	}
	defer pgStorage.Close()

	// Initialize the caching storage driver
	cacheStorage := cache.New(pgStorage)
	cacheStorage.Initialize(nil)
	defer cacheStorage.Close()

	// Create the API key quota tracker and schedule a task that flushes it
	quotaTracker := quota.NewTracker(cacheStorage.APIKeys())
	flushingTask := task.NewRepeating(func() {
		n, err := quotaTracker.Flush()
		if err != nil {
			log.Error().Err(err).Msg("could not flush changed API key quotas")
		} else {
			log.Debug().Int("amount", n).Msg("flushed changed API key quotas")
		}
	}, time.Minute)
	flushingTask.Start()
	defer flushingTask.Stop(true)

	// Start up the portal & data APIs
	log.Info().Str("portal_api", cfg.PortalAPIListenAddress).Str("data_api", cfg.DataAPIListenAddress).Msg("starting up portal & data APIs...")
	apis := &api.Service{
		Config:       cfg,
		Storage:      cacheStorage,
		QuotaTracker: quotaTracker,
	}
	apiErrs := make(chan error, 1)
	apis.Startup(apiErrs)
	go func() {
		err := <-apiErrs
		log.Fatal().Err(err).Msg("the API service raised an unexpected error")
	}()
	defer func() {
		log.Info().Msg("shutting down the portal & data APIs...")
		apis.Shutdown()
	}()

	log.Info().Msg("done!")
	defer log.Info().Msg("shutting down...")

	// Wait for the application to be terminated
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
}
