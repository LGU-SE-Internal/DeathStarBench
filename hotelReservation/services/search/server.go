package search

import (
	"fmt"
	"net"
	"time"

	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/dialer"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/registry"
	geo "github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/geo/proto"
	rate "github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/rate/proto"
	pb "github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/search/proto"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tls"
	"github.com/google/uuid"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const name = "srv-search"

// Server implments the search service
type Server struct {
	pb.UnimplementedSearchServer

	geoClient  geo.GeoClient
	rateClient rate.RateClient
	uuid       string

	Tracer     trace.Tracer
	Port       int
	IpAddr     string
	ConsulAddr string
	KnativeDns string
	Registry   *registry.Client
}

// Run starts the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	s.uuid = uuid.New().String()

	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Timeout: 120 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.UnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(),
		),
	}

	if tlsopt := tls.GetServerOpt(); tlsopt != nil {
		opts = append(opts, tlsopt)
	}

	srv := grpc.NewServer(opts...)
	pb.RegisterSearchServer(srv, s)

	// init grpc clients
	if err := s.initGeoClient("srv-geo"); err != nil {
		return err
	}
	if err := s.initRateClient("srv-rate"); err != nil {
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}

	err = s.Registry.Register(name, s.uuid, s.IpAddr, s.Port)
	if err != nil {
		return fmt.Errorf("failed register: %v", err)
	}
	log.Info().Msg("Successfully registered in consul")

	return srv.Serve(lis)
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
	s.Registry.Deregister(s.uuid)
}

func (s *Server) initGeoClient(name string) error {
	conn, err := s.getGprcConn(name)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.geoClient = geo.NewGeoClient(conn)
	return nil
}

func (s *Server) initRateClient(name string) error {
	conn, err := s.getGprcConn(name)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.rateClient = rate.NewRateClient(conn)
	return nil
}

func (s *Server) getGprcConn(name string) (*grpc.ClientConn, error) {
	if s.KnativeDns != "" {
		return dialer.Dial(
			fmt.Sprintf("consul://%s/%s.%s", s.ConsulAddr, name, s.KnativeDns),
			dialer.WithTracer(s.Tracer))
	} else {
		return dialer.Dial(
			fmt.Sprintf("consul://%s/%s", s.ConsulAddr, name),
			dialer.WithTracer(s.Tracer),
			dialer.WithBalancer(s.Registry.Client),
		)
	}
}

// Nearby returns ids of nearby hotels ordered by ranking algo
func (s *Server) Nearby(ctx context.Context, req *pb.NearbyRequest) (*pb.SearchResult, error) {
	// Get logger with trace context
	logger := zerolog.Ctx(ctx)
	if logger.GetLevel() == zerolog.Disabled {
		// If no logger in context, use global logger
		globalLogger := log.Logger
		logger = &globalLogger
	}
	
	// Extract trace information and add to logger
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		spanCtx := span.SpanContext()
		if spanCtx.HasTraceID() {
			newLogger := logger.With().Str("trace_id", spanCtx.TraceID().String()).Logger()
			logger = &newLogger
		}
		if spanCtx.HasSpanID() {
			newLogger := logger.With().Str("span_id", spanCtx.SpanID().String()).Logger()
			logger = &newLogger
		}
	}
	
	logger.Info().
		Float64("lat", float64(req.Lat)).
		Float64("lon", float64(req.Lon)).
		Str("in_date", req.InDate).
		Str("out_date", req.OutDate).
		Msg("Searching nearby hotels")
	
	// find nearby hotels
	logger.Debug().Msg("Querying geo service for nearby hotels")

	nearby, err := s.geoClient.Nearby(ctx, &geo.Request{
		Lat: req.Lat,
		Lon: req.Lon,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get nearby hotels from geo service")
		return nil, err
	}

	logger.Debug().
		Int("nearby_count", len(nearby.HotelIds)).
		Msg("Retrieved nearby hotels from geo service")

	for _, hid := range nearby.HotelIds {
		logger.Trace().Str("hotel_id", hid).Msg("Found nearby hotel")
	}

	// find rates for hotels
	logger.Debug().Msg("Querying rate service for hotel rates")
	rates, err := s.rateClient.GetRates(ctx, &rate.Request{
		HotelIds: nearby.HotelIds,
		InDate:   req.InDate,
		OutDate:  req.OutDate,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get rates from rate service")
		return nil, err
	}

	// TODO(hw): add simple ranking algo to order hotel ids:
	// * geo distance
	// * price (best discount?)
	// * reviews

	// build the response
	res := new(pb.SearchResult)
	for _, ratePlan := range rates.RatePlans {
		logger.Trace().
			Str("hotel_id", ratePlan.HotelId).
			Str("rate_code", ratePlan.Code).
			Msg("Adding hotel to search results")
		res.HotelIds = append(res.HotelIds, ratePlan.HotelId)
	}
	
	logger.Info().
		Int("results_count", len(res.HotelIds)).
		Msg("Search nearby completed")
	
	return res, nil
}
