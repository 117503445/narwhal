package command

import (
	"q/worker/server"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

type WorkerCmd struct {
}

func (*WorkerCmd) Run() error {
	goutils.InitZeroLog(goutils.WithNoColor{})
	log.Info().Msg("Worker Run")

	server.NewServer().Run()

	return nil
}
