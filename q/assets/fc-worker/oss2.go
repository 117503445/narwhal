package main

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"q/qrpc"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var OssClient *oss.Client

func (s *Server) OssCase2Start(ctx context.Context, req *qrpc.OssCase1StartRequest) (*emptypb.Empty, error) {
	log.Info().Msg("OssCase2Start")

	go func() {
		ak := req.Ak
		sk := req.Sk

		cfg := oss.LoadDefaultConfig().
			WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(ak, sk)).WithRegion("cn-hangzhou").WithEndpoint("oss-cn-hangzhou-internal.aliyuncs.com")

		OssClient = oss.NewClient(cfg)

		time.Sleep(10 * time.Second)

		for i := 0; i < 100; i++ {
			log.Debug().Int("i", i).Msg("Sending Batch")
			// go func(i int) {
			// 10MB
			payload := make([]byte, 1024*1024*10)
			_, err := OssClient.PutObject(context.TODO(), &oss.PutObjectRequest{
				Bucket: oss.Ptr("biye1024"),
				Key:    oss.Ptr(fmt.Sprintf("test-block-%d-%d", s.masterId, i)),
				Body:   bytes.NewReader(payload),
			})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to call PutObject")
				return
			}

			blockID := fmt.Sprintf("test-block-%d-%d", s.masterId, i)
			log.Info().Str("blockID", blockID).Msg("Sending Batch")

			var wg sync.WaitGroup
			for master, cs := range s.clients {
				if master == s.masterId {
					continue
				}
				c := cs[0]
				wg.Add(1)
				go func(c qrpc.WorkerSlave) {
					defer wg.Done()

					_, err := c.OssCase2SendBatch(context.TODO(), &qrpc.OssCase2Batch{
						Url: fmt.Sprintf("test-block-%d-%d", s.masterId, i),
						Id:  blockID,
					})
					if err != nil {
						log.Fatal().Err(err).Msg("failed to call OssCase2SendBatch")
					}
					log.Info().Str("blockID", blockID).Int("master", master).Msg("Sent Batch")
				}(c)
			}
			wg.Wait()
			log.Info().Str("blockID", blockID).Msg("Sent Batch done")

			ossCase2metricsLock.Lock()
			log.Info().Msg("Add Batch")
			ossCase2metrics.Batches = append(ossCase2metrics.Batches, &qrpc.OssCaseBatchMeta{
				ReceivedAt: timestamppb.Now(),
				Size:       1024 * 1024 * 10,
			})
			ossCase2metricsLock.Unlock()

			// }(i)
		}

	}()

	return &emptypb.Empty{}, nil
}

var ossCase2metrics = &qrpc.OssCase2Metrics{}
var ossCase2metricsLock sync.Mutex

func (s *Server) OssCase2GetMetrics(context.Context, *emptypb.Empty) (*qrpc.OssCase2Metrics, error) {
	log.Info().Msg("OssCase2GetMetrics")
	ossCase2metricsLock.Lock()
	// log.Info().Msg("OssCase2GetMetrics Lock")
	defer ossCase2metricsLock.Unlock()

	return ossCase2metrics, nil
}

// OssCase2SendBatch(context.Context, *OssCase2Batch) (*google_protobuf.Empty, error)
func (s *Server) OssCase2SendBatch(_ context.Context, req *qrpc.OssCase2Batch) (*emptypb.Empty, error) {
	log.Info().Str("blockID", req.Id).Str("url", req.Url).Msg("OssCase2SendBatch")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	_, err := OssClient.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr("biye1024"),
		Key:    oss.Ptr(req.Url),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to call GetObject")
	}

	log.Info().Str("blockID", req.Id).Msg("OssCase2SendBatch Done")

	return &emptypb.Empty{}, nil
}
