package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/registry"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/attractions"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tracing"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tune"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"time"
	// "github.com/bradfitz/gomemcache/memcache"
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

	tempLogger.Info().Msgf("Read database URL: %v", result["AttractionsMongoAddress"])
	tempLogger.Info().Msg("Initializing DB connection...")
	mongo_session, mongoClose := initializeDatabase(result["AttractionsMongoAddress"])
	defer mongoClose()
	tempLogger.Info().Msg("Successfull")

	// log.Info().Msgf("Read attractions memcashed address: %v", result["AttractionsMemcAddress"])
	// log.Info().Msg("Initializing Memcashed client...")
	// memc_client := memcache.New(result["AttractionsMemcAddress"])
	// memc_client.Timeout = time.Second * 2
	// memc_client.MaxIdleConns = 512
	// memc_client := tune.NewMemCClient2(result["AttractionsMemcAddress"])
	// log.Info().Msg("Successfull")

	serv_port, _ := strconv.Atoi(result["AttractionsPort"])
	serv_ip := result["AttractionsIP"]  // Will be empty, allowing auto-detection

	var (
		jaegeraddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger address")
		consuladdr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	flag.Parse()

	// Initialize OpenTelemetry with logging support
	tempLogger.Info().Msgf("Initializing OpenTelemetry with logging [service name: %v | host: %v]...", "attractions", *jaegeraddr)
	tracer, logger, err := tracing.InitWithLogging("attractions", *jaegeraddr)
	if err != nil {
		tempLogger.Panic().Msgf("Got error while initializing OpenTelemetry: %v", err)
	}
	
	// Set the global logger to the one with OTLP export
	log.Logger = logger
	logger.Info().Msg("OpenTelemetry tracer and logger initialized")

	logger.Info().Msgf("Initializing consul agent [host: %v]...", *consuladdr)
	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		logger.Panic().Msgf("Got error while initializing consul agent: %v", err)
	}
	logger.Info().Msg("Consul agent initialized")

	srv := attractions.Server{
		Tracer:      tracer,
		Registry:    registry,
		Port:        serv_port,
		IpAddr:      serv_ip,
		MongoClient: mongo_session,
	}

	logger.Info().Msg("Starting server...")
	logger.Fatal().Msg(srv.Run().Error())
}
