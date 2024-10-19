package common

import (
	"bytes"
	"context"
	"q/rpc"

	"github.com/rs/zerolog/log"
)

// SendTransactionToNarwhalWorker 向 Narwhal Worker 发送一笔交易
func SendTransactionToNarwhalWorker(client rpc.TransactionsClient, payload string) error {
	var err error
	size := 1024 // 假设大小为1024
	// var r uint64 = 0

	// 创建一个字节缓冲区
	tx := make([]byte, size)
	buf := bytes.NewBuffer(tx[:0])

	// 将标准交易的标识符和计数器值放入缓冲区
	// r++
	buf.WriteByte(1) // 标准交易以 1 开头
	// binary.Write(buf, binary.LittleEndian, r) // 确保所有客户端发送不同的交易
	buf.Write([]byte(payload))

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
	return err
}
