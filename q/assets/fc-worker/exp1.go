package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"q/qrpc"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var Exp1BatchSize = 5000
var Exp1TxSize = 128

// Exp1BoradcastStart(context.Context, *ExpStartRequest) (*google_protobuf.Empty, error)

// Exp1BoradcastRecvBatch(context.Context, *ExpBatch) (*google_protobuf.Empty, error)

// Exp1P2PStart(context.Context, *ExpStartRequest) (*google_protobuf.Empty, error)

// Exp1P2PRecvBatch(context.Context, *ExpBatch) (*google_protobuf.Empty, error)

// Exp1GetMetrics(context.Context, *google_protobuf.Empty) (*Exp1Metrics, error)

func Exp1ProduceBatch(req *qrpc.ExpStartRequest) {
	go func() {
		// produce batch
		for {
			start := time.Now()
			id := goutils.UUID4()
			batchesChan <- id

			Exp1SetBatchCreated(id)

			// press: 每秒钟的预期 tps
			// 预期每个批次耗费的毫秒数
			msPerBatch := int(float64(Exp1BatchSize) * 1000 / float64(req.Press))

			remain := msPerBatch - int(time.Since(start).Milliseconds())

			log.Info().Int("remain", remain).Str("batchID", id).Msg("produce batch")

			if remain > 0 {
				time.Sleep(time.Duration(remain) * time.Millisecond)
			} else {
				log.Warn().Int("remain", remain).Msg("Exp1BoradcastStart: batch is too slow")
			}
		}
	}()
}

func (s *Server) Exp1BoradcastStart(ctx context.Context, req *qrpc.ExpStartRequest) (*emptypb.Empty, error) {
	log.Info().Msg("Exp1BoradcastStart")
	Exp1SetIsMaster(true)

	go func() {
		var err error
		Exp1ProduceBatch(req)

		const PROCESS_NUM = 1
		for i := 0; i < PROCESS_NUM; i++ {
			go func(pid int) {
				for batchID := range batchesChan {
					log.Info().Str("batchID", batchID).Int("pid", pid).Msg("sending batch")
					payload := make([]byte, Exp1TxSize*Exp1BatchSize)

					for _, otherClient := range s.otherClients {
						_, err = otherClient.Exp1BoradcastRecvBatch(context.Background(), &qrpc.ExpBatch{
							Id:      batchID,
							Payload: payload,
							TxNum:   int64(Exp1BatchSize),
						})

						if err != nil {
							log.Error().Err(err).Msg("Exp1BoradcastRecvBatch")
						}
						log.Info().Int("pid", pid).Str("batchID", batchID).Msg("sending batch to one client success")
					}
					latency := Exp1GetBatchLatency(batchID).Milliseconds()
					Exp1AddLatency(latency)

					Exp1AddBatchMeta(&qrpc.ExpBatchMeta{
						SubmittedAt: timestamppb.Now(),
						TxNum:       int64(Exp1BatchSize),
					})

					// log.Info().Msg("sending batch to one client success")
				}
			}(i)
		}

	}()

	return &emptypb.Empty{}, nil
}

func (s *Server) Exp1BoradcastRecvBatch(ctx context.Context, batch *qrpc.ExpBatch) (*emptypb.Empty, error) {
	log.Info().Msg("Recv Exp1BoradcastRecvBatch")

	sleepForLatencyMock()

	return &emptypb.Empty{}, nil
}

func (s *Server) Exp1P2PStart(ctx context.Context, req *qrpc.ExpStartRequest) (*emptypb.Empty, error) {
	Exp1SetIsMaster(true)
	log.Debug().Int("num", len(s.otherClients)).Msg("Exp1P2PStart")

	go func() {
		// var err error
		Exp1ProduceBatch(req)

		const PROCESS_NUM = 1
		for i := 0; i < PROCESS_NUM; i++ {
			go func(pid int) {
				for batchID := range batchesChan {
					log.Info().Str("batchID", batchID).Int("pid", pid).Msg("sending batch")
					payload := make([]byte, Exp1TxSize*Exp1BatchSize)
					exp1BatchStorageLock.Lock()
					exp1BatchStorageStatus[batchID] = "received"
					exp1BatchStorageLock.Unlock()

					randIndex := make([]int, len(s.otherClients))
					for i := range randIndex {
						randIndex[i] = i
					}
					rand.Shuffle(len(randIndex), func(i, j int) {
						randIndex[i], randIndex[j] = randIndex[j], randIndex[i]
					})
					log.Debug().Ints("randIndex", randIndex).Msg("randIndex")

					sendCount := 0
					skipCount := 0

					for count, i := range randIndex {
						if count == BFTQuorumSize(len(s.otherClients)+1)-1 {
							latency := Exp1GetBatchLatency(batchID).Milliseconds()
							Exp1AddLatency(latency)
						}

						resp, err := s.otherClients[i].Exp1P2PGetBatchStatus(context.Background(), &qrpc.Exp1P2PGetBatchStatusRequest{Id: batchID, Worker: fmt.Sprintf("%v", s.masterId)})
						if err != nil {
							log.Fatal().Err(err).Msg("Exp1P2PGetBatchStatus")
						}
						// "give-me" "other-giving" "done"
						if resp.Status == "done" || resp.Status == "other-giving" {
							log.Info().Int("pid", pid).Str("batchID", batchID).Str("status", resp.Status).Str("GivingWorker", resp.GivingWorker).Msg("skip")
							skipCount++
							continue
						} else if resp.Status == "give-me" {
							_, err = s.otherClients[i].Exp1P2PRecvBatch(context.Background(), &qrpc.ExpBatch{
								Id:      batchID,
								Payload: payload,
								TxNum:   int64(Exp1BatchSize),
							})
							if err != nil {
								log.Fatal().Err(err).Msg("Exp1P2PRecvBatch")
							}
							sendCount++
							log.Info().Int("pid", pid).Str("batchID", batchID).Msg("sending batch to one client success")
						} else {
							log.Fatal().Str("status", resp.Status).Msg("unknown status")
						}
					}

					Exp1AddBatchMeta(&qrpc.ExpBatchMeta{
						SubmittedAt: timestamppb.Now(),
						TxNum:       int64(Exp1BatchSize),
					})
					log.Info().Str("batchID", batchID).Int("pid", pid).Int("sendCount", sendCount).Int("skipCount", skipCount).Msg("sending batch done")

					// log.Info().Msg("sending batch to one client success")
				}
			}(i)
		}

	}()

	return &emptypb.Empty{}, nil
}

