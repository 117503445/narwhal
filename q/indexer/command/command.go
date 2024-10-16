package command

import (
	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

type IndexerCmd struct {
}

func (e *IndexerCmd) Run() error {
	goutils.InitZeroLog(goutils.WithNoColor{})
	log.Info().Msg("Indexer Run")

	return nil
}

