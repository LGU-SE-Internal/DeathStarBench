package main

import (
	"fmt"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/registry"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/rate"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tracing"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tune"
	)

func main() {
	tune.Init()

	// Read config first to get jaeger address
	fmt.Println("Initializing OpenTelemetry with logging...")
	
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)

	logger.Info().Msg("Initializing DB connection...")
	mongoClient, mongoClose := initializeDatabase(result["RateMongoAddress"])
	defer mongoClose()

	logger.Info().Msgf("Read profile memcashed address: %v", result["RateMemcAddress"])
	logger.Info().Msg("Initializing Memcashed client...")
	memcClient := tune.NewMemCClient2(result["RateMemcAddress"])
	logger.Info().Msg("Success")

	servPort, _ := strconv.Atoi(result["RatePort"])
	servIP := result["RateIP"]

	var (
		jaegerAddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger address")
		consulAddr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	flag.Parse()

	// Initialize OpenTelemetry with logging support
	logger.Info().Msgf("Initializing OpenTelemetry with logging [service name: %v | host: %v]...", "rate", *jaegerAddr)
	tracer, logger, err := tracing.InitWithOtelLogging("rate", *jaegerAddr)
	if err != nil {
		fmt.Printf("Failed to initialize OpenTelemetry: %v
", err)
		os.Exit(1)
	}
	
	// Set the global logger to the one with OTLP export
	logger.Info().Msg("OpenTelemetry tracer and logger initialized")

	logger.Info().Msgf("Initializing consul agent [host: %v]...", *consulAddr)
	registry, err := registry.NewClient(*consulAddr)
	if err != nil {
		logger.Panic().Msgf("Got error while initializing consul agent: %v", err)
	}
	logger.Info().Msg("Consul agent initialized")

	srv := &rate.Server{
		Tracer:      tracer,
		Registry:    registry,
		Port:        servPort,
		IpAddr:      servIP,
		MongoClient: mongoClient,
		MemcClient:  memcClient,
	}

	logger.Info().Msg("Starting server...")
	if err := srv.Run(); err != nil {
		logger.Fatal().Msgf("Server error: %v", err)
	}
}
