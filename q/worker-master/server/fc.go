package server

import (
	"context"
	"fmt"
	"net/http"
	"q/common"
	"q/qrpc"
	"sync"
	"time"

	"github.com/117503445/goutils"
	fc20230330 "github.com/alibabacloud-go/fc-20230330/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/rs/zerolog/log"
)

type FcConnection struct {
	instanceID int
	status     string // running, stopped
	client     qrpc.WorkerSlave

	sync.Mutex
}

func (c *FcConnection) Start() {
	go func() {
		log.Info().Int("instanceID", c.instanceID).Msg("FcConnection Start")

		go func() {
			time.Sleep(10 * time.Second)
			log.Info().Int("instanceID", c.instanceID).Msg("FcConnection Stop")
			c.client.Stop(context.Background(), nil)
			log.Info().Int("instanceID", c.instanceID).Msg("FcConnection Stop done")
		}()
		
		c.client.Start(context.Background(), nil)
		log.Info().Int("instanceID", c.instanceID).Msg("FcConnection Start done")
	}()
}

func NewFcConnection(masterID int, instanceID int, baseURL string) *FcConnection {
	client := qrpc.NewWorkerSlaveProtobufClient(baseURL, &http.Client{})

	return &FcConnection{instanceID: instanceID, status: "stopped", client: client}
}

// FcManager 管理函数计算的实例
type FcManager struct {
	masterID int
	conns    map[int]*FcConnection
}

type URLMap struct {
	URLs map[string][]string `json:"urls"`
}

func NewFcManager(masterID int) *FcManager {
	var urlMap map[string][]string
	if err := goutils.ReadJSON("/validators/fc-urls.json", &urlMap); err != nil {
		log.Fatal().Err(err).Msg("read fc-urls.json failed")
	}

	conns := make(map[int]*FcConnection)
	for i := 0; i < 1; i++ {
		conns[i] = NewFcConnection(masterID, i, urlMap[fmt.Sprintf("%d", masterID)][i])
	}
	return &FcManager{masterID: masterID, conns: conns}
}

// IsInstanceRunning 获取第 index 个函数计算实例的信息
func (m *FcManager) IsInstanceRunning(index int) bool {
	functionName := tea.String(fmt.Sprintf("biye-%d-%d", m.masterID, index))
	result, err := common.FcClient.ListInstances(functionName, &fc20230330.ListInstancesRequest{})
	if err != nil {
		log.Fatal().Err(err).Msg("list instances failed")
	}

	// log.Debug().Msgf("list instances: %v", result)
	return len(result.Body.Instances) > 0
}

func (m *FcManager) MustStartInstance(index int) {
	if m.conns[index] == nil {
		log.Fatal().Int("index", index).Msg("FcConnection not found")
	}
	m.conns[index].Start()
}
