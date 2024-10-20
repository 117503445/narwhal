package main

import (
	// "fmt"
	"context"
	"net/http"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"

	"q/qrpc"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
}

func (s *Server) Start(ctx context.Context, in *emptypb.Empty) (*qrpc.StartResponse, error) {

	time.Sleep(60 * time.Minute)

	return &qrpc.StartResponse{
		Msg: "Hello, World!",
	}, nil
}

func main() {
	goutils.InitZeroLog()

	log.Info().Msg("Starting server...")

	rpcServer := &Server{}
	twirpHandler := qrpc.NewWorkerSlaveServer(rpcServer)

	http.Handle("/", twirpHandler)
	http.ListenAndServe(":9000", nil)
}
