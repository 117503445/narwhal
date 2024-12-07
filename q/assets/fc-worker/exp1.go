package main

import (
	"context"
	"sync"
	"time"

	"q/qrpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Exp1BoradcastStart(context.Context, *ExpStartRequest) (*google_protobuf.Empty, error)

// Exp1BoradcastRecvBatch(context.Context, *ExpBatch) (*google_protobuf.Empty, error)

// Exp1P2PStart(context.Context, *ExpStartRequest) (*google_protobuf.Empty, error)

// Exp1P2PRecvBatch(context.Context, *ExpBatch) (*google_protobuf.Empty, error)

// Exp1GetMetrics(context.Context, *google_protobuf.Empty) (*Exp1Metrics, error)

func (s *Server) Exp1BoradcastStart(ctx context.Context, req *qrpc.ExpStartRequest) (*emptypb.Empty, error) {
	log.Info().Msg("Exp1BoradcastStart")

	go func() {
		var client qrpc.WorkerSlave
		var err error
		for nodeID, clients := range s.clients {
			if nodeID == s.masterId {
				continue
			}
			client = clients[0]
		}

		for {
			start := time.Now()
			// 512KB
			payload := make([]byte, 512*1024)
			_, err = client.Exp1BoradcastRecvBatch(context.Background(), &qrpc.ExpBatch{
				Payload: payload,
				TxNum:   1000,
			})

			if err != nil {
				log.Error().Err(err).Msg("Exp1BoradcastRecvBatch")
			}
			log.Info().Msg("Send Exp1BoradcastRecvBatch")
			latency := time.Since(start).Milliseconds()
			AddExp1Latency(latency)
			AddExp1BatchMeta(&qrpc.ExpBatchMeta{
				SubmittedAt: timestamppb.Now(),
				TxNum:       1000,
			})
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Server) Exp1BoradcastRecvBatch(ctx context.Context, batch *qrpc.ExpBatch) (*emptypb.Empty, error) {
	log.Info().Msg("Recv Exp1BoradcastRecvBatch")

	return &emptypb.Empty{}, nil
}

func (s *Server) Exp1P2PStart(ctx context.Context, req *qrpc.ExpStartRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Server) Exp1P2PRecvBatch(ctx context.Context, batch *qrpc.ExpBatch) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Server) Exp1GetMetrics(ctx context.Context, req *emptypb.Empty) (*qrpc.Exp1Metrics, error) {
	log.Info().Msg("Exp1GetMetrics")
	m := GetExp1Metrics()
	return m, nil
}

var exp1Metrics = &qrpc.Exp1Metrics{}
var exp1MetricsLock sync.Mutex

func GetExp1Metrics() *qrpc.Exp1Metrics {
	exp1MetricsLock.Lock()
	defer exp1MetricsLock.Unlock()
	return exp1Metrics
}

func AddExp1BatchMeta(batch *qrpc.ExpBatchMeta) {
	exp1MetricsLock.Lock()
	defer exp1MetricsLock.Unlock()
	exp1Metrics.BatchMetas = append(exp1Metrics.BatchMetas, batch)
}

func AddExp1Latency(latency int64) {
	exp1MetricsLock.Lock()
	defer exp1MetricsLock.Unlock()
	exp1Metrics.LatenciesMS = append(exp1Metrics.LatenciesMS, latency)
}
