package command

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"q/common"
	"q/qrpc"
	"sync"
	"text/template"
	"time"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

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

type ECIMeta struct {
	ExpID    string
	NodeID   int
	WorkerID int
}

var expID string

func init() {
	expID = goutils.TimeStrSec()
}

func DeployECI() {
	httpProxy := os.Getenv("http_proxy")
	masterIp := os.Getenv("master_ip")

	NODE_COUNT := 4
	WORKER_COUNT := 1

	// mastersUrl := make([]string, 0)
	// for i := 0; i < NODE_COUNT; i++ {
	// 	mastersUrl = append(mastersUrl, fmt.Sprintf("http://%v:2312%d", masterIp, i))
	// }

	w := &qrpc.WorkersNetInfo{
		ExpId:      expID,
		Proxy:      httpProxy,
		// MastersUrl: mastersUrl,
	}
	var m sync.Mutex

	metas := make([]*ECIMeta, 0)

	for nodeID := 0; nodeID < NODE_COUNT; nodeID++ {
		for workerID := 0; workerID < WORKER_COUNT; workerID++ {
			metas = append(metas, &ECIMeta{
				ExpID:    expID,
				NodeID:   nodeID,
				WorkerID: workerID,
			})
		}
	}

	createContainer := func(meta *ECIMeta) {
		containerGroupName := fmt.Sprintf("biye-%d-%d-%s", meta.NodeID, meta.WorkerID, expID)
		result, err := common.EciClient.CreateContainerGroup(&eci20180808.CreateContainerGroupRequest{
			RegionId:           tea.String("cn-hangzhou"),
			ContainerGroupName: tea.String(containerGroupName),
			Container: []*eci20180808.CreateContainerGroupRequestContainer{
				{
					Name:  tea.String("worker"),
					Image: tea.String(fmt.Sprintf("registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%s", expID)),
					EnvironmentVar: []*eci20180808.CreateContainerGroupRequestContainerEnvironmentVar{
						{
							Key:   tea.String("SLAVE_ID"),
							Value: tea.String(fmt.Sprintf("%d", meta.WorkerID)),
						},
					},
				},
			},
			RestartPolicy:   tea.String("Never"),
			Cpu:             tea.Float32(2),
			Memory:          tea.Float32(2),
			SpotStrategy:    tea.String("SpotAsPriceGo"),
			AutoCreateEip:   tea.Bool(true),
			SecurityGroupId: tea.String("sg-bp1chrrv37a1jm22u1v8"),
			VSwitchId:       tea.String("vsw-bp1x16k8zehbf4rsicd0k"),
		})

		if err != nil {
			log.Fatal().Err(err).Msg("CreateContainerGroupRequest failed")
		}
		log.Info().Interface("result", result).Msg("CreateContainerGroupRequest success")

		var internetIp *string

		for {
			result, err := common.EciClient.DescribeContainerGroups(&eci20180808.DescribeContainerGroupsRequest{
				RegionId:           tea.String("cn-hangzhou"),
				ContainerGroupName: tea.String(containerGroupName),
			})
			if err != nil {
				log.Fatal().Err(err).Msg("DescribeContainerGroupsRequest failed")
			}
			if len(result.Body.ContainerGroups) > 0 {
				internetIp = result.Body.ContainerGroups[0].InternetIp
				intranetIp := result.Body.ContainerGroups[0].IntranetIp
				if result.Body.ContainerGroups[0].InternetIp != nil && result.Body.ContainerGroups[0].IntranetIp != nil {
					log.Info().Str("internetIp", *internetIp).Str("intranetIp", *intranetIp).Msg("DescribeContainerGroupsRequest success")
					m.Lock()
					w.Workers = append(w.Workers, &qrpc.WorkerNetInfo{
						Name:        fmt.Sprintf("biye-%d-%d", meta.NodeID, meta.WorkerID),
						InternetIp:  *internetIp,
						IntranetIp:  *intranetIp,
						NodeIndex:   int64(meta.NodeID),
						WorkerIndex: int64(meta.WorkerID),
					})
					m.Unlock()
					break
				}
			}

			time.Sleep(time.Second * 3)
			log.Info().Interface("meta", meta).Msg("wait for ip")
		}

	}

	var wg sync.WaitGroup
	for _, meta := range metas {
		wg.Add(1)
		go func(meta *ECIMeta) {
			defer wg.Done()
			createContainer(meta)
		}(meta)
	}

	wg.Wait()

	log.Info().Msg("all containers created")

	for _, worker := range w.Workers {
		wg.Add(1)
		go func(worker *qrpc.WorkerNetInfo) {

			defer wg.Done()
			client := qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", worker.InternetIp), &http.Client{})

			for {
				resp, err := client.PutWorkersNetInfo(context.TODO(), &qrpc.WorkersNetInfo{
					ExpId:   w.ExpId,
					Workers: w.Workers,
					Proxy:   w.Proxy,

					MasterUrl: fmt.Sprintf("http://%v:2412%d", masterIp, worker.NodeIndex),
					MasterId:  int64(worker.NodeIndex),
					SlaveId:   int64(worker.WorkerIndex),
				})
				if err != nil {
					log.Warn().Err(err).Msg("failed to call PutWorkersNetInfo")
					time.Sleep(time.Second * 3)
					continue
				}
				log.Info().Msgf("resp: %v", resp)
				break
			}
		}(worker)
	}
	wg.Wait()

	// write w to "../Docker/validators/eci.pb"
	wBytes, err := proto.Marshal(w)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal")
	}
	if err := os.WriteFile("../Docker/validators/eci.pb", wBytes, 0666); err != nil {
		log.Fatal().Err(err).Msg("failed to write file")
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

		goutils.Exec(fmt.Sprintf("docker build -t registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v .", expID), goutils.WithCwd("./assets/fc-worker"))

		goutils.Exec(fmt.Sprintf("docker push registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v", expID), goutils.WithCwd("./assets/fc-worker"))

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
