package command

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"q/qrpc"
	"time"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
	// "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Exp1CaseCMD struct {
}

type Exp1Param struct {
	Press int
	Mode  string // broadcast or p2p
	N     int
}

func Exp1RunOnce(param *Exp1Param) {
	if param.Press == 0 || param.Mode == "" || param.N == 0 {
		log.Fatal().Msg("Press and Mode are required")
	}

	var err error
	goutils.Exec("docker compose up -d", goutils.WithCwd("../"))

	goutils.Exec("docker compose exec -T q-dev /workspace/q/script/build.sh", goutils.WithCwd("../"))

	goutils.Exec(fmt.Sprintf("docker build -t registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v .", expID), goutils.WithCwd("./assets/fc-worker"))

	goutils.Exec(fmt.Sprintf("docker push registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v", expID), goutils.WithCwd("./assets/fc-worker"))

	log.Info().Msg("Exp1CaseCMD")

	// 80000 交易 * 512B/交易 * 3 = 120MB
	w := DeployECI(param.N, 1, make(chan struct{}), &ECIParam{
		Bandwidth: 12.5,
	})
	log.Info().Interface("w", w).Msg("DeployECI")

	clients := make([]qrpc.WorkerSlave, 0)
	time.Sleep(3 * time.Second)
	for _, w := range w.Workers {
		c := qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", w.InternetIp), &http.Client{})
		clients = append(clients, c)
	}

	if param.Mode == "broadcast" {
		_, err = clients[0].Exp1BoradcastStart(context.Background(), &qrpc.ExpStartRequest{
			Ak:    os.Getenv("ak"),
			Sk:    os.Getenv("sk"),
			Press: int64(param.Press),
		})
	} else {
		_, err = clients[0].Exp1P2PStart(context.Background(), &qrpc.ExpStartRequest{
			Ak:    os.Getenv("ak"),
			Sk:    os.Getenv("sk"),
			Press: int64(param.Press),
		})
	}

	if err != nil {
		log.Fatal().Err(err).Msg("failed to call Exp1BoradcastStart")
	}

	oldTpsList := make([]float64, 0)
	oldLatencyList := make([]float64, 0)

	for {
		log.Info().Msg("Exp1GetMetrics")
		metrics, err := clients[0].Exp1GetMetrics(context.Background(), &emptypb.Empty{})
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
	log.Info().Float64("tps", tps).Float64("latency", latency).Msg("Exp1CaseCMD Done")

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
		"debug_mode":         param.Mode,
		"debug_n":            param.N,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to WriteJSON")
	}
	RefreshExpID()
}

func (cmd *Exp1CaseCMD) Run() error {
	// for _, press := range []int{1000000} {
	// 	Exp1RunOnce(&Exp1Param{
	// 		Press: press,
	// 		Mode:  "broadcast",
	// 		N:     8,
	// 	})
	// }

	for _, press := range []int{1000000} {
		Exp1RunOnce(&Exp1Param{
			Press: press,
			Mode:  "p2p",
			N:     4,
		})
	}

	return nil
}

const EXP_BATCH_SIZE = 10 * 1024 * 1024

// ExpMetricsCalc 计算 TPS 和 延迟, 0 代表无效
func ExpMetricsCalc(batches []*qrpc.ExpBatchMeta, latenciesMS []int64) (float64, float64) {
	return ExpMetricsTps(batches), ExpMetricsLatency(latenciesMS)
}

// ExpMetricsTps 计算 TPS
func ExpMetricsTps(batches []*qrpc.ExpBatchMeta) float64 {
	// log.Info().Interface("batches", batches).Msg("ExpMetricsTps")
	if len(batches) == 0 {
		return 0
	}
	dur := batches[len(batches)-1].SubmittedAt.AsTime().Sub(batches[0].SubmittedAt.AsTime())
	if dur.Seconds() == 0 {
		return 0
	}
	txNum := 0
	for _, b := range batches {
		txNum += int(b.TxNum)
	}
	tps := float64(txNum) / dur.Seconds()
	return tps
}

// ExpMetricsLatency 计算延迟
func ExpMetricsLatency(latenciesMS []int64) float64 {
	if len(latenciesMS) == 0 {
		return 0
	}
	var sum int64
	for _, l := range latenciesMS {
		sum += l
	}

	dur := time.Duration(sum / int64(len(latenciesMS)) * int64(time.Millisecond))

	return dur.Seconds()
}
