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
	recvChan chan *rpc.MyTransaction
}

func (s *Server) PutExecuteInfo(_ context.Context, in *rpc.ExecuteInfo) (*emptypb.Empty, error) {
	log.Info().Int32("ConsensusRound", in.ConsensusRound).Int32("ExecuteHeight", in.ExecuteHeight).Uint64("id", in.Id).Msg("PutExecuteInfo")
	// s.recvChan <- in
	return &emptypb.Empty{}, nil
}

func (s *Server) SendTransaction(_ context.Context, in *rpc.MyTransaction) (*emptypb.Empty, error) {
	s.recvChan <- in
	return &emptypb.Empty{}, nil
}


func (s *Server) Run(port string) {
	port = ":" + port
    lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	grpcS := grpc.NewServer()
	rpc.RegisterExecutorServer(grpcS, s)

	if err := grpcS.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}

func NewServer(recvChan chan *rpc.MyTransaction) *Server {
	return &Server{
		recvChan: recvChan,
	}
}
