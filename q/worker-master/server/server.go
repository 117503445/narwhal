package server

import (
	"context"
	// "fmt"
	"net"
	"os"
	// "q/executor/store"
	"q/rpc"
	"strconv"
	// "strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"
	// "google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	rpc.UnimplementedPrimaryToWorkerServer

	id int // id of this worker

	// executorGrpcClients map[int]rpc.ExecutorClient // id -> client

	// checkPointStore *store.CheckPointStore
}

func (s *Server) SendMessage(_ context.Context, in *rpc.BincodeEncodedPayload) (*rpc.Empty, error) {
	log.Info().Msg("UnimplementedPrimaryToWorkerServer.SendMessage")

	return &rpc.Empty{}, nil
}

func (s *Server) Run() {

	// s.checkPointStore = store.NewCheckPointStore()

	idStr := os.Getenv("EXECUTOR_ID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid EXECUTOR_ID")
	}
	s.id = id

	// executorsAddr := os.Getenv("EXECUTORS_ADDR")
	// log.Info().Str("executorsAddr", executorsAddr).Msg("")
	// // executersAddr: qexecutor_0:50051,qexecutor_1:50051,qexecutor_2:50051,qexecutor_3:50051
	// executorsAddrList := strings.Split(executorsAddr, ",")
	// s.executorGrpcClients = make(map[int]rpc.ExecutorClient, len(executorsAddrList))
	// for i, addr := range executorsAddrList {
	// 	creds := insecure.NewCredentials()
	// 	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("failed to dial")
	// 	}
	// 	s.executorGrpcClients[i] = rpc.NewExecutorClient(conn)
	// }

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	grpcS := grpc.NewServer()
	rpc.RegisterPrimaryToWorkerServer(grpcS, s)

	if err := grpcS.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}

func NewServer() *Server {
	return &Server{}
}
