package command

import (
	"os"
	"os/signal"
	"q/indexer/common"
	"q/indexer/node"
	"q/indexer/server"
	"q/rpc"
	"strings"
	"syscall"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

type IndexerCmd struct {
}

func (ic *IndexerCmd) getGrpcPort() string {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "30050"
	}
	return port
}

func (ic *IndexerCmd) Run() error {
	goutils.InitZeroLog(goutils.WithNoColor{})
	log.Info().Msg("Indexer Run")

	port := ic.getGrpcPort()
	addr := os.Getenv("NODE_TYPE") + "_" + os.Getenv("WORKER_ID") + ":" + port
	prefix := os.Getenv("PRIFIX")
    parentStr := os.Getenv("PARENT_STR")
    childStr := os.Getenv("CHILD_STR")
	log.Info().Msgf("prefix: %s, parentStr: %s, childStr: %s", prefix, parentStr, childStr)
	// 解析 childStr
    var children []string
    if childStr != "" {
        children = strings.Split(childStr, " ")
    }

	reqCh := make(chan *common.ReqWithCh, 1024)
	respCh := make(chan *rpc.QueryMsg, 1024)
	node := node.NewNode(prefix, addr, parentStr, children, reqCh, respCh)
	server := server.NewServer(node, reqCh, respCh)
	go server.Run(port)
	node.Start()

	// 使用 select 语句阻塞主线程，直到接收到终止信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞主线程
    sig := <-sigChan
    log.Info().Msgf("Received signal: %s. Shutting down...", sig)
    // execMgr.Stop()

	return nil
}

