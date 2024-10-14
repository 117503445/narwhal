package command

import (
	"q/executor/server"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

type ExecutorCmd struct {
}

func (*ExecutorCmd) Run() error {
	goutils.InitZeroLog(goutils.WithNoColor{})
	log.Info().Msg("Executor Run")



	server.NewServer().Run()

	return nil
}
