package command

import (
	"github.com/rs/zerolog/log"
)

type ExecutorCmd struct {
}

func (*ExecutorCmd) Run() error {
	log.Info().Msg("ExecutorCmd Run")

	return nil
}
