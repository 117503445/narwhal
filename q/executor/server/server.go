package server

import (
	"context"
	"net"
	"q/rpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	rpc.UnimplementedExecutorServer
}

func (s *Server) PutExecuteInfo(_ context.Context, in *rpc.ExecuteInfo) (*emptypb.Empty, error) {
	log.Info().Int32("ConsensusRound", in.ConsensusRound).Int32("ExecuteHeight", in.ExecuteHeight).Msg("PutExecuteInfo")
	return &emptypb.Empty{}, nil
}

func (s *Server) Run(port int) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	grpcS := grpc.NewServer()
	rpc.RegisterExecutorServer(grpcS, s)

	if err := grpcS.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}

func NewServer() *Server {
	return &Server{}
}
