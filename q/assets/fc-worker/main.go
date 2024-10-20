package main

import (
	// "fmt"
	"context"
	"encoding/json"
	// "fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"

	"q/qrpc"

	"google.golang.org/protobuf/types/known/emptypb"

	_ "embed"
)

//go:embed fc-urls.json
var urlMapJSON string

type URLMap struct {
	URLs map[string][]string `json:"urls"`
}

type Server struct {
	qrpc.WorkerSlave

	id int

	quit chan struct{}

	started bool // start should be called only once

	clients map[int][]qrpc.WorkerSlave

	batchStatus map[string]map[int]bool // batchID -> NodeID -> status

	sync.Mutex
}

func (s *Server) Start(ctx context.Context, in *emptypb.Empty) (*qrpc.StartResponse, error) {
	s.Lock()
	if s.started {
		s.Unlock()
		log.Info().Msg("Already started")
		return &qrpc.StartResponse{
			Msg: "Already started",
		}, nil
	}
	s.started = true
	s.Unlock()

	// time.Sleep(60 * time.Minute)
	log.Info().Msg("Start")
	<-s.quit

	return &qrpc.StartResponse{
		Msg: "Ok",
	}, nil
}

func (s *Server) Stop(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	log.Info().Msg("Stop")
	close(s.quit)

	return &emptypb.Empty{}, nil
}

func (s *Server) Exit(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	log.Info().Msg("Exit")
	os.Exit(1)

	return &emptypb.Empty{}, nil
}

func (s *Server) PutBatch(ctx context.Context, in *qrpc.PutBatchRequest) (*emptypb.Empty, error) {
	log.Info().Str("batchID", in.Id).Msg("PutBatch")

	s.Lock()
	if _, ok := s.batchStatus[in.Id]; !ok {
		s.batchStatus[in.Id] = make(map[int]bool)
		for i := range s.clients {
			s.batchStatus[in.Id][i] = false
		}
		s.batchStatus[in.Id][s.id] = true
	}
	s.Unlock()

	log.Info().Str("batchID", in.Id).Msg("Broadcast")

	var wg sync.WaitGroup
	for i, clients := range s.clients {
		client := clients[0]

		wg.Add(1)
		go func(i int, client qrpc.WorkerSlave) {
			defer wg.Done()
			if i == s.id {
				return
			}
			log.Info().Str("batchID", in.Id).Int("nodeID", i).Msg("Call ReceiveBatch")
			_, err := client.ReceiveBatch(ctx, in)
			if err != nil {
				log.Error().Err(err).Msg("PutBatch")
			}
			log.Info().Str("batchID", in.Id).Int("nodeID", i).Msg("ReceiveBatch done")

			s.Lock()
			s.batchStatus[in.Id][i] = true
			count := 0
			for _, status := range s.batchStatus[in.Id] {
				if status {
					count++
				}
			}
			if count == 3 {
				log.Info().Str("batchID", in.Id).Msg("Quorum nodes received")
			}
			s.Unlock()
		}(i, client)
	}

	log.Info().Str("batchID", in.Id).Msg("Wait")
	wg.Wait()

	return &emptypb.Empty{}, nil
}

func (s *Server) ReceiveBatch(ctx context.Context, in *qrpc.PutBatchRequest) (*emptypb.Empty, error) {
	log.Info().Str("batchID", in.Id).Msg("ReceiveBatch")

	// s.Lock()
	// if _, ok := s.batchStatus[in.Id]; !ok {
	// 	s.batchStatus[in.Id] = make(map[int]bool)
	// 	for i := range s.clients {
	// 		s.batchStatus[in.Id][i] = false
	// 	}
	// }
	// s.batchStatus[in.Id][s.id] = true
	// s.Unlock()

	return &emptypb.Empty{}, nil
}

func NewServer() *Server {
	log.Info().Msg("NewServer")
	var err error

	log.Info().Str("urlMapJSON", urlMapJSON).Msg("parse")
	var urlMap URLMap
	if err = json.Unmarshal([]byte(urlMapJSON), &urlMap); err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal")
	}

	clients := make(map[int][]qrpc.WorkerSlave)
	for strI, urls := range urlMap.URLs {
		i, err := strconv.Atoi(strI)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to convert")
		}

		clients[i] = make([]qrpc.WorkerSlave, len(urls))
		for j, url := range urls {
			log.Info().Int("i", i).Int("j", j).Str("url", url).Msg("NewWorkerSlaveProtobufClient")
			clients[i][j] = qrpc.NewWorkerSlaveProtobufClient(url, &http.Client{})
		}
	}

	slaveID := os.Getenv("SLAVE_ID")
	id, err := strconv.Atoi(slaveID)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid SLAVE_ID")
	}

	return &Server{
		quit:        make(chan struct{}),
		batchStatus: make(map[string]map[int]bool),

		clients: clients,
		id:      id,
	}
}

func main() {
	goutils.InitZeroLog()

	log.Info().Msg("Starting server...")

	rpcServer := NewServer()
	twirpHandler := qrpc.NewWorkerSlaveServer(rpcServer)

	http.Handle("/", twirpHandler)
	http.ListenAndServe(":9000", nil)
}
