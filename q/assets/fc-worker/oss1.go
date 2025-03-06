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

func (s *Server) OssCase1Start(ctx context.Context, req *qrpc.OssCase1StartRequest) (*emptypb.Empty, error) {
	log.Info().Msg("OssCase0Start")

	go func() {
		ak := req.Ak
		sk := req.Sk

		cfg := oss.LoadDefaultConfig().
			WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(ak, sk)).WithRegion("cn-hangzhou").WithEndpoint("oss-cn-hangzhou-internal.aliyuncs.com")

		client := oss.NewClient(cfg)

		// 10MB
		payload := make([]byte, 1024*1024*10)

		start := time.Now()
		_, err := client.PutObject(context.TODO(), &oss.PutObjectRequest{
			Bucket: oss.Ptr("biye1024"),
			Key:    oss.Ptr("test-block"),
			Body:   bytes.NewReader(payload),
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to call PutObject")
			return
		}

		for i := 0; i < 100; i++ {
			go func (i int) {
				_, err := client.PutObject(context.TODO(), &oss.PutObjectRequest{
					Bucket: oss.Ptr("biye1024"),
					Key:    oss.Ptr(fmt.Sprintf("test-block-%d", i)),
					Body:   bytes.NewReader(payload),
				})
				if err != nil {
					log.Fatal().Err(err).Msg("failed to call PutObject")
					return
				}
				ossCase1metricsLock.Lock()
				ossCase1metrics.Uploads = append(ossCase1metrics.Uploads, &qrpc.OssCaseBatchMeta{
					ReceivedAt: timestamppb.Now(),
					Size:       1024 * 1024 * 10,
				})
				ossCase1metricsLock.Unlock()
				
			}(i)
		}

		dur := time.Since(start)
		log.Info().Dur("dur", dur).Msg("PutObject")

		ossCase1metricsLock.Lock()
		ossCase1metrics.UploadDurMs = dur.Milliseconds()
		ossCase1metricsLock.Unlock()

		for i := 0; i < 100; i++ {
			_, err := client.GetObject(context.TODO(), &oss.GetObjectRequest{
				Bucket: oss.Ptr("biye1024"),
				Key:    oss.Ptr("test-block-0"),
			})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to call GetObject")
				return
			}
			ossCase1metricsLock.Lock()
			ossCase1metrics.Downloads = append(ossCase1metrics.Downloads, &qrpc.OssCaseBatchMeta{
				ReceivedAt: timestamppb.Now(),
				Size:       1024 * 1024 * 10,
			})
			ossCase1metricsLock.Unlock()
		}

	}()

	return &emptypb.Empty{}, nil
}

var ossCase1metrics = &qrpc.OssCase1Metrics{}
var ossCase1metricsLock sync.Mutex

func (s *Server) OssCase1GetMetrics(context.Context, *emptypb.Empty) (*qrpc.OssCase1Metrics, error) {

	ossCase1metricsLock.Lock()
	defer ossCase1metricsLock.Unlock()

	return ossCase1metrics, nil
}
