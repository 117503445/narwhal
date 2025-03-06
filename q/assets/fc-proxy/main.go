package main

import (
	"net/http"
	"q/qrpc"
	"time"

	"context"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"os/exec"
)

type Server struct {
	qrpc.WorkerSlave
}

var LastRefreshTime time.Time = time.Now()

func (s *Server) ProxyRefresh(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	LastRefreshTime = time.Now()
	log.Info().Msg("ProxyRefresh")
	return &emptypb.Empty{}, nil
}

func main() {
	goutils.InitZeroLog(goutils.WithNoColor{})

	go func() {
		for {
			time.Sleep(5 * time.Second)
			if time.Since(LastRefreshTime) > 10*time.Minute {
				log.Fatal().Msg("proxy is dead")
			}
		}
	}()

	go func() {
		// 长期运行的任务
		// sing-box -D /var/lib/sing-box -C /etc/sing-box run
		err := exec.Command("sing-box", "-D", "/var/lib/sing-box", "-C", "/etc/sing-box", "run").Run()
		if err != nil {
			log.Fatal().Err(err).Msg("sing-box run failed")
		}
	}()

	rpcServer := NewServer()
	twirpHandler := qrpc.NewWorkerSlaveServer(rpcServer)

	http.Handle("/", twirpHandler)
	http.ListenAndServe(":9000", nil)
}

func NewServer() *Server {
	log.Info().Msg("NewServer")

	return &Server{}
}
