package common

import (
	"bytes"
	"context"
	"q/rpc"
	"sync"

	"os"
	"strings"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

// 放大 1000 倍
const NarwhalTxN = 10000

// SendTransactionToNarwhalWorker 向 Narwhal Worker 发送交易
func SendTransactionToNarwhalWorker(client rpc.TransactionsClient, payload string, n int) error {
	var err error
	size := 512 * NarwhalTxN // 假设大小为1024
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

	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err = client.SubmitTransaction(context.Background(), &rpc.Transaction{Transaction: tx})
			if err != nil {
				log.Fatal().Err(err).Msg("SubmitTransaction")
			}
		}()
	}
	wg.Wait()

	return err
}

func LoadENV() {
	fileENV := "../Docker/.env"
	if goutils.PathExists(fileENV) {
		env, err := goutils.ReadText(fileENV)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read file")
		}
		for _, line := range strings.Split(env, "\n") {
			if line == "" {
				continue
			}
			parts := strings.Split(line, "=")
			if len(parts) != 2 {
				log.Fatal().Msg("invalid .env")
			}
			os.Setenv(parts[0], parts[1])
		}
	}
}
