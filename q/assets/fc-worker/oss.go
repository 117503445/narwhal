package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

// 	OssCase0Start(context.Context, *google_protobuf.Empty) (*google_protobuf.Empty, error)

func (s *Server) OssCase0Start(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	log.Info().Msg("OssCase0Start")

	return &emptypb.Empty{}, nil
}
