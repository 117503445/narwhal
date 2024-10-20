package command

import (
	"fmt"
	"os"
	"sync"
	"text/template"
	"time"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

// UpdateTemplate 更新模板
func UpdateTemplate() {
	// 基于 ../Docker/docker-compose.yml.tmpl 生成 ../Docker/docker-compose.yml
	log.Debug().Msg("update template")

	tmplBytes, err := os.ReadFile("../Docker/docker-compose.yml.tmpl")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read file")
	}
	tmpl := string(tmplBytes)

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

	f, err := os.Create("../Docker/docker-compose.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create file")
	}
	defer f.Close()

	err = t.Execute(f, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to execute template")
	}
}

func SendReq() {
	// goutils.Exec("docker compose exec -T worker_0 ./bin/benchmark_client --nodes http://localhost:4001 --rate 20 --size 10 http://localhost:4001", goutils.WithCwd("../Docker"))

	goutils.Exec("docker compose exec -T worker_0 ./bin/q send-req", goutils.WithCwd("../Docker"))
}

func DeployFC() {
	var err error
	// 创建 "q/assets/fc-worker/data"
	// if err = os.MkdirAll("../q/assets/fc-worker/data", 0755); err != nil {
	// 	log.Fatal().Err(err).Msg("failed to create dir")
	// }


	for nodeIndex := 0; nodeIndex < 4; nodeIndex++ {
		i := 0

		// 复制 "q/assets/fc-worker/s.yaml" 到 "q/assets/fc-worker/data/nodeIndex_i/s.yaml"
		// src := "../q/assets/fc-worker/s.yaml"
		// dst := fmt.Sprintf("../q/assets/fc-worker/data/%d_%d/s.yaml", nodeIndex, i)

		// if err = goutils.CopyFile(src, dst); err != nil {
		// 	log.Fatal().Err(err).Msg("failed to copy file")
		// }

		// 创建 q/assets/fc-worker/data/nodeIndex_i
		if err = os.MkdirAll(fmt.Sprintf("../q/assets/fc-worker/data/%d_%d", nodeIndex, i), 0755); err != nil {
			log.Fatal().Err(err).Msg("failed to create dir")
		}


	}
}

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

	UpdateTemplate()

	goutils.Exec("docker compose up -d --build", goutils.WithCwd("../Docker"))

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		time.Sleep(3 * time.Second)
		SendReq()
		wg.Done()
	}()

	DeployFC()
	// goutils.Exec("docker compose exec -T --workdir /workspace/q/assets/fc-worker fc s deploy -y", goutils.WithCwd("../"))

	wg.Wait()

	return nil
}

type ReqCMD struct {
}

func (r *ReqCMD) Run() error {
	SendReq()

	return nil
}

type Dev0CMD struct {
}

func (r *Dev0CMD) Run() error {
	log.Debug().Msg("dev-0")

	return nil
}
