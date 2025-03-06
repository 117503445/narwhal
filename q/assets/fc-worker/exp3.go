package main

// rpc Exp3Start(ExpStartRequest) returns (google.protobuf.Empty);
// rpc Exp3RecvBatch(Exp3Batch) returns (google.protobuf.Empty);

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"q/qrpc"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// rpc Exp3Start(ExpStartRequest) returns (google.protobuf.Empty);
// rpc Exp3RecvBatch(Exp3Batch) returns (google.protobuf.Empty);
func (s *Server) Exp3Start(ctx context.Context, req *qrpc.ExpStartRequest) (*emptypb.Empty, error) {
	log.Info().Msg("Exp3Start")
	Exp1ProduceBatch(int(req.Press))

	go func() {
		batchIndex := 0
		var batchIndexLock sync.Mutex

		const PROCESS_NUM = 10
		for i := 0; i < PROCESS_NUM; i++ {
			go func(pid int) {
				for batchID := range batchesChan {
					batchIndexLock.Lock()
					batchIndex++
					currentBatchIndex := batchIndex
					batchIndexLock.Unlock()
					
					f := func(batchID string, batchIndex int) {
						log.Info().Str("batchID", batchID).Int("batchIndex", batchIndex).Msg("send batch to worker")
						payload := make([]byte, Exp1TxSize*Exp1BatchSize)

						ossName := fmt.Sprintf("test-block-%d-%d", s.masterId, batchIndex)

						_, err := OssClient.PutObject(context.TODO(), &oss.PutObjectRequest{
							Bucket: oss.Ptr("biye1024"),
							Key:    oss.Ptr(ossName),
							Body:   bytes.NewReader(payload),
						})
						if err != nil {
							log.Error().Err(err).Msg("failed to call PutObject")
							return
						}

						count := 0
						var countLock sync.Mutex

						var wg sync.WaitGroup
						for _, c := range s.otherClients {
							wg.Add(1)
							go func(c qrpc.WorkerSlave) {
								defer wg.Done()
								_, err := c.Exp3RecvBatch(context.Background(), &qrpc.Exp3Batch{
									Id:      batchID,
									TxNum:   int64(Exp1BatchSize),
									OssName: ossName,
								})
								if err != nil {
									log.Error().Err(err).Msg("Exp3RecvBatch failed")
								}

								countLock.Lock()
								count++
								if count+2 == BFTQuorumSize(len(s.otherClients)+1) {
									go func(id string) {
										log.Info().Str("batchID", id).Msg("Quorum nodes received")
										latency := Exp1GetBatchLatency(id).Milliseconds()
										Exp1AddLatency(latency)

										Exp1AddBatchMeta(&qrpc.ExpBatchMeta{
											SubmittedAt: timestamppb.Now(),
											TxNum:       int64(Exp1BatchSize),
										})

									}(batchID)
								}
								countLock.Unlock()

							}(c)
						}
						wg.Wait()

					}

					f(batchID, currentBatchIndex)
				}
			}(i)
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Server) Exp3RecvBatch(ctx context.Context, req *qrpc.Exp3Batch) (*emptypb.Empty, error) {
	log.Info().Msg("Exp3RecvBatch")

	_, err := OssClient.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr("biye1024"),
		Key:    oss.Ptr(req.OssName),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to call GetObject")
	}

	return &emptypb.Empty{}, nil
}
