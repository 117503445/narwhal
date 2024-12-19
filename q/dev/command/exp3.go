package command

import (
	"context"
	"fmt"

	// "net/http"
	"q/qrpc"
	"time"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"

	// "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Exp3Param struct {
	Press       int
	N           int
	LatencyMock bool
	Bandwidth   float64
}

func Exp3RunOnce(param *Exp3Param) {
	if param.Press == 0 || param.N == 0 {
		log.Fatal().Msg("Press and Mode are required")
	}

	var err error
	goutils.Exec("docker build -t 117503445/narwhal .", goutils.WithCwd("../"))
	goutils.Exec("docker compose up -d --remove-orphans", goutils.WithCwd("../"))

	goutils.Exec("docker compose exec -T q-dev /workspace/q/script/build.sh", goutils.WithCwd("../"))

	goutils.Exec(fmt.Sprintf("docker build -t registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v .", expID), goutils.WithCwd("./assets/fc-worker"))

	goutils.Exec(fmt.Sprintf("docker push registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v", expID), goutils.WithCwd("./assets/fc-worker"))

	goutils.Exec(fmt.Sprintf("docker build -t registry.cn-hangzhou.aliyuncs.com/117503445/biye-proxy:%v .", expID), goutils.WithCwd("./assets/fc-proxy"))

	goutils.Exec(fmt.Sprintf("docker push registry.cn-hangzhou.aliyuncs.com/117503445/biye-proxy:%v", expID), goutils.WithCwd("./assets/fc-proxy"))

	log.Info().Msg("Exp3CaseCMD")

	// 80000 交易 * 512B/交易 * 3 = 120MB
	w, proxyClient := ECIDeploy(param.N, 1, make(chan struct{}), &ECIParam{
		Bandwidth:   param.Bandwidth,
		LatencyMock: param.LatencyMock,
	})
	log.Info().Interface("w", w).Msg("DeployECI")

	// var client0 qrpc.WorkerSlave
	// clients := make([]qrpc.WorkerSlave, 0)
	time.Sleep(3 * time.Second)
	clients := make(map[int]qrpc.WorkerSlave)
	for _, w := range w.Workers {
		clients[int(w.NodeIndex)] = qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", w.IntranetIp), proxyClient)
	}

	// for _, c := range clients {
	// 	go func(c qrpc.WorkerSlave) {
	// 		_, err := c.Exp2Start(context.Background(), &qrpc.Exp2StartRequest{
	// 			Press: int64(param.Press),
	// 		})

	// 		if err != nil {
	// 			log.Fatal().Err(err).Msg("failed to call Exp2Start")
	// 		}
	// 	}(c)
	// }

	client0 := clients[0]
	_, err = client0.Exp3Start(context.Background(), &qrpc.ExpStartRequest{
		Press: int64(param.Press),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to call Exp2Start")
	}

	oldTpsList := make([]float64, 0)
	oldLatencyList := make([]float64, 0)

	for {
		log.Info().Msg("Exp1GetMetrics")
		metrics, err := client0.Exp1GetMetrics(context.Background(), &emptypb.Empty{})
		if err != nil {
			log.Error().Err(err).Msg("failed to call Exp1GetMetrics")
			continue
		}
		// log.Info().Str("metrics", protojson.Format(metrics)).Msg("Exp1GetMetrics Done")
		if len(metrics.BatchMetas) < 1 {
			time.Sleep(time.Second * 10)
			continue
		}

		tps, latency := ExpMetricsCalc(metrics.BatchMetas, metrics.LatenciesMS)
		log.Info().Float64("tps", tps).Float64("latency", latency).Msg("ExpMetricsCalc")
		if tps > 0 && latency > 0 {
			oldTpsList = append(oldTpsList, tps)
			oldLatencyList = append(oldLatencyList, latency)
		}

		if len(oldTpsList) > 5 {
			getTpsChange := func(index int) float64 {
				return (oldTpsList[index] - oldTpsList[len(oldTpsList)-1]) / oldTpsList[len(oldTpsList)-1]
			}
			if getTpsChange(len(oldTpsList)-2) < 0.01 && getTpsChange(len(oldTpsList)-3) < 0.01 {
				break
			}
		}

		time.Sleep(time.Second * 10)
	}

	tps := oldTpsList[len(oldTpsList)-1]
	latency := oldLatencyList[len(oldLatencyList)-1]
	log.Info().Float64("tps", tps).Float64("latency", latency).Msg("Exp2CaseCMD Done")

	dirRoot, err := goutils.FindGitRepoRoot()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to FindGitRepoRoot")
	}

	err = goutils.WriteJSON(fmt.Sprintf("%s/paper-exp-data/%v.json", dirRoot, w.ExpId), map[string]interface{}{
		"tps":     tps,
		"latency": latency,
		"figure":  "batchsize 对 txpool 的影响",
		"line":    "broadcast-txpool",

		"debug_tps_list":     oldTpsList, // for debug
		"debug_latency_list": oldLatencyList,
		"debug_press":        param.Press,
		"debug_n":            param.N,
		"debug_latency_mock": param.LatencyMock,
		"debug_func":         "Exp3RunOnce",
		"debug_bandwidth":    param.Bandwidth,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to WriteJSON")
	}

	ECIDelete(w)

	RefreshExpID()
}
