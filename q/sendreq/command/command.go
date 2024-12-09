package command

import (
	"q/common"
	"q/rpc"
	"sync"
	"time"

	// "time"

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

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				common.SendTransactionToNarwhalWorker(client, "hello", 1)
				time.Sleep(1 * time.Second)
			}
		}()
	}

	wg.Wait()

	return nil
}
