package execmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"q/executor/contract"
	"q/rpc"
	"time"

	"github.com/rs/zerolog/log"
)

type ExecMgr struct {
	recvCh chan *rpc.MyTransaction
    stopChan chan struct{}
	contract contract.SmartContract
}

func NewExecMgr(recvCh chan *rpc.MyTransaction) *ExecMgr {

	// 从环境变量中获取账本文件路径
	nodeType := os.Getenv("NODE_TYPE")
	workerID := os.Getenv("WORKER_ID")
	if nodeType == "" || workerID == "" {
		log.Fatal().Msg("NODE_TYPE 或 WORKER_ID 环境变量未设置")
	}
	filePath := fmt.Sprintf("/data/%s-%s/ledger.json", nodeType, workerID)

    return &ExecMgr{
		recvCh: recvCh,
        stopChan: make(chan struct{}),
		contract: *contract.NewSmartContract(filePath),
    }
}

func (e *ExecMgr) Start() {
	go e.process()
}

func (e *ExecMgr) Stop() {
    close(e.stopChan)
}

func (e *ExecMgr) process() {
    for {
        select {
        case transaction := <-e.recvCh:
            log.Info().Msgf("Processing transaction: %v", transaction)
            // 在这里处理接收到的消息
			e.saveToFile(transaction, "/data") // test
        case <-e.stopChan:
            log.Info().Msg("Stopping process goroutine")
            return
        }
    }
}

func (e *ExecMgr) saveToFile(transaction *rpc.MyTransaction, dir string) {
    // 确保目录存在
    if err := os.MkdirAll(dir, 0755); err != nil {
        log.Error().Err(err).Msg("Failed to create directory")
        return
    }

    // 生成文件名
    filename := fmt.Sprintf("transaction_%d.txt", time.Now().UnixNano())
    filepath := filepath.Join(dir, filename)

    // 将 executeInfo 写入文件
    file, err := os.Create(filepath)
    if err != nil {
        log.Error().Err(err).Msg("Failed to create file")
        return
    }
    defer file.Close()

    if _, err := file.WriteString(fmt.Sprintf("%v", transaction)); err != nil {
        log.Error().Err(err).Msg("Failed to write transaction to file")
        return
    }

    log.Info().Str("file", filepath).Msg("transaction saved")
}

