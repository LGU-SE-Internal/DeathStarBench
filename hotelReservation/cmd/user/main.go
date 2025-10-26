package main

import (
	"fmt"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strconv"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/registry"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/user"
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

	log.Info().Msg("Initializing DB connection...")
	mongoClient, mongoClose := initializeDatabase(result["UserMongoAddress"])
	defer mongoClose()

	servPort, _ := strconv.Atoi(result["UserPort"])
	servIP := result["UserIP"]

	var (
		jaegerAddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger address")
		consulAddr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	flag.Parse()

	// Initialize OpenTelemetry with logging support
	logger.Info().Msgf("Initializing OpenTelemetry with logging [service name: %v | host: %v]...", "user", *jaegerAddr)
	tracer, logger, err := tracing.InitWithOtelLogging("user", *jaegerAddr)
	if err != nil {
		fmt.Printf("Failed to initialize OpenTelemetry: %v\n", err)
		os.Exit(1)
	}
	
		logger.Info().Msg("OpenTelemetry tracer and logger initialized")

	logger.Info().Msgf("Initializing consul agent [host: %v]...", *consulAddr)
	registry, err := registry.NewClient(*consulAddr)
	if err != nil {
		logger.Panic().Msgf("Got error while initializing consul agent: %v", err)
	}
	logger.Info().Msg("Consul agent initialized")

	srv := &user.Server{
		Port:        servPort,
		IpAddr:      servIP,
		Tracer:      tracer,
		Registry:    registry,
		MongoClient: mongoClient,
	}

	logger.Info().Msg("Starting server...")
	if err := srv.Run(); err != nil {
		logger.Fatal().Msgf("Server error: %v", err)
	}
}
