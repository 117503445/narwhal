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
	recvCh chan *rpc.ExecuteInfo
    stopChan chan struct{}
	contract contract.SmartContract
}

func NewExecMgr(recvCh chan *rpc.ExecuteInfo) *ExecMgr {
    return &ExecMgr{
		recvCh: recvCh,
        stopChan: make(chan struct{}),
		contract: *contract.NewSmartContract(),
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
        case executeInfo := <-e.recvCh:
            log.Info().Msgf("Processing ExecuteInfo: %v", executeInfo)
            // 在这里处理接收到的消息
			e.saveExecuteInfoToFile(executeInfo, "./output") // todo 路径待解决
        case <-e.stopChan:
            log.Info().Msg("Stopping process goroutine")
            return
        }
    }
}

func (e *ExecMgr) saveExecuteInfoToFile(executeInfo *rpc.ExecuteInfo, dir string) {
    // 确保目录存在
    if err := os.MkdirAll(dir, 0755); err != nil {
        log.Error().Err(err).Msg("Failed to create directory")
        return
    }

    // 生成文件名
    filename := fmt.Sprintf("executeInfo_%d.txt", time.Now().UnixNano())
    filepath := filepath.Join(dir, filename)

    // 将 executeInfo 写入文件
    file, err := os.Create(filepath)
    if err != nil {
        log.Error().Err(err).Msg("Failed to create file")
        return
    }
    defer file.Close()

    if _, err := file.WriteString(fmt.Sprintf("%v", executeInfo)); err != nil {
        log.Error().Err(err).Msg("Failed to write executeInfo to file")
        return
    }

    log.Info().Str("file", filepath).Msg("executeInfo saved")
}

