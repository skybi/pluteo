package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/skybi/data-server/internal/api"
	"github.com/skybi/data-server/internal/config"
	"os"
	"os/signal"
	"strings"
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
	if strings.ToLower(cfg.Environment) == "dev" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	log.Debug().Str("config", fmt.Sprintf("%+v", cfg)).Msg("")

	// Start up the portal & data APIs
	log.Info().Str("portal_api", cfg.PortalAPIAddress).Msg("starting up portal & data APIs...")
	apis := &api.Service{
		Config: cfg,
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

	// TODO: startup logic

	log.Info().Msg("done!")
	defer log.Info().Msg("shutting down...")

	// Wait for the application to be terminated
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
}
