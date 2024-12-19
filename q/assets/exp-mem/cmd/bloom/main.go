package main

import (
	"exp-mem/internal/common"
	"os"
	"strconv"
	"time"

	"github.com/117503445/goutils"
	"github.com/bits-and-blooms/bloom"
	"github.com/rs/zerolog/log"
)

func main() {
	goutils.InitZeroLog()
	var err error

	filters := make([]*bloom.BloomFilter, 0)

	pressStr := os.Getenv("PRESS")
	press := common.DefaultPress
	if pressStr != "" {
		press, err = strconv.Atoi(pressStr)
		if err != nil {
			log.Fatal().Err(err).Msg("strconv.Atoi")
		}
	}
	n := 1000000
	// durs := make([]time.Duration, 0)
	durMsList := make([]int64, 0)
	for {
		filter := bloom.NewWithEstimates(uint(n), 0.01)
		for i := 0; i < n; i++ {
			filter.Add([]byte(goutils.UUID4()))
		}

		filters = append(filters, filter)
		time.Sleep(time.Second * time.Duration(n) / time.Duration(press))

		payload := []byte(goutils.UUID4())
		t := time.Now()
		for _,f := range filters {
			for i := 0; i < 10000; i++ {
				f.Test(payload)
			}
		}
		// durs = append(durs, time.Since(t))
		durMsList = append(durMsList, time.Since(t).Milliseconds())

		if len(durMsList) % 10 == 0 {
			goutils.WriteJSON("bloom-durs.json", durMsList)
		}
	}
}
