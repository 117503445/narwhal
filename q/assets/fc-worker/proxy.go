package main

import (
	"context"
	"sync"
	"time"

	"q/qrpc"

	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) PutWorkersNetInfoPublic(ctx context.Context, in *qrpc.WorkersNetInfo) (*emptypb.Empty, error) {
	log.Info().Interface("workersNetInfo", in).Msg("PutWorkersNetInfo")

	lastActiveTime = time.Now()
	isProxy = true

	var wg sync.WaitGroup
	for _, worker := range in.Workers {
		wg.Add(1)
		go func(worker *qrpc.WorkerNetInfo) {
			defer wg.Done()
			c := qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", worker.IntranetIp), &http.Client{})
			log.Info().Str("intranetIp", worker.IntranetIp).Msg("PutWorkersNetInfo")
			_, err := c.PutWorkersNetInfo(context.Background(),
				&qrpc.WorkersNetInfo{
					ExpId:   in.ExpId,
					Workers: in.Workers,
					Proxy:   in.Proxy,

					// MasterUrl: fmt.Sprintf("http://%v:2412%d", masterIp, worker.NodeIndex),
					MasterId: worker.NodeIndex,
					SlaveId:  worker.WorkerIndex,
				})
			if err != nil {
				log.Fatal().Err(err).Msg("PutWorkersNetInfo")
			}
		}(worker)
	}
	wg.Wait()

	return &emptypb.Empty{}, nil
}
