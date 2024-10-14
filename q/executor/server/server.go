package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"q/executor/store"
	"q/rpc"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	rpc.UnimplementedExecutorServer

	id int // id of this executor

	executorGrpcClients map[int]rpc.ExecutorClient // id -> client

	checkPointStore *store.CheckPointStore
}

func (s *Server) PutExecuteInfo(_ context.Context, in *rpc.ExecuteInfo) (*emptypb.Empty, error) {
	log.Info().Int32("ConsensusRound", in.ConsensusRound).Int32("ExecuteHeight", in.ExecuteHeight).Msg("PutExecuteInfo")

	go func() {
		shouldGenerateCheckPoint := true
		if !shouldGenerateCheckPoint {
			return
		}

		// generate checkPoint
		checkPoint := &rpc.SignedCheckpoint{
			Checkpoint: &rpc.Checkpoint{
				ExecuteHeight:  in.ExecuteHeight,
				ConsensusRound: in.ConsensusRound,
			},
			AuthorId:  int32(s.id),
			Signature: fmt.Sprintf("sign_by_%d", s.id),
		}
		s.checkPointStore.AddPendingCheckPoint(checkPoint)
	}()

	return &emptypb.Empty{}, nil
}

// executersAddr: http://qexecutor_0:50051,http://qexecutor_1:50051,http://qexecutor_2:50051,http://qexecutor_3:50051
func (s *Server) Run(port int, executorsAddr string) {

	s.checkPointStore = store.NewCheckPointStore()

	idStr := os.Getenv("EXECUTOR_ID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid EXECUTOR_ID")
	}
	s.id = id

	executorsAddrList := strings.Split(executorsAddr, ",")
	s.executorGrpcClients = make(map[int]rpc.ExecutorClient, len(executorsAddrList))
	for i, addr := range executorsAddrList {
		conn, err := grpc.NewClient(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to dial")
		}
		s.executorGrpcClients[i] = rpc.NewExecutorClient(conn)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	grpcS := grpc.NewServer()
	rpc.RegisterExecutorServer(grpcS, s)

	if err := grpcS.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}

func NewServer() *Server {
	return &Server{}
}
