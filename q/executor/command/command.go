package command

import (
	"os"
	"q/executor/server"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

type ExecutorCmd struct {
}

func (e *ExecutorCmd) getGrpcPort() string {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50050"
	}
	return port
}

func (e *ExecutorCmd) Run() error {
	goutils.InitZeroLog(goutils.WithNoColor{})
	log.Info().Msg("Executor Run")
	
	port := e.getGrpcPort()
	server.NewServer().Run(port)

	return nil
}

