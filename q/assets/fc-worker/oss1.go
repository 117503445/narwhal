package main

import (
	"context"
	"sync"

	"q/qrpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

		// 10MB
		payload := make([]byte, 10*1024*1024)
		for {
			_, err = client.OssCase0SendBatch(context.Background(), &qrpc.OssCase0Batch{Payload: payload})

			if err != nil {
				log.Error().Err(err).Msg("OssCase0SendBatch")
			}
			log.Info().Msg("Send OssCase0SendBatch")
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Server) OssCase0SendBatch(ctx context.Context, batch *qrpc.OssCase0Batch) (*emptypb.Empty, error) {
	log.Info().Msg("Recv OssCase0SendBatch")
	metricsLock.Lock()
	defer metricsLock.Unlock()

	metrics.Batches = append(metrics.Batches, &qrpc.OssCase0BatchMeta{
		ReceivedAt: timestamppb.Now(),
		Size:       int64(len(batch.Payload)),
	})

	log.Info().Int("len", len(metrics.Batches)).Msg("OssCase0SendBatch")

	return &emptypb.Empty{}, nil
}

var metrics = &qrpc.OssCase0Metrics{}
var metricsLock sync.Mutex

func (s *Server) OssCase0GetMetrics(context.Context, *emptypb.Empty) (*qrpc.OssCase0Metrics, error) {
	log.Info().Msg("OssCase0GetMetrics")
	metricsLock.Lock()
	defer metricsLock.Unlock()

	log.Info().Int("len", len(metrics.Batches)).Msg("OssCase0GetMetrics")

	return metrics, nil
}
