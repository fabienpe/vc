package db

import (
	"context"
	"time"

	apiv1_status "vc/internal/gen/status/apiv1.status"
	"vc/pkg/logger"
	"vc/pkg/model"
	"vc/pkg/trace"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/codes"
)

// Service is the database service
type Service struct {
	dbClient   *mongo.Client
	cfg        *model.Cfg
	log        *logger.Log
	tp         *trace.Tracer
	probeStore *apiv1_status.StatusProbeStore

	VCDatastoreColl VCDatastoreColl
}

// New creates a new database service
func New(ctx context.Context, cfg *model.Cfg, tp *trace.Tracer, log *logger.Log) (*Service, error) {
	service := &Service{
		log:        log,
		cfg:        cfg,
		tp:         tp,
		probeStore: &apiv1_status.StatusProbeStore{},
	}

	if err := service.connect(ctx); err != nil {
		return nil, err
	}

	service.VCDatastoreColl = VCDatastoreColl{
		Service: service,
		Coll:    service.dbClient.Database("vc").Collection("datastore"),
	}
	if err := service.VCDatastoreColl.createIndex(ctx); err != nil {
		return nil, err
	}

	service.log.Info("Started")
	return service, nil
}

// Status returns the status of the database
func (s *Service) Status(ctx context.Context) *apiv1_status.StatusProbe {
	ctx, span := s.tp.Start(ctx, "db:status")
	defer span.End()

	if time.Now().Before(s.probeStore.NextCheck.AsTime()) {
		return s.probeStore.PreviousResult
	}
	probe := &apiv1_status.StatusProbe{
		Name:          "db",
		Healthy:       true,
		Message:       "OK",
		LastCheckedTS: timestamppb.Now(),
	}

	if err := s.dbClient.Ping(ctx, nil); err != nil {
		probe.Message = err.Error()
		probe.Healthy = false
	}

	s.probeStore.PreviousResult = probe
	s.probeStore.NextCheck = timestamppb.New(time.Now().Add(10 * time.Second))

	return probe
}

// connect connects to the database
func (s *Service) connect(ctx context.Context) error {
	ctx, span := s.tp.Start(ctx, "db:connect")
	defer span.End()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(s.cfg.Common.Mongo.URI))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	s.dbClient = client

	return nil
}

// Close closes db service
func (s *Service) Close(ctx context.Context) error {
	return s.dbClient.Disconnect(ctx)
}
