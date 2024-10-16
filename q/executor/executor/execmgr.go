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
    return &ExecMgr{
		recvCh: recvCh,
        stopChan: make(chan struct{}),
		contract: *contract.NewSmartContract("/data"),
    }
}

func (e *ExecMgr) Start() {
	go e.process()
}

func (e *ExecMgr) Stop() {
    close(e.stopChan)
	e.contract.Stop()
}

func (e *ExecMgr) process() {
    for {
        select {
        case transaction := <-e.recvCh:
            log.Info().Msgf("Processing transaction: %v", transaction)
            // 在这里处理接收到的消息
			// e.saveToFile(transaction, "/data") // test
			e.handleExecute(transaction)
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

func (e *ExecMgr) handleExecute(transaction *rpc.MyTransaction) {
	// 日志打印transaction.TxType
	log.Info().Msgf("transaction.TxType: %v", transaction.TxType)
    switch transaction.TxType {
    case rpc.MyTransaction_REGISTTX:
        e.contract.CreateAsset(transaction.Id, transaction.Value)
    case rpc.MyTransaction_UPDATETX:
        e.contract.UpdateAsset(transaction.Id, transaction.Value)
	case rpc.MyTransaction_DELETETX:
		e.contract.DeleteAsset(transaction.Id)
    // 添加更多的交易类型和对应的处理方法
    default:
        log.Warn().Msgf("Unknown transaction type: %s", transaction.TxType)
    }
	// todo 执行结果外传到索引节点 可以用gookit/ini来读取索引节点配置
}



