package server

import (
	"fmt"
	"os"
	"q/qrpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type EciManager struct {
	masterID int
	clients  map[int]qrpc.WorkerSlave
}

func NewEciManager(masterID int) *EciManager {
	data, err := os.ReadFile("/validators/eci.pb")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read eci.pb")
	}

	w := &qrpc.WorkersNetInfo{}
	if err := proto.Unmarshal(data, w); err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal eci.pb")
	}

	clients := make(map[int]qrpc.WorkerSlave)
	for _, worker := range w.Workers {

		nodeID := int(worker.NodeIndex)
		workerID := int(worker.WorkerIndex)

		if nodeID == masterID {
			client := qrpc.NewWorkerSlaveProtobufClient(fmt.Sprintf("http://%s:9000", worker.InternetIp), nil)
			clients[workerID] = client
		}

	}

	return &EciManager{masterID: masterID, clients: clients}
}
