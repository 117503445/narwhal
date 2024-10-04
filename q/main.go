package main

import (
	"github.com/117503445/goutils"
	"github.com/alecthomas/kong"
	// "github.com/rs/zerolog/log"
)

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
		goutils.Exec("docker compose down", goutils.WithCwd("../Docker"))
	}

	goutils.Exec("docker compose up -d", goutils.WithCwd("../"))


	// -T 避免 the input device is not a TTY
	goutils.Exec("docker compose exec -T builder cargo build --target-dir docker-target --bin node --bin benchmark_client", goutils.WithCwd("../"))

	goutils.Exec("docker build -t 117503445/narwhal .", goutils.WithCwd("../"))

	goutils.Exec("docker compose up -d --build", goutils.WithCwd("../Docker"))

	// 等待服务启动
	// sleepDuration := 3 * time.Second
	// log.Info().Dur("sleepDuration", sleepDuration).Msg("sleep")
	// time.Sleep(sleepDuration)

	return nil
}

type ReqCMD struct {
}

func (r *ReqCMD) Run(ctx *Context) error {
	goutils.Exec("docker compose exec -T worker_0 ./bin/benchmark_client --nodes http://localhost:4001 --rate 20 --size 10 http://localhost:4001", goutils.WithCwd("../Docker"))

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
