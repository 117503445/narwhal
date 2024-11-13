package command

import (
	"context"
	"fmt"
	"net/http"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"

	"q/qrpc"
)

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

	internetIp := w.Workers[0].InternetIp

	c1 := qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", internetIp), &http.Client{})
	_, err = c1.OssCase0Start(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to call OssCase0Start")
	}
	log.Info().Str("internetIp", internetIp).Msg("OssCase0Start")

	return nil
}
