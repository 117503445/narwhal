package main

import (
	"exp-mem/internal/common"
	"os"
	"strconv"
	"time"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

func main() {
	goutils.InitZeroLog()
	var err error

	m := make(map[string]interface{})
	pressStr := os.Getenv("PRESS")
	press := common.DefaultPress
	if pressStr != "" {
		press, err = strconv.Atoi(pressStr)
		if err != nil {
			log.Fatal().Err(err).Msg("strconv.Atoi")
		}
	}

	for {
		for i := 0; i < press; i++ {
			m[goutils.UUID4()] = nil
		}
		log.Info().Int("len", len(m)).Msg("map")
		time.Sleep(time.Second)
	}

}
