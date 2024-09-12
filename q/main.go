package main

import (
	"strings"

	"github.com/117503445/goutils"
	"github.com/alecthomas/kong"
	"github.com/rs/zerolog/log"
)

func cmd(cwd string, command string) error {
	strs := strings.Split(command, " ")
	return goutils.CMD(cwd, strs[0], strs[1:]...)
}

type Context struct {
}

type DefaultCmd struct {
}

func (d *DefaultCmd) Run(ctx *Context) error {
	return new(BuildCmd).Run(ctx)
}

type BuildCmd struct {
}

func (b *BuildCmd) Run(ctx *Context) error {
	deleteOld := false // 删除旧的数据
	if deleteOld {
		if err := goutils.CMD("../Docker", "docker", "compose", "down"); err != nil {
			log.Fatal().Err(err).Msg("run cmd failed")
		}
	}

	// -T 避免 the input device is not a TTY
	if err := cmd("../", "docker compose exec -T builder cargo build --target-dir docker-target --bin node --bin benchmark_client"); err != nil {
		log.Fatal().Err(err).Msg("run cmd failed")
	}

	if err := cmd("../", "docker build -t 117503445/narwhal ."); err != nil {
		log.Fatal().Err(err).Msg("run cmd failed")
	}

	if err := cmd("../Docker", "docker compose up -d"); err != nil {
		log.Fatal().Err(err).Msg("run cmd failed")
	}

	// 等待服务启动
	// sleepDuration := 3 * time.Second
	// log.Info().Dur("sleepDuration", sleepDuration).Msg("sleep")
	// time.Sleep(sleepDuration)

	return nil
}

type ReqCMD struct {
}

func (r *ReqCMD) Run(ctx *Context) error {
	if err := cmd("../Docker", "docker compose exec -T worker_0 ./bin/benchmark_client --nodes http://localhost:4001 --rate 20 --size 10 http://localhost:4001"); err != nil {
		log.Fatal().Err(err).Msg("run cmd failed")
	}
	return nil
}

var cli struct {
	DefaultCmd DefaultCmd `cmd:"" hidden:"" default:"1"`
	Build      BuildCmd   `cmd:""`
	Req        ReqCMD     `cmd:""`
}

func main() {
	goutils.InitZeroLog()

	ctx := kong.Parse(&cli)

	err := ctx.Run(&Context{})
	ctx.FatalIfErrorf(err)
}
