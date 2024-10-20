package server

import (
	"context"
	"time"

	// "fmt"
	"net"
	"os"

	// "q/executor/store"
	"q/common"
	"q/qrpc"
	"q/rpc"
	"strconv"

	// "strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	rpc.UnimplementedQWorkerMasterServer

	id int // id of this master server

	// executorGrpcClients map[int]rpc.ExecutorClient // id -> client

	transactionsClient rpc.TransactionsClient

	fcManager *FcManager
}

func (s *Server) PutTestTx(ctx context.Context, in *rpc.QTransaction) (*emptypb.Empty, error) {
	common.SendTransactionToNarwhalWorker(s.transactionsClient, in.Payload)

	return &emptypb.Empty{}, nil
}

func (s *Server) PutTx(ctx context.Context, in *rpc.QTransaction) (*emptypb.Empty, error) {
	// common.SendTransactionToNarwhalWorker(s.transactionsClient, in.Payload)
	log.Info().Msg("PutTx")
	s.fcManager.MustStartInstance(0)

	time.Sleep(1 * time.Second)
	if _, err := s.fcManager.conns[0].client.PutBatch(context.Background(), &qrpc.PutBatchRequest{
		Payload: in.Payload,
		Id:      "batch0",
	}); err != nil {
		log.Fatal().Err(err).Msg("failed to PutBatch")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Run() {

	// s.checkPointStore = store.NewCheckPointStore()

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
	rpc.RegisterQWorkerMasterServer(grpcS, s)

	if err := grpcS.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}

func NewServer() *Server {
	idStr := os.Getenv("WORKER_MASTER_ID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid WORKER_MASTER_ID")
	}

	creds := insecure.NewCredentials()
	conn, err := grpc.NewClient("localhost:4001", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to dial")
	}
	client := rpc.NewTransactionsClient(conn)

	s := &Server{
		transactionsClient: client,
		id:                 id,
		fcManager:          NewFcManager(id),
	}

	if s.fcManager.IsInstanceRunning(0) {
		s.fcManager.conns[0].client.Exit(context.Background(), nil)
		time.Sleep(1 * time.Second)
		if s.fcManager.IsInstanceRunning(0){
			log.Fatal().Msg("failed to stop instance")
		}
	}

	// 测试用
	go func() {
		time.Sleep(1 * time.Second)
		s.PutTx(context.Background(), &rpc.QTransaction{})
	}()

	return s
}
