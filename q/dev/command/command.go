package command

import (
	"os"
	"text/template"

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

	// goutils.Exec("go build .", goutils.WithCwd("./assets/fc-worker"), goutils.WithEnv(map[string]string{
	// 	"GOOS":        "linux",
	// 	"GOARCH":      "amd64",
	// 	"CGO_ENABLED": "0",
	// }))

	// goutils.Exec("docker compose exec -T --workdir /workspace/q/assets/fc-worker fc s deploy -y", goutils.WithCwd("../"))

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

	tmpl := `name: narwhal-nodes

services:
{{- range $i := (until 4) }}  # 直接在模板中定义 NodeCount 为 4
  primary_{{$i}}:
    image: 117503445/narwhal
    environment:
      - NODE_TYPE=primary
      - VALIDATOR_ID={{ $i }}
      - LOG_LEVEL=-vvv
      - CONSENSUS_DISABLED=--consensus-disabled
    expose:
      - "3000" # Port to listen on messages from other primary nodes
      - "3001" # Port to listen on messages from our worker nodes
    ports:
      - "800{{$i}}:8000" # expose the gRPC port to be publicly accessible
      - "810{{$i}}:8010" # expose the port listening to metrics endpoint for prometheus metrics
    volumes:
      - ./validators:/validators
      - ./validators/validator-{{$i}}/logs:/home/logs
      - ./logs:/logs
    init: true

  worker_{{$i}}:
    image: 117503445/narwhal
    depends_on:
      - primary_{{$i}}
    environment:
      - NODE_TYPE=worker
      - VALIDATOR_ID={{ $i }}
      - LOG_LEVEL=-vvv
      - WORKER_ID=0
    expose:
      - "4000" # Port to listen on messages from our primary node
      - "4001" # Port to listen for transactions from clients
      - "4002" # Port to listen on messages from other worker nodes
    ports:
      - "700{{$i}}:4001" # expose the port for the worker to accept transactions from clients
    volumes:
      - ./validators:/validators
      - ./validators/validator-{{$i}}/logs:/home/logs
      - ./logs:/logs
    init: true

  qexecutor_{{$i}}:
    image: 117503445/narwhal
    depends_on:
      - primary_{{$i}}
    environment:
      - NODE_TYPE=qexecutor
      - WORKER_ID={{ $i }}
    volumes:
      - ./validators:/validators
      - ./validators/validator-{{$i}}/logs:/home/logs
      - ./logs:/logs
    init: true
{{- end }}`

	// 注册一个自定义函数，用于生成 0 到 n-1 的数字
	funcMap := template.FuncMap{
		"until": func(n int) []int {
			result := make([]int, n)
			for i := 0; i < n; i++ {
				result[i] = i
			}
			return result
		},
	}

	t := template.Must(template.New("config").Funcs(funcMap).Parse(tmpl))
	// write to dc.yml

	f, err := os.Create("../Docker/dc.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create file")
	}
	defer f.Close()

	err = t.Execute(f, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to execute template")
	}

	return nil
}
