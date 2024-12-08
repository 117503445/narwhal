package main

import (
	"context"
	"sync"
	"time"

	"q/qrpc"

	"github.com/117503445/goutils"
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
	Exp1SetIsMaster(true)

	go func() {
		var otherClients []qrpc.WorkerSlave
		var err error
		for nodeID, clients := range s.clients {
			if nodeID == s.masterId {
				continue
			}
			for _, c := range clients {
				otherClients = append(otherClients, c)
			}
		}

		go func() {
			// produce batch
			for {
				start := time.Now()
				id := goutils.UUID4()
				batchesChan <- id

				Exp1SetBatchCreated(id)

				// press: 每秒钟的预期 tps
				// 预期每个批次耗费的毫秒数
				msPerBatch := int(1000 * 1000 / req.Press)

				remain := msPerBatch - int(time.Since(start).Milliseconds())

				log.Info().Int("remain", remain).Str("batchID", id).Msg("produce batch")

				if remain > 0 {
					time.Sleep(time.Duration(remain) * time.Millisecond)
				} else {
					log.Warn().Int("remain", remain).Msg("Exp1BoradcastStart: batch is too slow")
				}
			}
		}()

		const PROCESS_NUM = 3
		for i := 0; i < PROCESS_NUM; i++ {
			go func(pid int) {
				for batchID := range batchesChan {
					log.Info().Str("batchID", batchID).Int("pid", pid).Msg("sending batch")
					// 512KB
					payload := make([]byte, 512*1024)

					for _, otherClient := range otherClients {
						_, err = otherClient.Exp1BoradcastRecvBatch(context.Background(), &qrpc.ExpBatch{
							Id:      batchID,
							Payload: payload,
							TxNum:   1000,
						})

						if err != nil {
							log.Error().Err(err).Msg("Exp1BoradcastRecvBatch")
						}
						log.Info().Msg("Send Exp1BoradcastRecvBatch")
					}
					latency := Exp1GetBatchLatency(batchID).Milliseconds()
					Exp1AddLatency(latency)

					Exp1AddBatchMeta(&qrpc.ExpBatchMeta{
						SubmittedAt: timestamppb.Now(),
						TxNum:       1000,
					})
				}
			}(i)
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

func (s *Server) Exp1P2PGetBatchStatus(ctx context.Context, req *qrpc.Exp1P2PGetBatchStatusRequest) (*qrpc.Exp1P2PGetBatchStatusResponse, error) {

	return &qrpc.Exp1P2PGetBatchStatusResponse{}, nil
}

func (s *Server) Exp1GetMetrics(ctx context.Context, req *emptypb.Empty) (*qrpc.Exp1Metrics, error) {
	log.Info().Msg("Exp1GetMetrics")
	m := Exp1GetMetrics()
	return m, nil
}

var exp1Metrics = &qrpc.Exp1Metrics{}
var exp1MetricsLock sync.Mutex

var exp1BatchCreated map[string]time.Time = make(map[string]time.Time, 0)
var exp1BatchCreatedLock sync.Mutex

// batch id -> "receiving", "received"
var exp1BatchStorageStatus map[string]string = make(map[string]string, 0)
var exp1BatchStorageLock sync.RWMutex

var isMaster bool
var isMasterLock sync.RWMutex

var batchesChan = make(chan string, 1000)

func Exp1GetMetrics() *qrpc.Exp1Metrics {
	exp1MetricsLock.Lock()
	defer exp1MetricsLock.Unlock()
	return exp1Metrics
}

func Exp1AddBatchMeta(batch *qrpc.ExpBatchMeta) {
	exp1MetricsLock.Lock()
	defer exp1MetricsLock.Unlock()
	exp1Metrics.BatchMetas = append(exp1Metrics.BatchMetas, batch)
}

func Exp1AddLatency(latency int64) {
	exp1MetricsLock.Lock()
	defer exp1MetricsLock.Unlock()
	exp1Metrics.LatenciesMS = append(exp1Metrics.LatenciesMS, latency)
}

func Exp1SetBatchCreated(id string) {
	exp1BatchCreatedLock.Lock()
	defer exp1BatchCreatedLock.Unlock()
	exp1BatchCreated[id] = time.Now()
}

func Exp1GetBatchLatency(id string) time.Duration {
	exp1BatchCreatedLock.Lock()
	defer exp1BatchCreatedLock.Unlock()
	start := exp1BatchCreated[id]
	return time.Since(start)
}

func Exp1GetIsMaster() bool {
	isMasterLock.RLock()
	defer isMasterLock.RUnlock()
	return isMaster
}

func Exp1SetIsMaster(b bool) {
	isMasterLock.Lock()
	defer isMasterLock.Unlock()
	isMaster = b
}