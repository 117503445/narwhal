package common

import (
	"os"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	fc20230330 "github.com/alibabacloud-go/fc-20230330/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/rs/zerolog/log"
)

var FcClient *fc20230330.Client

func init() {
	ak := os.Getenv("ak")
	if ak == "" {
		return
	}
	sk := os.Getenv("sk")

	config := &openapi.Config{
		AccessKeyId:     tea.String(ak),
		AccessKeySecret: tea.String(sk),
	}
	// Endpoint 请参考 https://api.aliyun.com/product/FC
	config.Endpoint = tea.String("1035038953803932.cn-hangzhou.fc.aliyuncs.com")
	var err error
	FcClient, err = fc20230330.NewClient(config)
	if err != nil {
		log.Fatal().Err(err).Msg("init fc client failed")
	}
}
