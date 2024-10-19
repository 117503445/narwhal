package command

import (
	"q/common"
	"q/rpc"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SendReqCmd struct {
}

func (*SendReqCmd) Run() error {
	goutils.InitZeroLog(goutils.WithNoColor{})
	log.Info().Msg("SendReq Run")

	creds := insecure.NewCredentials()
	conn, err := grpc.NewClient("localhost:4001", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to dial")
	}
	client := rpc.NewTransactionsClient(conn)

	common.SendTransactionToNarwhalWorker(client, "hello")

	return nil
}
