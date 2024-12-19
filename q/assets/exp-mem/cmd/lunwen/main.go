package main

import (
	"time"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

func main() {
	goutils.InitZeroLog()
	h1 := make(map[string]interface{})
	for i := 0; i < 50000; i++ {
		h1[goutils.UUID4()] = nil
	}
	// durs := make([]time.Duration, 0)
	durMsList := make([]int64, 0)

	c := 0

	for {
		c++
		t := time.Now()
		uuid := goutils.UUID4()
		ok := false
		for i := 0; i < 10000; i++ {
			_, ok = h1[uuid]
		}
		log.Info().Bool("ok", ok).Dur("dur", time.Since(t)).Discard().Msg("map")
		// durs = append(durs, time.Since(t))
		durMsList = append(durMsList, time.Since(t).Milliseconds())

		if c%10 == 0 {
			goutils.WriteJSON("lunwen-durs.json", durMsList)
		}
		time.Sleep(time.Second)
	}

	// sleep forever
	// select {}

}
