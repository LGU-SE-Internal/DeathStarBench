package tune

import (
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tracing"
)

var (
	defaultGCPercent        int    = 100
	defaultMemCTimeout      int    = 2
	defaultMemCMaxIdleConns int    = 512
	defaultLogLevel         string = "trace"
	globalLogLevel          string = "trace" // Store the log level for later use
)

func setGCPercent() {
	ratio := defaultGCPercent
	if val, ok := os.LookupEnv("GC"); ok {
		ratio, _ = strconv.Atoi(val)
	}

	debug.SetGCPercent(ratio)
	tracing.Log.Info().Msgf("Tune: setGCPercent to %d", ratio)
}

func setLogLevel() {
	logLevel := defaultLogLevel
	if val, ok := os.LookupEnv("LOG_LEVEL"); ok {
		logLevel = val
	}
	
	// Store the log level for potential future use
	// Note: OpenTelemetry log provider doesn't use global log levels the same way zerolog did
	// The log level filtering is typically done at the collector or backend side
	globalLogLevel = logLevel
	
	tracing.Log.Info().Msgf("Set global log level: %s", logLevel)
}

func GetMemCTimeout() int {
	timeout := defaultMemCTimeout
	if val, ok := os.LookupEnv("MEMC_TIMEOUT"); ok {
		timeout, _ = strconv.Atoi(val)
	}
	tracing.Log.Info().Msgf("Tune: GetMemCTimeout %d", timeout)
	return timeout
}

// Hack of memcache.New to avoid 'no server error' during running
func NewMemCClient(server ...string) *memcache.Client {
	ss := new(memcache.ServerList)
	err := ss.SetServers(server...)
	if err != nil {
		// Hack: panic early to avoid pod restart during running
		panic(err)
		//return nil, err
	} else {
		memc_client := memcache.NewFromSelector(ss)
		memc_client.Timeout = time.Second * time.Duration(GetMemCTimeout())
		memc_client.MaxIdleConns = defaultMemCMaxIdleConns
		return memc_client
	}
}

func NewMemCClient2(servers string) *memcache.Client {
	ss := new(memcache.ServerList)
	server_list := strings.Split(servers, ",")
	err := ss.SetServers(server_list...)
	if err != nil {
		// Hack: panic early to avoid pod restart during running
		panic(err)
		//return nil, err
	} else {
		memc_client := memcache.NewFromSelector(ss)
		memc_client.Timeout = time.Second * time.Duration(GetMemCTimeout())
		memc_client.MaxIdleConns = defaultMemCMaxIdleConns
		return memc_client
	}
}

func Init() {
	setLogLevel()
	setGCPercent()
}
