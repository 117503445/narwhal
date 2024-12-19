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
	// durs := make([]time.Duration, 0)

	durMsList := make([]int64, 0)

	c := 0

	for {
		c++
		for i := 0; i < press; i++ {
			m[goutils.UUID4()] = nil
		}
		log.Info().Int("len", len(m)).Msg("map")
		t := time.Now()
		uuid := goutils.UUID4()
		ok := false
		for i := 0; i < 10000; i++ {
			_, ok = m[uuid]
		}
		log.Info().Bool("ok", ok).Dur("dur", time.Since(t)).Discard().Msg("map")
		// durs = append(durs, time.Since(t))
		durMsList = append(durMsList, time.Since(t).Milliseconds())
		if c%10 == 0 {
			goutils.WriteJSON("map-durs.json", durMsList)
		}

		time.Sleep(time.Second)
	}

}
