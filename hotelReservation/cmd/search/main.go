package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"time"

	"strconv"

	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/registry"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/search"
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

	tempLogger.Info().Msg("Reading config...")
	jsonFile, err := os.Open("config.json")
	if err != nil {
		tempLogger.Error().Msgf("Got error while reading config: %v", err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)

	servPort, _ := strconv.Atoi(result["SearchPort"])
	servIP := result["SearchIP"]
	knativeDNS := result["KnativeDomainName"]

	var (
		jaegerAddr = flag.String("jaegerAddr", result["jaegerAddress"], "Jaeger address")
		consulAddr = flag.String("consulAddr", result["consulAddress"], "Consul address")
	)
	flag.Parse()

	// Initialize OpenTelemetry with logging support
	tempLogger.Info().Msgf("Initializing OpenTelemetry with logging [service name: %v | host: %v]...", "search", *jaegerAddr)
	tracer, logger, err := tracing.InitWithLogging("search", *jaegerAddr)
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

	srv := &search.Server{
		Tracer:     tracer,
		Port:       servPort,
		IpAddr:     servIP,
		ConsulAddr: *consulAddr,
		KnativeDns: knativeDNS,
		Registry:   registry,
	}

	logger.Info().Msg("Starting server...")
	logger.Fatal().Msg(srv.Run().Error())
}
