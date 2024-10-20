package server

import (
	"fmt"
	"q/common"

	fc20230330 "github.com/alibabacloud-go/fc-20230330/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/rs/zerolog/log"
)

// FcManager 管理函数计算的实例
type FcManager struct {
	masterID int
}

func NewFcManager(masterID int) *FcManager {
	return &FcManager{masterID: masterID}
}

// GetInfo 获取第 index 个函数计算实例的信息
func (m *FcManager) GetInfo(index int) string {
	functionName := tea.String(fmt.Sprintf("biye-%d-%d", m.masterID, index))
	result, err := common.FcClient.ListInstancesWithOptions(functionName, &fc20230330.ListInstancesRequest{}, make(map[string]*string), &util.RuntimeOptions{})
	if err != nil {
		log.Fatal().Err(err).Msg("list instances failed")
	}
	log.Debug().Msgf("list instances: %v", result)
	return ""
}
