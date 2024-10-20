package command

import (
	"context"
	"fmt"
	"io"
	"q/rpc"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
		"os"

)




type IndexerClientCmd struct {
}

func (e *IndexerClientCmd) Run() error {
    fmt.Println("Hello from client!")
	log.Debug().Msg("Hello from client!")

    // 连接到 gRPC 服务器
	conn, err := grpc.Dial("localhost:30050", grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        log.Error().Msgf("did not connect: %v", err)
    }
    defer conn.Close()
    c := rpc.NewIndexerClient(conn)

    // 创建上下文并发送请求
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
    defer cancel()
	log.Info().Msg("111111111111")
    stream, err := c.SendIndexReq(ctx)
    if err != nil {
        log.Error().Msgf("could not send request: %v", err)
        os.Exit(1)
    }
	log.Info().Msg("2222222222222")
    // 启动一个 goroutine 发送消息
    go func() {
        for {
            if err := stream.Send(&rpc.QueryMsg{Type: rpc.QueryMsg_FIRST, Prefix: "/123"}); err != nil {
                log.Error().Msgf("failed to send a request: %v", err)
                return
            }
			log.Info().Msg("send")
            time.Sleep(1 * time.Second) // 模拟发送间隔
        }
    }()

    // 接收服务器响应
    for {
        resp, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Error().Msgf("failed to receive a response: %v", err)
            break
        }
        log.Info().Msgf("Received response: %v", resp)
    }

	return nil
}

