package command

import (
	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

type BuildCmd struct {
}

func (b *BuildCmd) Run() error {
	deleteOld := false // 删除旧的数据
	if deleteOld {
		goutils.Exec("docker compose down", goutils.WithCwd("../Docker"))
	}

	goutils.Exec("docker compose up -d", goutils.WithCwd("../"))

	// -T 避免 the input device is not a TTY
	goutils.Exec("docker compose exec -T builder cargo build --target-dir docker-target --bin node --bin benchmark_client", goutils.WithCwd("../"))

	goutils.Exec("docker compose exec -T q-dev /workspace/q/script/build.sh", goutils.WithCwd("../"))

	goutils.Exec("docker compose exec -T q-dev /workspace/q/script/proto.sh", goutils.WithCwd("../"))

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

func (r *ReqCMD) Run() error {
	goutils.Exec("docker compose exec -T worker_0 ./bin/benchmark_client --nodes http://localhost:4001 --rate 20 --size 10 http://localhost:4001", goutils.WithCwd("../Docker"))

	return nil
}

type Dev0CMD struct {
}

func (r *Dev0CMD) Run() error {
	log.Debug().Msg("dev-0")

	goutils.Exec("go build .", goutils.WithCwd("./assets/fc-worker"), goutils.WithEnv(map[string]string{
		"GOOS":        "linux",
		"GOARCH":      "amd64",
		"CGO_ENABLED": "0",
	}))

	goutils.Exec("docker compose exec -T --workdir /workspace/q/assets/fc-worker fc s deploy -y", goutils.WithCwd("../"))

	return nil
}
