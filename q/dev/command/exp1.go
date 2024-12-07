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

func (cmd *Exp1CaseCMD) Run() error {
	var err error
	goutils.Exec("docker compose up -d", goutils.WithCwd("../"))

	goutils.Exec("docker compose exec -T q-dev /workspace/q/script/build.sh", goutils.WithCwd("../"))

	goutils.Exec(fmt.Sprintf("docker build -t registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v .", expID), goutils.WithCwd("./assets/fc-worker"))

	goutils.Exec(fmt.Sprintf("docker push registry.cn-hangzhou.aliyuncs.com/117503445/biye-slave:%v", expID), goutils.WithCwd("./assets/fc-worker"))

	log.Info().Msg("Exp1CaseCMD")

	w := DeployECI(4, 1, make(chan struct{}))
	log.Info().Interface("w", w).Msg("DeployECI")

	clients := make([]qrpc.WorkerSlave, 0)
	time.Sleep(3 * time.Second)
	for _, w := range w.Workers {
		c := qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", w.InternetIp), &http.Client{})
		clients = append(clients, c)
	}

	_, err = clients[0].Exp1BoradcastStart(context.Background(), &qrpc.ExpStartRequest{
		Ak:    os.Getenv("ak"),
		Sk:    os.Getenv("sk"),
		Press: 1000,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to call Exp1BoradcastStart")
	}

	for {
		log.Info().Msg("Exp1GetMetrics")
		metrics, err := clients[0].Exp1GetMetrics(context.Background(), &emptypb.Empty{})
		if err != nil {
			log.Error().Err(err).Msg("failed to call Exp1GetMetrics")
			continue
		}
		// log.Info().Str("metrics", protojson.Format(metrics)).Msg("Exp1GetMetrics Done")
		if len(metrics.BatchMetas) < 1 {
			continue
		}

		tps, latency := ExpMetricsCalc(metrics.BatchMetas, metrics.LatenciesMS)
		log.Info().Float64("tps", tps).Float64("latency", latency).Msg("ExpMetricsCalc")
		time.Sleep(time.Second * 10)
	}

	return err
}

const EXP_BATCH_SIZE = 10 * 1024 * 1024

// ExpMetricsCalc 计算 TPS 和 延迟
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
