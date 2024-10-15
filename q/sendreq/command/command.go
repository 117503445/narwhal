package command

import (
	"context"
	"q/rpc"

	"bytes"
	"encoding/binary"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SendReqCmd struct {
}

func (*SendReqCmd) Run() error {
	goutils.InitZeroLog(goutils.WithNoColor{})
	log.Info().Msg("SendReq Run")

	creds := insecure.NewCredentials()
	conn, err := grpc.NewClient("localhost:4001", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to dial")
	}
	client := rpc.NewTransactionsClient(conn)

	size := 1024 // 假设大小为1024
	var r uint64 = 0

	// 创建一个字节缓冲区
	tx := make([]byte, size)
	buf := bytes.NewBuffer(tx[:0])

	// 将标准交易的标识符和计数器值放入缓冲区
	r++
	buf.WriteByte(1)                          // 标准交易以 1 开头
	binary.Write(buf, binary.LittleEndian, r) // 确保所有客户端发送不同的交易

	// 调整缓冲区大小以匹配指定的大小
	if buf.Len() < size {
		buf.Write(make([]byte, size-buf.Len()))
	}

	// 将缓冲区转换为字节数组
	tx = buf.Bytes()

	_, err = client.SubmitTransaction(context.Background(), &rpc.Transaction{Transaction: tx})
	if err != nil {
		log.Fatal().Err(err).Msg("SubmitTransaction")
	}

	return nil
}
