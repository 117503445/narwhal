package command

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"

	"q/qrpc"
)

func OssCaseProcessBatches(batches []*qrpc.OssCaseBatchMeta) {
	if len(batches) <= 1 {
		return
	}

	startTime := batches[0].ReceivedAt
	endTime := batches[len(batches)-1].ReceivedAt
	durSeconds := endTime.Seconds - startTime.Seconds
	if durSeconds == 0 {
		return
	}
	sizeSum := 0
	for _, batch := range batches {
		sizeSum += int(batch.Size)
	}
	bps := sizeSum / int(durSeconds)
	log.Info().Int64("dur", int64(durSeconds)).Int("sizeSum", sizeSum).Int("bps(MB)", bps/1024/1024).Msg("OssCase0GetMetrics")
}

type Oss0CaseCMD struct {
}

func (c *Oss0CaseCMD) Run() error {
	var err error
	goutils.Exec("docker compose up -d", goutils.WithCwd("../"))

	goutils.Exec("docker compose exec -T q-dev /workspace/q/script/build.sh", goutils.WithCwd("../"))

	goutils.Exec(fmt.Sprintf("docker build -t registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v .", expID), goutils.WithCwd("./assets/fc-worker"))

	goutils.Exec(fmt.Sprintf("docker push registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v", expID), goutils.WithCwd("./assets/fc-worker"))

	w := DeployECI(2, 1)
	log.Info().Interface("w", w).Msg("DeployECI")

	c1InternetIp := w.Workers[0].InternetIp
	c2InternetIp := w.Workers[1].InternetIp

	time.Sleep(3 * time.Second)

	c1 := qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", c1InternetIp), &http.Client{})
	_, err = c1.OssCase0Start(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to call OssCase0Start")
	}
	log.Info().Str("internetIp", c1InternetIp).Msg("OssCase0Start")

	c2 := qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", c2InternetIp), &http.Client{})
	for {
		metrics, err := c2.OssCase0GetMetrics(context.Background(), &emptypb.Empty{})
		if err != nil {
			log.Error().Err(err).Msg("failed to call OssCase0GetMetrics")
			time.Sleep(time.Second * 3)
			continue
		}
		log.Info().Str("metrics",
			protojson.Format(metrics)).Msg("OssCase0GetMetrics")
		// log.Info().Str("metrics",protojson.Format(metrics)).Msg("OssCase0GetMetrics")

		OssCaseProcessBatches(metrics.Batches)

		time.Sleep(time.Second * 10)

	}

}
