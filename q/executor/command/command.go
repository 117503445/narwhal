package command

import (
	"os"
	"os/signal"
	execmgr "q/executor/executor"
	"q/executor/server"
	"q/rpc"
	"syscall"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

type ExecutorCmd struct {
}

func (e *ExecutorCmd) getGrpcPort() string {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50050"
	}
	return port
}

func (e *ExecutorCmd) Run() error {
	goutils.InitZeroLog(goutils.WithNoColor{})
	log.Info().Msg("Executor Run")
	
	port := e.getGrpcPort()
	txCh := make(chan *rpc.MyTransaction, 1024)
	server := server.NewServer(txCh)
	execMgr := execmgr.NewExecMgr(txCh)

	go server.Run(port)
	execMgr.Start()

	// 使用 select 语句阻塞主线程，直到接收到终止信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞主线程
    sig := <-sigChan
    log.Info().Msgf("Received signal: %s. Shutting down...", sig)
    execMgr.Stop()

	return nil
}

