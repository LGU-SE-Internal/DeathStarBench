package review

import (
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	// "io/ioutil"
	"net"
	// "os"
	// "sort"
	"time"
	//"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/registry"
	pb "github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/review/proto"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tls"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	// "strings"

	"github.com/bradfitz/gomemcache/memcache"
)

const name = "srv-review"

// Server implements the rate service
type Server struct {
	pb.UnimplementedReviewServer

	Tracer      trace.Tracer
	Port        int
	IpAddr      string
	MongoClient *mongo.Client
	Registry    *registry.Client
	MemcClient  *memcache.Client
	uuid        string
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

	pb.RegisterReviewServer(srv, s)

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

type ReviewHelper struct {
	ReviewId    string    `bson:"reviewId"`
	HotelId     string    `bson:"hotelId"`
	Name        string    `bson:"name"`
	Rating      float32   `bson:"rating"`
	Description string    `bson:"description"`
	Image       *pb.Image `bson:"images"`
}

type ImageHelper struct {
	Url     string `bson:"url"`
	Default bool   `bson:"default"`
}

func (s *Server) GetReviews(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	// Get logger with trace context
	logger := zerolog.Ctx(ctx)
	if logger.GetLevel() == zerolog.Disabled {
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
	
	logger.Info().Msgf("Getting hotel reviews: hotel_id=%s", req.HotelId)

	res := new(pb.Result)
	reviews := make([]*pb.ReviewComm, 0)

	hotelId := req.HotelId

	ctx, memSpan := s.Tracer.Start(ctx, "memcached_get_review")
	memSpan.SetAttributes(attribute.String("span.kind", "client"))
	item, err := s.MemcClient.Get(hotelId)
	memSpan.End()
	if err != nil && err != memcache.ErrCacheMiss {
		logger.Panic().Msgf("Memcached error while getting reviews: hotel_id=%s, error=%v", hotelId, err)
	} else {
		if err == memcache.ErrCacheMiss {
			logger.Debug().Msgf("Review cache miss, fetching from database: hotel_id=%s", hotelId)
			
			_, mongoSpan := s.Tracer.Start(ctx, "mongo_review")
			mongoSpan.SetAttributes(attribute.String("span.kind", "client"))

			//session := s.MongoSession.Copy()
			//defer session.Close()
			//c := session.DB("review-db").C("reviews")
			c := s.MongoClient.Database("review-db").Collection("reviews")

			curr, err := c.Find(context.TODO(), bson.M{"hotelId": hotelId})
			if err != nil {
				logger.Error().Msgf("Failed to get reviews from database: hotel_id=%s, error=%v", hotelId, err)
			}

			var reviewHelpers []ReviewHelper
			//err = c.Find(bson.M{"hotelId": hotelId}).All(&reviewHelpers)
			curr.All(context.TODO(), &reviewHelpers)
			if err != nil {
				logger.Error().Msgf("Failed to parse review data: hotel_id=%s, error=%v", hotelId, err)
			}

			for _, reviewHelper := range reviewHelpers {
				revComm := pb.ReviewComm{
					ReviewId:    reviewHelper.ReviewId,
					Name:        reviewHelper.Name,
					Rating:      reviewHelper.Rating,
					Description: reviewHelper.Description,
					Images:      reviewHelper.Image}
				reviews = append(reviews, &revComm)
			}

			reviewJson, err := json.Marshal(reviews)
			if err != nil {
				logger.Error().Msgf("Failed to marshal reviews: hotel_id=%s, error=%v", hotelId, err)
			}
			memcStr := string(reviewJson)

			s.MemcClient.Set(&memcache.Item{Key: hotelId, Value: []byte(memcStr)})
			
			logger.Debug().Msgf("Review cache populated: hotel_id=%s, reviews_count=%d", hotelId, len(reviews))
		} else {
			reviewsStr := string(item.Value)
			logger.Debug().Msgf("Review cache hit: hotel_id=%s, size=%d", hotelId, len(reviewsStr))
			if err := json.Unmarshal([]byte(reviewsStr), &reviews); err != nil {
				logger.Panic().Msgf("Failed to unmarshal reviews: hotel_id=%s, error=%v", hotelId, err)
			}
		}
	}

	//reviewsEmpty := make([]*pb.ReviewComm, 0)

	res.Reviews = reviews
	logger.Info().Msgf("Returning reviews: hotel_id=%s, reviews_count=%d", hotelId, len(reviews))
	return res, nil
}
