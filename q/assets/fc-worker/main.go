package main

import (
	// "fmt"
	"context"
	"encoding/json"
	"fmt"
	"time"

	// "fmt"
	"net/http"
	"net/url"
	"os"

	// "strconv"
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

	masterId int
	slaveId  int

	quit chan struct{}

	started bool // start should be called only once

	clients map[int]map[int]qrpc.WorkerSlave

	batchStatus map[string]map[int]bool // batchID -> NodeID -> status

	masterClient qrpc.WorkerMaster

	sync.Mutex
}

func (s *Server) PutWorkersNetInfo(ctx context.Context, in *qrpc.WorkersNetInfo) (*emptypb.Empty, error) {
	log.Info().Interface("workersNetInfo", in).Msg("PutWorkersNetInfo")

	s.Lock()
	defer s.Unlock()

	s.clients = make(map[int]map[int]qrpc.WorkerSlave)
	for _, worker := range in.Workers {
		nodeID := int(worker.NodeIndex)
		workerID := int(worker.WorkerIndex)
		if _, ok := s.clients[nodeID]; !ok {
			s.clients[nodeID] = make(map[int]qrpc.WorkerSlave)
			s.clients[nodeID][workerID] = qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", worker.IntranetIp), &http.Client{})
		}
	}
	go func() {
		time.Sleep(10 * time.Second)
		log.Info().Msg("PutBatch")
		_, err := s.PutBatch(context.Background(), &qrpc.PutBatchRequest{
			Id:      fmt.Sprintf("from %d", s.slaveId),
			Payload: fmt.Sprintf("from %d", s.slaveId),
		})
		if err != nil {
			log.Fatal().Err(err).Msg("PutBatch failed")
		}
		log.Info().Msg("PutBatch done")
	}()

	s.masterId = int(in.MasterId)
	s.slaveId = int(in.SlaveId)

	// if err := os.Setenv("HTTPS_PROXY", in.Proxy); err != nil {
	// 	log.Fatal().Err(err).Msg("failed to set HTTPS_PROXY")
	// }
	// log.Info().Str("proxy", in.Proxy).Msg("set HTTPS_PROXY")
	// creds := insecure.NewCredentials()
	// conn, err := grpc.NewClient(in.MasterUrl, grpc.WithTransportCredentials(creds))

	// if err != nil {
	// 	log.Fatal().Err(err).Msg("failed to dial")
	// }

	log.Info().Str("masterUrl", in.MasterUrl).Msg("NewWorkerMasterProtobufClient")

	httpClient := &http.Client{}
	// proxy
	if in.Proxy != "" {
		proxyUrl, err := url.Parse(in.Proxy)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse proxy")
		}
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}

	s.masterClient = qrpc.NewWorkerMasterProtobufClient(in.MasterUrl, httpClient)

	return &emptypb.Empty{}, nil
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
		s.batchStatus[in.Id][s.slaveId] = true
	}
	s.Unlock()

	log.Info().Str("batchID", in.Id).Msg("Broadcast")

	var wg sync.WaitGroup
	for i, clients := range s.clients {
		client := clients[0]

		wg.Add(1)
		go func(i int, client qrpc.WorkerSlave) {
			defer wg.Done()
			if i == s.slaveId {
				return
			}
			log.Info().Str("batchID", in.Id).Int("nodeID", i).Msg("Call ReceiveBatch")
			_, err := client.ReceiveBatch(context.Background(), in)
			if err != nil {
				log.Fatal().Err(err).Msg("ReceiveBatch")
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
				go func(id string) {
					_, err := s.masterClient.BatchConfirmed(context.Background(), &qrpc.BatchMeta{
						Id: id,
					})
					if err != nil {
						log.Warn().Err(err).Msg("failed to call BatchConfirmed")
						time.Sleep(30 * time.Second)
						_, err = s.masterClient.BatchConfirmed(context.Background(), &qrpc.BatchMeta{
							Id: id,
						})
						if err != nil {
							log.Fatal().Err(err).Msg("failed to call BatchConfirmed")
						}
					}
				}(in.Id)
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

	go func() {
		time.Sleep(10 * time.Minute)
		log.Info().Msg("Timeout")
		os.Exit(1)
	}()

	log.Info().Str("urlMapJSON", urlMapJSON).Msg("parse")
	var urlMap URLMap
	if err = json.Unmarshal([]byte(urlMapJSON), &urlMap); err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal")
	}

	// clients := make(map[int][]qrpc.WorkerSlave)
	// for strI, urls := range urlMap.URLs {
	// 	i, err := strconv.Atoi(strI)
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("failed to convert")
	// 	}

	// 	clients[i] = make([]qrpc.WorkerSlave, len(urls))
	// 	for j, url := range urls {
	// 		log.Info().Int("i", i).Int("j", j).Str("url", url).Msg("NewWorkerSlaveProtobufClient")
	// 		clients[i][j] = qrpc.NewWorkerSlaveProtobufClient(url, &http.Client{})
	// 	}
	// }

	// slaveID := os.Getenv("SLAVE_ID")
	// id, err := strconv.Atoi(slaveID)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("invalid SLAVE_ID")
	// }

	return &Server{
		quit:        make(chan struct{}),
		batchStatus: make(map[string]map[int]bool),

		// clients: clients,
		// id:      id,
	}
}

func main() {
	goutils.InitZeroLog(goutils.WithNoColor{})

	log.Info().Msg("Starting server...")

	rpcServer := NewServer()
	twirpHandler := qrpc.NewWorkerSlaveServer(rpcServer)

	http.Handle("/", twirpHandler)
	http.ListenAndServe(":9000", nil)
}
