package server

import (
	"context"

	api "github.com/mfilipav/dwal/api/v1"
	"google.golang.org/grpc"
)

// Config carries Commit Log
type Config struct {
	CommitLog CommitLog
}

type grpcServer struct {
	*Config
}

// Register service (via config) to gRPC server
func newgrpcServer(config *Config) (srv *grpcServer, err error) {
	srv = &grpcServer{
		Config: config,
	}
	return srv, nil
}

// instantiates a gRPC server, register our service to that server,
func NewGRPCServer(config *Config) (*grpc.Server, error) {
	gsrv := grpc.NewServer()
	srv, err := newgrpcServer(config)
	if err != nil {
		return nil, err
	}
	api.RegisterLogServer(gsrv, &api.LogServer{
		Produce:       srv.Produce,
		Consume:       srv.Consume,
		ConsumeStream: srv.ConsumeStream,
		ProduceStream: srv.ProduceStream,
	})
	return gsrv, nil
}

// Appends Record from request to CommitLog, returns the new offset
func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (*api.ProduceResponse, error) {
	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &api.ProduceResponse{Offset: offset}, nil
}

// Reads at request-specified offset and returns a Record
func (s *grpcServer) Consume(ctx context.Context, req *api.ConsumeRequest) (*api.ConsumeResponse, error) {
	record, err := s.CommitLog.Read(req.Offset)
	if err != nil {
		return nil, err
	}
	return &api.ConsumeResponse{Record: record}, nil
}

// bidirectional streaming RPC, client can streams data into the server's log,
// and server can tell client whether each request succeeded
func (s *grpcServer) ProduceStream(stream api.Log_ProduceStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		res, err := s.Produce(stream.Context(), req)
		if err != nil {
			return err
		}
		if err = stream.Send(res); err != nil {
			return err
		}
	}
}

// server-side streaming RPC
func (s *grpcServer) ConsumeStream(req *api.ConsumeRequest, stream api.Log_ConsumeStreamServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			res, err := s.Consume(stream.Context(), req)
			switch err.(type) {
			case nil:
			case api.ErrOffsetOutOfRange:
				continue
			default:
				return err
			}
			if err = stream.Send(res); err != nil {
				return err
			}
			req.Offset++
		}
	}
}

// Interface for CommitLog, it should do 2 things: Append and Read
// Allows different CommitLog implementations (fast testing in memory vs
// production with disk writing)
type CommitLog interface {
	Append(*api.Record) (uint64, error)
	Read(uint64) (*api.Record, error)
}
