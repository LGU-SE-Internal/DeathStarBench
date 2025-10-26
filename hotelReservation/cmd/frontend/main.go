package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/registry"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/frontend"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tracing"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tune"
)

func main() {
	tune.Init()

	// Initialize OpenTelemetry with logging support
	fmt.Println("Initializing OpenTelemetry with logging...")
	
	// Read config first to get jaeger address
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
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

	// Initialize OpenTelemetry with native logging
	tracer, logger, err := tracing.InitWithOtelLogging("frontend", *jaegerAddr)
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

	srv := &frontend.Server{
		KnativeDns: knativeDNS,
		Registry:   registry,
		Tracer:     tracer,
		IpAddr:     servIP,
		ConsulAddr: *consulAddr,
		Port:       servPort,
	}

	logger.Info().Msg("Starting server...")
	if err := srv.Run(); err != nil {
		logger.Fatal().Msgf("Server error: %v", err)
	}
}
