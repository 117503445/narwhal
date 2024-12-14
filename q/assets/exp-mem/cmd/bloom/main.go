package main

import (
	"github.com/117503445/goutils"
	"github.com/bits-and-blooms/bloom"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"time"
)

func main() {
	goutils.InitZeroLog()
	var err error

	filters := make([]*bloom.BloomFilter, 0)

	pressStr := os.Getenv("PRESS")
	press := 10000
	if pressStr != "" {
		press, err = strconv.Atoi(pressStr)
		if err != nil {
			log.Fatal().Err(err).Msg("strconv.Atoi")
		}
	}
	n := 1000000

	for {
		filter := bloom.NewWithEstimates(uint(n), 0.01)
		for i := 0; i < n; i++ {
			filter.Add([]byte(goutils.UUID4()))
		}

		filters = append(filters, filter)
		time.Sleep(time.Second * time.Duration(n) / time.Duration(press))
	}
}
