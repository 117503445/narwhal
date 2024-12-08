package command

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"

	"q/qrpc"
)

type Oss1CaseCMD struct {
}

func (c *Oss1CaseCMD) Run() error {
	var err error
	goutils.Exec("docker compose up -d", goutils.WithCwd("../"))

	goutils.Exec("docker compose exec -T q-dev /workspace/q/script/build.sh", goutils.WithCwd("../"))

	goutils.Exec(fmt.Sprintf("docker build -t registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v .", expID), goutils.WithCwd("./assets/fc-worker"))

	goutils.Exec(fmt.Sprintf("docker push registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v", expID), goutils.WithCwd("./assets/fc-worker"))

	w := DeployECI(1, 1, make(chan struct{}), nil)
	log.Info().Interface("w", w).Msg("DeployECI")

	c1InternetIp := w.Workers[0].InternetIp

	time.Sleep(3 * time.Second)

	c1 := qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", c1InternetIp), &http.Client{})
	_, err = c1.OssCase1Start(context.Background(), &qrpc.OssCase1StartRequest{
		Ak: os.Getenv("ak"),
		Sk: os.Getenv("sk"),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to call OssCase0Start")
	}
	log.Info().Str("internetIp", c1InternetIp).Msg("OssCase0Start")

	for {
		metrics, err := c1.OssCase1GetMetrics(context.Background(), &emptypb.Empty{})
		if err != nil {
			log.Error().Err(err).Msg("failed to call OssCase0GetMetrics")
			time.Sleep(time.Second * 3)
			continue
		}
		log.Info().Str("metrics",
			protojson.Format(metrics)).Msg("OssCase0GetMetrics")
		// log.Info().Str("metrics",protojson.Format(metrics)).Msg("OssCase0GetMetrics")
		if len(metrics.Downloads) < 1 {
			time.Sleep(time.Second * 3)
			continue
		}

		log.Info().Int64("UploadDurMs", metrics.UploadDurMs).Msg("UploadDurMs")
		OssCaseProcessBatches(metrics.Uploads)
		OssCaseProcessBatches(metrics.Downloads)

		time.Sleep(time.Second * 10)

	}

}
