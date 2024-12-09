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

type Oss2CaseCMD struct {
}

func (c *Oss2CaseCMD) Run() error {
	var err error
	goutils.Exec("docker compose up -d", goutils.WithCwd("../"))

	goutils.Exec("docker compose exec -T q-dev /workspace/q/script/build.sh", goutils.WithCwd("../"))

	goutils.Exec(fmt.Sprintf("docker build -t registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v .", expID), goutils.WithCwd("./assets/fc-worker"))

	goutils.Exec(fmt.Sprintf("docker push registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v", expID), goutils.WithCwd("./assets/fc-worker"))

	w,_ := DeployECI(4, 1, make(chan struct{}), nil)
	log.Info().Interface("w", w).Msg("DeployECI")

	clients := make([]qrpc.WorkerSlave, 0)
	time.Sleep(3 * time.Second)
	for _, w := range w.Workers {

		c1 := qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", w.InternetIp), &http.Client{})
		clients = append(clients, c1)
		_, err = c1.OssCase2Start(context.Background(), &qrpc.OssCase1StartRequest{
			Ak: os.Getenv("ak"),
			Sk: os.Getenv("sk"),
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to call OssCase2Start")
		}
		log.Info().Str("internetIp", w.InternetIp).Msg("OssCase2Start")
	}

	for {
		for _, c1 := range clients {
			log.Info().Msg("OssCase2GetMetrics")
			metrics, err := c1.OssCase2GetMetrics(context.Background(), &emptypb.Empty{})
			if err != nil {
				log.Error().Err(err).Msg("failed to call OssCase0GetMetrics")
				continue
			}
			log.Info().Str("metrics",
				protojson.Format(metrics)).Msg("OssCase2GetMetrics Done")
			// log.Info().Str("metrics",protojson.Format(metrics)).Msg("OssCase0GetMetrics")
			if len(metrics.Batches) < 1 {
				continue
			}

			// log.Info().Int64("UploadDurMs", metrics.UploadDurMs).Msg("UploadDurMs")
			OssCaseProcessBatches(metrics.Batches)
		}
		time.Sleep(time.Second * 10)
	}

}
