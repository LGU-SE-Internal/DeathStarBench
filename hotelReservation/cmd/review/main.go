package main

import (
	"fmt"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/registry"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/review"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tracing"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tune"
	// "github.com/bradfitz/gomemcache/memcache"
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

	servPort, _ := strconv.Atoi(result["ReviewPort"])
	servIP := result["ReviewIP"]

	var (
		jaegerAddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger server addr")
		consulAddr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	flag.Parse()

	// Initialize OpenTelemetry with native logging
	tracer, logger, err := tracing.InitWithOtelLogging("review", *jaegerAddr)
	if err != nil {
		fmt.Printf("Failed to initialize OpenTelemetry: %v\n", err)
		os.Exit(1)
	}
	
	logger.Info().Msg("OpenTelemetry tracer and logger initialized")

	logger.Info().Msgf("Read database URL: %v", result["ReviewMongoAddress"])
	logger.Info().Msg("Initializing DB connection...")
	mongoSession, mongoClose := initializeDatabase(result["ReviewMongoAddress"])
	defer mongoClose()
	logger.Info().Msg("Successful")

	logger.Info().Msgf("Read review memcached address: %v", result["ReviewMemcAddress"])
	logger.Info().Msg("Initializing Memcached client...")
	memcClient := tune.NewMemCClient2(result["ReviewMemcAddress"])
	logger.Info().Msg("Successful")

	logger.Info().Msgf("Initializing consul agent [host: %v]...", *consulAddr)
	registry, err := registry.NewClient(*consulAddr)
	if err != nil {
		logger.Panic().Msgf("Got error while initializing consul agent: %v", err)
	}
	logger.Info().Msg("Consul agent initialized")

	srv := review.Server{
		Tracer: tracer,
		// Port:     *port,
		Registry:    registry,
		Port:        servPort,
		IpAddr:      servIP,
		MongoClient: mongoSession,
		MemcClient:  memcClient,
	}

	logger.Info().Msg("Starting server...")
	logger.Fatal().Msg(srv.Run().Error())
}
