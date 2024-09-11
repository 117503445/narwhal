package main

import (
	"strings"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

func cmd(cwd string, command string) error {
	strs := strings.Split(command, " ")
	return goutils.CMD(cwd, strs[0], strs[1:]...)
}

func main() {
	goutils.InitZeroLog()
	log.Info().Msg("hello world")

	deleteOld := false // 删除旧的数据
	if deleteOld {
		if err := goutils.CMD("../Docker", "docker", "compose", "down"); err != nil {
			log.Error().Err(err).Msg("run cmd failed")
		}
	}

	if err := cmd("../", "docker exec narwhal-builder cargo build --target-dir docker-target --bin node --bin benchmark_client"); err != nil {
		log.Error().Err(err).Msg("run cmd failed")
	}

    if err := cmd("../", "docker build -t 117503445/narwhal ."); err != nil {
		log.Error().Err(err).Msg("run cmd failed")
	}

    if err := cmd("../Docker", "docker compose up -d"); err != nil {
		log.Error().Err(err).Msg("run cmd failed")
	}
}
