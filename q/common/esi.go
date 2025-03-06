package common

import (
	"os"

	"github.com/rs/zerolog/log"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	eci20180808 "github.com/alibabacloud-go/eci-20180808/v3/client"
	"github.com/alibabacloud-go/tea/tea"
)

var EciClient *eci20180808.Client

func init() {
	LoadENV()

	ak := os.Getenv("ak")
	if ak == "" {
		return
	}
	log.Info().Msg("init eci client")
	sk := os.Getenv("sk")
	config := &openapi.Config{
		AccessKeyId:     tea.String(ak),
		AccessKeySecret: tea.String(sk),
	}
	config.Endpoint = tea.String("eci.cn-hangzhou.aliyuncs.com")
	config.RegionId = tea.String("cn-hangzhou")
	var err error
	EciClient, err = eci20180808.NewClient(config)
	if err != nil {
		log.Fatal().Err(err).Msg("init eci client failed")
	}
}
