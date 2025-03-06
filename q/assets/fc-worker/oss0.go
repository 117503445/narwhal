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
	ossCase0metricsLock.Lock()
	defer ossCase0metricsLock.Unlock()

	ossCase0metrics.Batches = append(ossCase0metrics.Batches, &qrpc.OssCaseBatchMeta{
		ReceivedAt: timestamppb.Now(),
		Size:       int64(len(batch.Payload)),
	})

	log.Info().Int("len", len(ossCase0metrics.Batches)).Msg("OssCase0SendBatch")

	return &emptypb.Empty{}, nil
}

var ossCase0metrics = &qrpc.OssCase0Metrics{}
var ossCase0metricsLock sync.Mutex

func (s *Server) OssCase0GetMetrics(context.Context, *emptypb.Empty) (*qrpc.OssCase0Metrics, error) {
	log.Info().Msg("OssCase0GetMetrics")
	ossCase0metricsLock.Lock()
	defer ossCase0metricsLock.Unlock()

	log.Info().Int("len", len(ossCase0metrics.Batches)).Msg("OssCase0GetMetrics")

	return ossCase0metrics, nil
}
