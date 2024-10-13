package command

import (
	"q/executor/server"

	"github.com/rs/zerolog/log"
)

type ExecutorCmd struct {
}

func (*ExecutorCmd) Run() error {
	log.Info().Msg("Executor Run")

	server.NewServer().Run(50051)

	return nil
}
