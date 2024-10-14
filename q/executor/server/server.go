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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	rpc.UnimplementedExecutorServer

	id int // id of this executor

	executorGrpcClients map[int]rpc.ExecutorClient // id -> client

	checkPointStore *store.CheckPointStore
}

func (s *Server) broadcastQuorumCheckPoint(checkPoint *rpc.QuorumCheckpoint) {
	for id, client := range s.executorGrpcClients {
		if id == s.id {
			continue
		}
		go func(client rpc.ExecutorClient, checkPoint *rpc.QuorumCheckpoint) {
			_, err := client.PutQuorumCheckpoint(context.Background(), checkPoint)
			if err != nil {
				log.Error().Err(err).Msg("PutQuorumCheckpoint")
			}
		}(client, checkPoint)
	}
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
		newQuorumCheckpoint := s.checkPointStore.AddSignedCheckpoint(checkPoint)
		if newQuorumCheckpoint != nil {
			success := s.checkPointStore.AddQuorumCheckPoint(newQuorumCheckpoint)
			if success {
				go s.broadcastQuorumCheckPoint(newQuorumCheckpoint)
			}
		}

		for id, client := range s.executorGrpcClients {
			if id == s.id {
				continue
			}
			go func(client rpc.ExecutorClient, checkPoint *rpc.SignedCheckpoint) {
				_, err := client.PutSignedCheckpoint(context.Background(), checkPoint)
				if err != nil {
					log.Error().Err(err).Msg("PutSignedCheckpoint")
				}
			}(client, checkPoint)
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Server) PutSignedCheckpoint(_ context.Context, in *rpc.SignedCheckpoint) (*emptypb.Empty, error) {
	log.Info().Int64("executeHeight", int64(in.Checkpoint.ExecuteHeight)).Int64("authorId", int64(in.AuthorId)).Msg("PutSignedCheckpoint")
	newQuorumCheckpoint := s.checkPointStore.AddSignedCheckpoint(in)
	if newQuorumCheckpoint != nil {
		success := s.checkPointStore.AddQuorumCheckPoint(newQuorumCheckpoint)
		if success {
			go s.broadcastQuorumCheckPoint(newQuorumCheckpoint)
		}
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) PutQuorumCheckpoint(_ context.Context, in *rpc.QuorumCheckpoint) (*emptypb.Empty, error) {
	log.Info().Int64("executeHeight", int64(in.Checkpoint.ExecuteHeight)).Msg("PutQuorumCheckpoint")
	s.checkPointStore.AddQuorumCheckPoint(in)
	return &emptypb.Empty{}, nil
}

func (s *Server) Run() {

	s.checkPointStore = store.NewCheckPointStore()

	idStr := os.Getenv("EXECUTOR_ID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid EXECUTOR_ID")
	}
	s.id = id

	executorsAddr := os.Getenv("EXECUTORS_ADDR")
	log.Info().Str("executorsAddr", executorsAddr).Msg("")
	// executersAddr: qexecutor_0:50051,qexecutor_1:50051,qexecutor_2:50051,qexecutor_3:50051
	executorsAddrList := strings.Split(executorsAddr, ",")
	s.executorGrpcClients = make(map[int]rpc.ExecutorClient, len(executorsAddrList))
	for i, addr := range executorsAddrList {
		creds := insecure.NewCredentials()
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
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
