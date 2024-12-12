package command

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	// "strings"
	"sync"
	"text/template"
	"time"

	"github.com/117503445/goutils"
	eci20180808 "github.com/alibabacloud-go/eci-20180808/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	"q/common"
	"q/qrpc"
)

const proxyContainerGroupName = "biye-proxy"

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

func RefreshExpID() {
	expID = goutils.TimeStrSec()
}

type ECIParam struct {
	Bandwidth float64 // 带宽限制，单位 MB
}

func ECIDelete(w *qrpc.WorkersNetInfo) {
	log.Info().Msg("ECIDelete")
	for _, worker := range w.Workers {
		_, err := common.EciClient.DeleteContainerGroup(&eci20180808.DeleteContainerGroupRequest{
			ContainerGroupId: tea.String(worker.EciId),
			RegionId:         tea.String("cn-hangzhou"),
		})
		if err != nil {
			log.Warn().Err(err).Msg("DeleteContainerGroupRequest failed")
		}
		// log.Info().Str("id", worker.EciId).Msg("DeleteContainerGroupRequest success")
	}
}

func ECIDeploy(
	nodeCount int, workerCount int, stop chan struct{}, param *ECIParam,
) (*qrpc.WorkersNetInfo, *http.Client) {
	if param == nil {
		param = &ECIParam{
			Bandwidth: 125,
		}
	}
	// 以 Byte per second 为单位
	bandwidth := int64(param.Bandwidth * 1024 * 1024)

	httpProxy := os.Getenv("http_proxy")
	masterIp := os.Getenv("master_ip")

	NODE_COUNT := 4
	if nodeCount > 0 {
		NODE_COUNT = nodeCount
	}
	WORKER_COUNT := 1
	if workerCount > 0 {
		WORKER_COUNT = workerCount
	}

	// mastersUrl := make([]string, 0)
	// for i := 0; i < NODE_COUNT; i++ {
	// 	mastersUrl = append(mastersUrl, fmt.Sprintf("http://%v:2312%d", masterIp, i))
	// }

	w := &qrpc.WorkersNetInfo{
		ExpId: expID,
		Proxy: httpProxy,
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

	dirLogs := "./logs"
	if err := os.MkdirAll(dirLogs, 0777); err != nil {
		log.Fatal().Err(err).Msg("failed to create dir")
	}

	// create proxy container client

	// return nil if failed
	getProxyClient := func() *http.Client {
		resp, err := common.EciClient.DescribeContainerGroups(&eci20180808.DescribeContainerGroupsRequest{
			RegionId:           tea.String("cn-hangzhou"),
			ContainerGroupName: tea.String(proxyContainerGroupName),
		})
		if err != nil {
			log.Fatal().Err(err).Msg("DescribeContainerGroupsRequest failed")
		}
		log.Info().Interface("resp", resp).Msg("DescribeContainerGroupsRequest success")
		// log.Fatal().Msg("proxyContainerGroupName")
		if len(resp.Body.ContainerGroups) == 0 {
			return nil
		}

		internetIp := resp.Body.ContainerGroups[0].InternetIp
		if internetIp == nil {
			return nil
		}

		proxyURL, err := url.Parse(fmt.Sprintf("http://%s:20001", *internetIp))
		if err != nil {
			log.Fatal().Err(err).Msg("Error parsing proxy URL")
		}

		client := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}

		// visit baidu.com to test proxy
		_, err = client.Get("http://www.baidu.com")
		if err != nil {
			log.Warn().Err(err).Msg("failed to test proxy")
			return nil
		}

		pClient := qrpc.NewWorkerSlaveProtobufClient("http://localhost:9000", client)
		_, err = pClient.ProxyRefresh(context.Background(), &emptypb.Empty{})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to call ProxyRefresh")
		}

		return client
	}
	proxyClient := getProxyClient()
	if proxyClient == nil {
		log.Info().Msg("create proxy container")
		// create proxy container
		_, err := common.EciClient.CreateContainerGroup(&eci20180808.CreateContainerGroupRequest{
			RegionId:           tea.String("cn-hangzhou"),
			ContainerGroupName: tea.String(proxyContainerGroupName),
			Container: []*eci20180808.CreateContainerGroupRequestContainer{
				{
					Name:  tea.String("worker"),
					Image: tea.String(fmt.Sprintf("registry.cn-hangzhou.aliyuncs.com/117503445/biye-proxy:%s", expID)),
				},
			},
			RestartPolicy:   tea.String("Never"),
			Cpu:             tea.Float32(0.25),
			Memory:          tea.Float32(0.5),
			SpotStrategy:    tea.String("SpotAsPriceGo"),
			AutoCreateEip:   tea.Bool(true),
			SecurityGroupId: tea.String("sg-bp1c2remwvj2rpef5upy"),
			VSwitchId:       tea.String("vsw-bp1f2g1unvama51zc04cd"),
		})
		if err != nil {
			log.Fatal().Err(err).Msg("CreateContainerGroupRequest failed")
		}
	} else {
		log.Info().Msg("proxy container already exists")
	}

	for {
		log.Info().Msg("wait for proxy container")
		proxyClient = getProxyClient()
		if proxyClient != nil {
			break
		}
		time.Sleep(time.Second * 3)
	}

	stops := make([]chan struct{}, 0)

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
			RestartPolicy:    tea.String("Never"),
			Cpu:              tea.Float32(0.25),
			Memory:           tea.Float32(0.5),
			SpotStrategy:     tea.String("SpotAsPriceGo"),
			AutoCreateEip:    tea.Bool(false),
			SecurityGroupId:  tea.String("sg-bp1c2remwvj2rpef5upy"),
			VSwitchId:        tea.String("vsw-bp1f2g1unvama51zc04cd"),
			IngressBandwidth: tea.Int64(bandwidth),
			EgressBandwidth:  tea.Int64(bandwidth),
		})
		if err != nil {
			log.Fatal().Err(err).Msg("CreateContainerGroupRequest failed")
		}
		log.Info().Interface("result", result).Msg("CreateContainerGroupRequest success")

		cStop := make(chan struct{})
		go func(stop chan struct{}) {
			// collect log
			for {
				select {
				case <-stop:
					return
				default:
					result, err := common.EciClient.DescribeContainerLog(&eci20180808.DescribeContainerLogRequest{
						RegionId:         tea.String("cn-hangzhou"),
						ContainerGroupId: result.Body.ContainerGroupId,
						ContainerName:    tea.String("worker"),
					})
					if err != nil {
						log.Error().Err(err).Msg("DescribeContainerLogRequest failed")
					}
					if result.Body != nil && result.Body.Content != nil {
						goutils.WriteText(fmt.Sprintf("%s/%s/%s-%d-%d.log", dirLogs, expID, containerGroupName, meta.NodeID, meta.WorkerID), *result.Body.Content)
					}
					if nodeCount < 32 {
						time.Sleep(time.Second * 10)
					} else if nodeCount < 64 {
						time.Sleep(time.Second * 20)
					} else {
						time.Sleep(time.Second * 30)
					}

				}
			}
		}(cStop)
		stops = append(stops, cStop)

		// var internetIp *string

		for {
			result, err := common.EciClient.DescribeContainerGroups(&eci20180808.DescribeContainerGroupsRequest{
				RegionId:           tea.String("cn-hangzhou"),
				ContainerGroupName: tea.String(containerGroupName),
			})
			if err != nil {
				log.Fatal().Err(err).Msg("DescribeContainerGroupsRequest failed")
			}
			if len(result.Body.ContainerGroups) > 0 {
				// internetIp = result.Body.ContainerGroups[0].InternetIp
				intranetIp := result.Body.ContainerGroups[0].IntranetIp
				if intranetIp != nil && *intranetIp != "" {
					internetIp := ""
					if result.Body.ContainerGroups[0].InternetIp != nil {
						internetIp = *result.Body.ContainerGroups[0].InternetIp
					}

					log.Info().Str("internetIp", internetIp).Str("intranetIp", *intranetIp).Msg("DescribeContainerGroupsRequest success")
					m.Lock()
					w.Workers = append(w.Workers, &qrpc.WorkerNetInfo{
						Name:        fmt.Sprintf("biye-%d-%d", meta.NodeID, meta.WorkerID),
						InternetIp:  internetIp,
						IntranetIp:  *intranetIp,
						NodeIndex:   int64(meta.NodeID),
						WorkerIndex: int64(meta.WorkerID),
						EciId:       *result.Body.ContainerGroups[0].ContainerGroupId,
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

	go func() {
		<-stop
		for _, st := range stops {
			close(st)
		}
	}()

	for _, worker := range w.Workers {
		wg.Add(1)
		go func(worker *qrpc.WorkerNetInfo) {

			defer wg.Done()
			client := qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", worker.IntranetIp), proxyClient)
			log.Info().Str("intranetIp", worker.IntranetIp).Msg("Init WorkerSlaveProtobufClient")

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
					log.Warn().Err(err).Int("nodeIndex", int(worker.NodeIndex)).Int("workerIndex", int(worker.WorkerIndex)).Msg("failed to call PutWorkersNetInfo")
					time.Sleep(time.Second * 3)
					continue
				}
				log.Info().Msgf("resp: %v", resp)
				break
			}
		}(worker)
	}
	wg.Wait()
	// for {
	// 	_, err := proxyClient.PutWorkersNetInfoPublic(context.TODO(), w)
	// 	if err != nil {
	// 		log.Error().Err(err).Msg("failed to call PutWorkersNetInfoPublic")
	// 		time.Sleep(time.Second * 3)
	// 	} else {
	// 		break
	// 	}
	// }

	// write w to "../Docker/validators/eci.pb"
	wBytes, err := proto.Marshal(w)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal")
	}
	if err := os.WriteFile("../Docker/validators/eci.pb", wBytes, 0666); err != nil {
		log.Fatal().Err(err).Msg("failed to write file")
	}

	return w, proxyClient
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

		goutils.Exec(fmt.Sprintf("docker build -t registry.cn-hangzhou.aliyuncs.com/117503445/biye-proxy:%v .", expID), goutils.WithCwd("./assets/fc-proxy"))

		goutils.Exec(fmt.Sprintf("docker push registry.cn-hangzhou.aliyuncs.com/117503445/biye-proxy:%v", expID), goutils.WithCwd("./assets/fc-proxy"))

		// DeployECI(4, 1, make(chan struct{}), nil)

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
			if strings.Contains(*containerGroup.ContainerGroupName, "proxy") {
				continue
			}

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

		// resp, err := common.EciClient.DescribeContainerGroups(&eci20180808.DescribeContainerGroupsRequest{
		// 	RegionId:           tea.String("cn-hangzhou"),
		// 	ContainerGroupName: tea.String(proxyContainerGroupName),
		// })
		// if err != nil {
		// 	log.Fatal().Err(err).Msg("DescribeContainerGroupsRequest failed")
		// }
		// if len(resp.Body.ContainerGroups) > 0 {
		// 	// The time follows the RFC 3339 standard and must be in UTC
		// 	t := *resp.Body.ContainerGroups[0].CreationTime
		// 	createAt, err := time.Parse(time.RFC3339, t)
		// 	if err != nil {
		// 		log.Fatal().Err(err).Msg("failed to parse time")
		// 	}
		// 	if time.Since(createAt) > time.Minute*15 {
		// 		log.Info().Time("createAt", createAt).Msg("delete proxy container")
		// 		_, err := common.EciClient.DeleteContainerGroup(&eci20180808.DeleteContainerGroupRequest{
		// 			ContainerGroupId: resp.Body.ContainerGroups[0].ContainerGroupId,
		// 			RegionId:         tea.String("cn-hangzhou"),
		// 		})
		// 		if err != nil {
		// 			log.Fatal().Err(err).Msg("DeleteContainerGroupRequest failed")
		// 		}
		// 	} else {
		// 		log.Info().Time("createAt", createAt).Dur("dur", time.Since(createAt)).Msg("proxy container exists")
		// 	}
		// } else {
		// 	log.Info().Msg("proxy container not exists")
		// }

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
