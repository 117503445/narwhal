package command

import (
	"os"
	"q/common"
	"sync"
	"text/template"
	"time"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"

	eci20180808 "github.com/alibabacloud-go/eci-20180808/v3/client"
	"github.com/alibabacloud-go/tea/tea"
)

// UpdateTemplate 更新模板
func UpdateTemplate() {
	// 基于 ../Docker/docker-compose.yml.tmpl 生成 ../Docker/docker-compose.yml
	log.Debug().Msg("update template")

	var err error

	tmpl, err := goutils.ReadText("../Docker/docker-compose.yml.tmpl")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read file")
	}

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

func DeployECI() {
	result, err := common.EciClient.CreateContainerGroup(&eci20180808.CreateContainerGroupRequest{
		RegionId:           tea.String("cn-hangzhou"),
		ContainerGroupName: tea.String("biye-0-0"),
		Container: []*eci20180808.CreateContainerGroupRequestContainer{
			{
				Name:  tea.String("worker"),
				Image: tea.String("registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave"),
				EnvironmentVar: []*eci20180808.CreateContainerGroupRequestContainerEnvironmentVar{
					{
						Key:   tea.String("SLAVE_ID"),
						Value: tea.String("0"),
					},
				},
			},
		},
		RestartPolicy: tea.String("Never"),
		Cpu:           tea.Float32(2),
		Memory:        tea.Float32(2),
		SpotStrategy:  tea.String("SpotAsPriceGo"),
		AutoCreateEip: tea.Bool(true),
	})

	if err != nil {
		log.Fatal().Err(err).Msg("CreateContainerGroupRequest failed")
	}
	log.Info().Interface("result", result).Msg("CreateContainerGroupRequest success")

	for {
		result, err := common.EciClient.DescribeContainerGroups(&eci20180808.DescribeContainerGroupsRequest{
			RegionId:           tea.String("cn-hangzhou"),
			ContainerGroupName: tea.String("biye-0-0-" + goutils.TimeStrSec()),
		})
		if err != nil {
			log.Fatal().Err(err).Msg("DescribeContainerGroupsRequest failed")
		}

		internetIp := result.Body.ContainerGroups[0].InternetIp
		intranetIp := result.Body.ContainerGroups[0].IntranetIp
		if result.Body.ContainerGroups[0].InternetIp != nil && result.Body.ContainerGroups[0].IntranetIp != nil {
			log.Info().Str("internetIp", *internetIp).Str("intranetIp", *intranetIp).Msg("DescribeContainerGroupsRequest success")
			break
		}

		time.Sleep(time.Second * 3)
	}
}

type BuildCmd struct {
}

func (b *BuildCmd) Run() error {
	var wg sync.WaitGroup

	deleteOld := false // 删除旧的数据
	if deleteOld {
		goutils.Exec("docker compose down", goutils.WithCwd("../Docker"))
	}

	goutils.Exec("docker compose up -d", goutils.WithCwd("../"))

	wg.Add(1)
	go func() {
		defer wg.Done()
		// -T 避免 the input device is not a TTY
		goutils.Exec("docker compose exec -T builder cargo build --target-dir docker-target --bin node --bin benchmark_client", goutils.WithCwd("../"))

		goutils.Exec("docker compose exec -T q-dev /workspace/q/script/build.sh", goutils.WithCwd("../"))

		goutils.Exec("docker build -t registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave .", goutils.WithCwd("./assets/fc-worker"))

		goutils.Exec("docker push registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave", goutils.WithCwd("./assets/fc-worker"))

		DeployECI()

		// registry-vpc.cn-hangzhou.aliyuncs.com/117503445/biye-slave

		goutils.Exec("docker compose exec -T q-dev /workspace/q/script/proto.sh", goutils.WithCwd("../"))

		goutils.Exec("docker build -t 117503445/narwhal .", goutils.WithCwd("../"))

		UpdateTemplate()

		goutils.Exec("docker compose up -d", goutils.WithCwd("../"))
	}()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	DeployFC(&wg)
	// }()

	// goutils.Exec("docker compose exec -T --workdir /workspace/q/assets/fc-worker fc s deploy -y", goutils.WithCwd("../"))

	wg.Wait()

	goutils.Exec("docker compose up -d --build", goutils.WithCwd("../Docker"))
	time.Sleep(3 * time.Second)
	SendReq()

	return nil
}

type ReqCMD struct {
}

func (r *ReqCMD) Run() error {
	SendReq()

	return nil
}

type DeleteECICMD struct {
}

func (r *DeleteECICMD) Run() error {
	common.LoadENV()

	for {
		result, err :=
			common.EciClient.DescribeContainerGroups(&eci20180808.DescribeContainerGroupsRequest{
				RegionId: tea.String("cn-hangzhou"),
				Status:   tea.String("Failed"),
			})
		if err != nil {
			log.Fatal().Err(err).Msg("DescribeContainerGroupsRequest failed")
		}
		log.Info().Interface("result", result).Int("total", len(result.Body.ContainerGroups)).
			Msg("DescribeContainerGroupsRequest success")

		ids := make([]string, 0)
		// result.Body.ContainerGroups
		for _, containerGroup := range result.Body.ContainerGroups {
			log.Info().Interface("containerGroup", containerGroup.ContainerGroupId).Msg("containerGroup")
			ids = append(ids, *containerGroup.ContainerGroupId)
		}

		log.Info().Strs("ids", ids).Msg("ids")

		for _, id := range ids {
			_, err := common.EciClient.DeleteContainerGroup(&eci20180808.DeleteContainerGroupRequest{
				ContainerGroupId: tea.String(id),
				RegionId:         tea.String("cn-hangzhou"),
			})
			if err != nil {
				log.Fatal().Err(err).Msg("DeleteContainerGroupRequest failed")
			}
			log.Info().Str("id", id).Msg("DeleteContainerGroupRequest success")
		}

		time.Sleep(time.Second * 10)
	}
}

type Dev0CMD struct {
}

func (r *Dev0CMD) Run() error {
	log.Debug().Msg("dev-0")

	// ak := os.Getenv("ak")
	// if ak == "" {
	// 	log.Fatal().Msg("ak is empty")
	// }
	// sk := os.Getenv("sk")
	// config := &openapi.Config{
	// 	AccessKeyId:     tea.String(ak),
	// 	AccessKeySecret: tea.String(sk),
	// }
	// // Endpoint 请参考 https://api.aliyun.com/product/FC
	// config.Endpoint = tea.String("eci.cn-hangzhou.aliyuncs.com")
	// var err error
	// eciClient, err := eci20180808.NewClient(config)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("init eci client failed")
	// }

	// result, err := eciClient.CreateContainerGroup(&eci20180808.CreateContainerGroupRequest{})
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("create container group failed")
	// }
	// log.Info().Interface("result", result).Msg("create container group success")

	return nil
}
