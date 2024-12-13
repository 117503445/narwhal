package main

import (
	"context"

	"q/qrpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// rpc Exp2Start(ExpStartRequest) returns (google.protobuf.Empty);
// rpc Exp2WorkerRecvBatch(ExpBatch) returns (google.protobuf.Empty); // worker 收到 node0 的 batch
// rpc Exp2NodeRecvBatch(ExpBatch) returns (google.protobuf.Empty); // nodes 收到 worker 的 batch
// rpc Exp2PrimaryConfirmBatch(BatchMeta) returns (google.protobuf.Empty); // node0 确认 batch

// node0 is primary
var Exp2NodeType = "" // primary, worker, node

var Exp2WorkerClient []qrpc.WorkerSlave = []qrpc.WorkerSlave{}
var Exp2NodeClient []qrpc.WorkerSlave = []qrpc.WorkerSlave{}

func (s *Server) Exp2Start(ctx context.Context, req *qrpc.Exp2StartRequest) (*emptypb.Empty, error) {
	log.Info().Msg("Exp2Start")

	go func() {

		if Exp2NodeType == "primary" {
			// start batch production
			Exp1ProduceBatch(int(req.Press))
			go func() {
				nextWorkerIndex := 0

				for batchID := range batchesChan {
					nextWorkerIndex = (nextWorkerIndex + 1) % len(Exp2WorkerClient)
					workerClient := Exp2WorkerClient[nextWorkerIndex]

					log.Info().Str("batchID", batchID).Int("WorkerIndex", int(nextWorkerIndex)).Msg("send batch to worker")
					payload := make([]byte, Exp1TxSize*Exp1BatchSize)

					_, err := workerClient.Exp2WorkerRecvBatch(context.Background(), &qrpc.ExpBatch{
						Id:      batchID,
						Payload: payload,
						TxNum:   int64(Exp1BatchSize),
					})
					if err != nil {
						log.Error().Err(err).Msg("Exp2WorkerRecvBatch failed")
					}
				}
			}()
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Server) Exp2PrimaryConfirmBatch(ctx context.Context, req *qrpc.ExpBatchMeta) (*emptypb.Empty, error) {
	if Exp2NodeType != "primary" {
		log.Fatal().Msg("Exp2PrimaryConfirmBatch: not primary")
	}

	log.Info().Str("batchID", req.Id).Msg("Exp2NodeRecvBatch")

	latency := Exp1GetBatchLatency(req.Id).Milliseconds()
	Exp1AddLatency(latency)

	Exp1AddBatchMeta(&qrpc.ExpBatchMeta{
		SubmittedAt: timestamppb.Now(),
		TxNum:       int64(Exp1BatchSize),
	})

	return &emptypb.Empty{}, nil
}

func (s *Server) Exp2WorkerRecvBatch(ctx context.Context, req *qrpc.ExpBatch) (*emptypb.Empty, error) {
	if Exp2NodeType != "worker" {
		log.Fatal().Msg("Exp2WorkerRecvBatch: not worker")
	}

	log.Info().Str("batchID", req.Id).Msg("Exp2WorkerRecvBatch")

	go func() {
		bftSize := BFTQuorumSize(len(Exp2NodeClient) + 1)
		log.Debug().Int("bftSize", bftSize).Int("Exp2NodeClient", len(Exp2NodeClient)).
			Msg("Exp2WorkerRecvBatch")

		for count, nodeClient := range Exp2NodeClient {
			log.Debug().Int("count", count).Msg("Exp2WorkerRecvBatch sending to node")
			if count+2 == bftSize {
				go func() {
					_, err := s.clients[0][0].Exp2PrimaryConfirmBatch(context.Background(), &qrpc.ExpBatchMeta{
						Id:          req.Id,
						SubmittedAt: timestamppb.Now(),
						TxNum:       int64(Exp1BatchSize),
					})
					if err != nil {
						log.Error().Err(err).Msg("Exp2PrimaryConfirmBatch failed")
					}
				}()
			}

			_, err := nodeClient.Exp2NodeRecvBatch(context.Background(), req)
			if err != nil {
				log.Error().Err(err).Int("count", count).Msg("Exp2NodeRecvBatch failed")
			}
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Server) Exp2NodeRecvBatch(ctx context.Context, req *qrpc.ExpBatch) (*emptypb.Empty, error) {
	if Exp2NodeType != "node" {
		log.Fatal().Msg("Exp2NodeRecvBatch: not node")
	}

	log.Info().Str("batchID", req.Id).Msg("Exp2NodeRecvBatch")

	// sleepForLatencyMock()

	return &emptypb.Empty{}, nil
}