func (s *Server) Exp1P2PRecvBatch(ctx context.Context, batch *qrpc.ExpBatch) (*emptypb.Empty, error) {
	log.Info().Str("batchID", batch.Id).Msg("Recv Exp1P2PRecvBatch")

	exp1BatchStorageLock.Lock()
	defer exp1BatchStorageLock.Unlock()

	if status, ok := exp1BatchStorageStatus[batch.Id]; ok {
		if status == "receiving" {
			exp1BatchStorageStatus[batch.Id] = "received"
			go func() {
				randIndex := make([]int, len(s.otherClients))
				for i := range randIndex {
					randIndex[i] = i
				}
				rand.Shuffle(len(randIndex), func(i, j int) {
					randIndex[i], randIndex[j] = randIndex[j], randIndex[i]
				})
				for _, i := range randIndex {
					resp, err := s.otherClients[i].Exp1P2PGetBatchStatus(context.Background(), &qrpc.Exp1P2PGetBatchStatusRequest{Id: batch.Id, Worker: fmt.Sprintf("%v", s.masterId)})
					if err != nil {
						log.Fatal().Err(err).Msg("Exp1P2PGetBatchStatus")
					}
					// "give-me" "other-giving" "done"
					if resp.Status == "done" || resp.Status == "other-giving" {
						continue
					}
					if resp.Status == "give-me" {
						_, err = s.otherClients[i].Exp1P2PRecvBatch(context.Background(), batch)
						if err != nil {
							log.Fatal().Err(err).Msg("Exp1P2PRecvBatch")
						}
					}
				}
			}()
		}
	} else {
		log.Warn().Str("batchID", batch.Id).Msg("batch not exist")
	}

	sleepForLatencyMock()

	return &emptypb.Empty{}, nil
}

func (s *Server) Exp1P2PGetBatchStatus(ctx context.Context, req *qrpc.Exp1P2PGetBatchStatusRequest) (*qrpc.Exp1P2PGetBatchStatusResponse, error) {
	log.Info().Msg("Exp1P2PGetBatchStatus")

	exp1BatchStorageLock.Lock()
	defer exp1BatchStorageLock.Unlock()

	if _, ok := exp1BatchStorageStatus[req.Id]; ok {
		return &qrpc.Exp1P2PGetBatchStatusResponse{
			Status:       "other-giving",
			GivingWorker: exp1BatchStorageWorker[req.Id],
		}, nil
	} else {
		exp1BatchStorageStatus[req.Id] = "receiving"
		exp1BatchStorageWorker[req.Id] = req.Worker
		return &qrpc.Exp1P2PGetBatchStatusResponse{
			Status: "give-me",
		}, nil
	}
}

func (s *Server) Exp1GetMetrics(ctx context.Context, req *emptypb.Empty) (*qrpc.Exp1Metrics, error) {
	log.Info().Msg("Exp1GetMetrics")
	m := Exp1GetMetrics()
	return m, nil
}

func BFTQuorumSize(n int) int {
	b := n*2/3 + 1
	// log.Info().Int("n", n).Int("b", b).Msg("BFTQuorumSize")
	return b
}

var exp1Metrics = &qrpc.Exp1Metrics{}
var exp1MetricsLock sync.Mutex

var exp1BatchCreated map[string]time.Time = make(map[string]time.Time, 0)
var exp1BatchCreatedLock sync.Mutex

// batch id -> "receiving", "received"
var exp1BatchStorageStatus map[string]string = make(map[string]string, 0)
var exp1BatchStorageWorker map[string]string = make(map[string]string, 0)
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
