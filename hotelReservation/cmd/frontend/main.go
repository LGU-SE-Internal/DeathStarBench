package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/registry"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/frontend"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tracing"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tune"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	tune.Init()
	// Initialize temporary logger for startup
	tempLogger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Timestamp().Caller().Logger()
	log.Logger = tempLogger

	log.Info().Msg("Reading config...")
	jsonFile, err := os.Open("config.json")
	if err != nil {
		log.Error().Msgf("Got error while reading config: %v", err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)

	servPort, _ := strconv.Atoi(result["FrontendPort"])
	servIP := result["FrontendIP"]
	knativeDNS := result["KnativeDomainName"]

	var (
		jaegerAddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger address")
		consulAddr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	flag.Parse()
	// Initialize OpenTelemetry with logging support
	tempLogger.Info().Msgf("Initializing OpenTelemetry with logging [service name: %v | host: %v]...", "frontend", *jaegerAddr)
	tracer, logger, err := tracing.InitWithLogging("frontend", *jaegerAddr)
	if err != nil {
		tempLogger.Panic().Msgf("Got error while initializing OpenTelemetry: %v", err)
	}
	
	// Set the global logger to the one with OTLP export
	log.Logger = logger
	logger.Info().Msg("OpenTelemetry tracer and logger initialized")

	logger.Info().Msgf("Initializing consul agent [host: %v]...", *consulAddr)
	registry, err := registry.NewClient(*consulAddr)
	if err != nil {
		logger.Panic().Msgf("Got error while initializing consul agent: %v", err)
	}
	logger.Info().Msg("Consul agent initialized")

	srv := &frontend.Server{
		KnativeDns: knativeDNS,
		Registry:   registry,
		Tracer:     tracer,
		IpAddr:     servIP,
		ConsulAddr: *consulAddr,
		Port:       servPort,
	}

	logger.Info().Msg("Starting server...")
	logger.Fatal().Msg(srv.Run().Error())
}
