package node

import (
	"q/rpc"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

type Node struct {
	prifix           string
	parentAddr       map[string]string
	childrenAddr     map[string]string  // domain -> addr
	parentClient  	 *rpc.IndexerClient
	childClients     map[string]*rpc.IndexerClient // domain -> client
	recvCh           chan *rpc.QueryMsg
	stopChan         chan struct{}
	mutex            sync.Mutex
}

func NewNode(prifix string, parentStr string, childrenStr []string, recvCh chan *rpc.QueryMsg) *Node {
	parentAddr := make(map[string]string)
	childrenAddr := make(map[string]string)
	if parentStr != "" {
		parts := strings.Split(parentStr, ":")
		parentAddr[parts[0]] = "localhost:" + parts[1]
		
	}
	for _, child := range childrenStr {
		if child != "" {
		parts := strings.Split(child, ":")
		childrenAddr[parts[0]] = "localhost:" + parts[1]
	}

	}
    return &Node{
		prifix : prifix,
		parentAddr: parentAddr,
		childrenAddr: childrenAddr,
		recvCh: recvCh,
        stopChan: make(chan struct{}),
    }
}

func (nd *Node) Start() {



	go nd.process()
}

func (nd *Node) Stop() {
    close(nd.stopChan)
}

func (nd *Node) process() {
    for {
        select {
        case req := <-nd.recvCh:
            log.Info().Msgf("Processing index req: %v", req)
            // 在这里处理接收到的消息
			nd.handleQuery(req)
        case <-nd.stopChan:
            log.Info().Msg("Stopping process goroutine")
            return
        }
    }
}

func (nd *Node) handleQuery(req *rpc.QueryMsg) {
	switch req.Type {
	case rpc.QueryMsg_FIRST:
		// nd.handleFirstQuery(req)
	}
	
}

// func (nd *Node) handleFirstQuery(req *rpc.QueryMsg) {
// 	if nd.matchesDomain(req.Prefix) {
//         // req.ResponseCh <- &QueryResponse{NodeDomain: nd.domain}

//         return
//     }

//     ndParts := strings.Split(nd.domain, "/")
//     targetParts := strings.Split(req.Prefix, "/")

//     // 找到第一个不匹配的子域
//     var firstMismatchIndex int
//     for i := range ndParts {
//         if i >= len(targetParts) || ndParts[i] != targetParts[i] {
//             firstMismatchIndex = i
//             break
//         }
//     }

//     // 根据层级关系选择向父节点或子节点请求
//     if firstMismatchIndex < len(ndParts) {
//         // 向父节点请求
//         if nd.parentAddr != "" {
//             nd.sendQueryToParent(req)
//         }
//     } else {
//         // 向子节点请求
//         for _, addr := range nd.childrenAddr {
//             nd.sendQueryToChild(addr, req)
//         }
//     }

//     // req.ResponseCh <- nil
// }

// func (nd *Node) matchesDomain(targetDomain string) bool {
//     ndParts := strings.Split(nd.domain, "/")
//     targetParts := strings.Split(targetDomain, "/")

//     if len(ndParts) > len(targetParts) {
//         return false
//     }

//     for i := range ndParts {
//         if ndParts[i] != targetParts[i] {
//             return false
//         }
//     }

//     return true
// }

// func (nd *Node) sendQueryToParent(req *QueryRequest) {
//     conn, err := grpc.Dial(nd.parentAddr, grpc.WithInsecure(), grpc.WithBlock())
//     if err != nil {
//         log.Error().Msgf("Failed to connect to parent: %v", err)
//         return
//     }
//     defer conn.Close()

//     client := NewIndexerClient(conn)
//     ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
//     defer cancel()

//     resp, err := client.Query(ctx, &QueryRequest{TargetDomain: req.TargetDomain})
//     if err != nil {
//         log.Error().Msgf("Failed to query parent: %v", err)
//         return
//     }

//     req.ResponseCh <- &QueryResponse{NodeDomain: resp.NodeDomain}
// }

// func (nd *Node) sendQueryToChild(addr string, req *QueryRequest) {
//     conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
//     if err != nil {
//         log.Error().Msgf("Failed to connect to child: %v", err)
//         return
//     }
//     defer conn.Close()

//     client := NewIndexerClient(conn)
//     ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
//     defer cancel()

//     resp, err := client.Query(ctx, &QueryRequest{TargetDomain: req.TargetDomain})
//     if err != nil {
//         log.Error().Msgf("Failed to query child: %v", err)
//         return
//     }

//     req.ResponseCh <- &QueryResponse{NodeDomain: resp.NodeDomain}
// }