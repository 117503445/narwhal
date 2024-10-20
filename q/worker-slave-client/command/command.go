package command

import (
	"context"
	"net/http"
	"q/qrpc"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

type WorkerSlaveClientCmd struct {
}

func (*WorkerSlaveClientCmd) Run() error {
	goutils.InitZeroLog(goutils.WithNoColor{})
	log.Info().Msg("WorkerSlaveClientCmd Run")

	var err error

	client := qrpc.NewWorkerSlaveProtobufClient("https://biye-niohlpafwu.cn-hangzhou.fcapp.run", &http.Client{})
	resp, err := client.Start(context.Background(), nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to call Start")
	}
	log.Info().Msgf("resp: %v", resp)

	return nil
}
