package common

import (
	"os"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/rs/zerolog/log"
)

var OssClient *oss.Client

func init() {
	ak := os.Getenv("ak")
	if ak == "" {
		return
	}
	log.Info().Msg("init oss client")
	sk := os.Getenv("sk")

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(ak, sk)).WithRegion("cn-hangzhou")

	OssClient = oss.NewClient(cfg)
}
