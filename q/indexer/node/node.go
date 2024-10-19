package node

import (
	"q/rpc"

	"github.com/rs/zerolog/log"
)

type Node struct {
	recvCh chan *rpc.IndexerReq
	stopChan chan struct{}
}

func NewNode(recvCh chan *rpc.IndexerReq) *Node {
    return &Node{
		recvCh: recvCh,
        stopChan: make(chan struct{}),
    }
}

func (nd *Node) Start() {
	go nd.process()
}

func (nd *Node) Stop() {
    close(nd.stopChan)
}

func (nd *Node) process() {
    for {
        select {
        case req := <-nd.recvCh:
            log.Info().Msgf("Processing index req: %v", req)
            // 在这里处理接收到的消息
			// nd.handleExecute(transaction)
        case <-nd.stopChan:
            log.Info().Msg("Stopping process goroutine")
            return
        }
    }
}