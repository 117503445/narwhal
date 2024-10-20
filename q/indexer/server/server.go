package server

import (
	"context"
	"io"
	"net"
	"q/rpc"
	"q/indexer/node"
	"q/indexer/common"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Server struct {
	rpc.UnimplementedIndexerServer
	node      *node.Node
	recvChan  chan *common.ReqWithCh
	respCh    chan *rpc.QueryMsg
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

		resCh := make(chan *rpc.QueryMsg)
        s.recvChan <- &common.ReqWithCh{Req: req, ResCh: resCh}
		resp := <- resCh
		// todo

        if err := stream.Send(resp); err != nil {
            log.Error().Err(err).Msg("failed to send response")
            return err
        }
    }
}

func (s *Server) Send(_ context.Context, in *rpc.QueryMsg) (*rpc.QueryMsg, error) {
	resCh := make(chan *rpc.QueryMsg)
	// 打印in
	log.Info().Msgf("收到请求 : %v", in)
	msg := &common.ReqWithCh{Req: in, ResCh: resCh}
	s.recvChan <- msg
	resp := <- resCh
	return resp, nil
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

func NewServer(node *node.Node, recvChan chan *common.ReqWithCh, respCh chan *rpc.QueryMsg) *Server {
	return &Server{
		node: node,
		recvChan: recvChan,
		respCh: respCh,
	}
}
