package main

import (
	// "fmt"
	"context"
	"net/http"
	"sync"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"

	"q/qrpc"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	quit chan struct{}

	started bool // start should be called only once

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

func NewServer() *Server {
	return &Server{
		quit: make(chan struct{}),
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
