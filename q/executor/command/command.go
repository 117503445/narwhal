package command

import (
	"os"
	"q/executor/server"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

type ExecutorCmd struct {
}

func (*ExecutorCmd) Run() error {
	goutils.InitZeroLog(goutils.WithNoColor{})
	log.Info().Msg("Executor Run")

	executorsAddr:= os.Getenv("EXECUTORS_ADDR")
	log.Info().Str("executorsAddr", executorsAddr).Msg("") // http://qexecutor_0:50051,http://qexecutor_1:50051,http://qexecutor_2:50051,http://qexecutor_3:50051



	server.NewServer().Run(50051, executorsAddr)

	return nil
}
