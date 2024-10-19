package server

import (
	"io"
	"net"
	"q/rpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Server struct {
	rpc.UnimplementedIndexerServer
	recvChan chan *rpc.IndexerReq
}

func (s *Server) SendIndexReq(stream rpc.Indexer_SendIndexReqServer) error {
	for {
        req, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            log.Error().Err(err).Msg("failed to receive request")
            return err
        }

        s.recvChan <- req

		// todo
        resp := &rpc.IndexerResp{
            Id:   req.Id,
            Addr: "some-address", // 这里可以根据实际情况设置
        }

        if err := stream.Send(resp); err != nil {
            log.Error().Err(err).Msg("failed to send response")
            return err
        }
    }
}

func (s *Server) Run(port string) {
	port = ":" + port
    lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	grpcS := grpc.NewServer()
	rpc.RegisterIndexerServer(grpcS, s)

	log.Info().Msg("Server is running...")

	if err := grpcS.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}

func NewServer(recvChan chan *rpc.IndexerReq) *Server {
	return &Server{
		recvChan: recvChan,
	}
}
