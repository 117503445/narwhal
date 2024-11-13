package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"q/qrpc"
)

func (s *Server) OssCase0Start(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	log.Info().Msg("OssCase0Start")

	go func() {
		var client qrpc.WorkerSlave
		var err error
		for nodeID, clients := range s.clients {
			if nodeID == s.masterId {
				continue
			}
			client = clients[0]
		}

		_, err = client.OssCase0SendBatch(context.Background(), &qrpc.OssCase0Batch{})
		if err != nil {
			log.Error().Err(err).Msg("OssCase0SendBatch")
		}
		log.Info().Msg("Send OssCase0SendBatch")

	}()

	return &emptypb.Empty{}, nil
}

func (s *Server) OssCase0SendBatch(context.Context, *qrpc.OssCase0Batch) (*emptypb.Empty, error) {
	log.Info().Msg("Recv OssCase0SendBatch")

	return &emptypb.Empty{}, nil
}
