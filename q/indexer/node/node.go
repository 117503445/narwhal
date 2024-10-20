package node

import (
	"context"
	"q/indexer/common"
	"q/rpc"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	prefix           string
	addr             string
	level            uint32
	parentAddr       map[string]string
	childrenAddr     map[string]string  // domain -> addr
	parentClient  	 rpc.IndexerClient
	childClients     map[string]rpc.IndexerClient // domain -> client
	recvCh           chan *common.ReqWithCh
	respCh           chan *rpc.QueryMsg
	stopChan         chan struct{}
	mutex            sync.Mutex
}

func NewNode(prefix string, addr string, parentStr string, childrenStr []string, recvCh chan *common.ReqWithCh, respCh chan *rpc.QueryMsg) *Node {
	parentAddr := make(map[string]string)
	childrenAddr := make(map[string]string)
	var parentClient rpc.IndexerClient
	childClients := make(map[string]rpc.IndexerClient)

    level := calculateLevel(prefix)
	log.Info().Msgf("level: %d", level)
	
	if parentStr != "" {
		parts := strings.Split(parentStr, "#")
		addr := parts[1]
		parentAddr[parts[0]] = addr

		creds := insecure.NewCredentials()
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to dial")
		}
		parentClient = rpc.NewIndexerClient(conn)
	}
	for _, child := range childrenStr {
		if child != "" {
			parts := strings.Split(child, "#")
			cPrefix := parts[0]
			addr := parts[1]
			childrenAddr[cPrefix] = addr

			creds := insecure.NewCredentials()
			conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to dial")
			}
			childClients[cPrefix] = rpc.NewIndexerClient(conn)
		}
	}
    return &Node{
		prefix : prefix,
		addr: addr,
		level: uint32(level),
		parentAddr: parentAddr,
		childrenAddr: childrenAddr,
		parentClient: parentClient,
		childClients: childClients,
		recvCh: recvCh,
		respCh: respCh,
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

func (nd *Node) handleQuery(req *common.ReqWithCh) {
	switch req.Req.Type {
	case rpc.QueryMsg_FIRST: {
		go nd.handleFirstQuery(req)
	}
	case rpc.QueryMsg_RELAY: {
		go nd.handleRelayQuery(req)
	}
	}
	
}

func (nd *Node) handleFirstQuery(msg *common.ReqWithCh) {
	req := msg.Req
	targetLevel := calculateLevel(req.Prefix)
	if targetLevel == int(nd.level) && req.Prefix == nd.prefix {
        // req.ResponseCh <- &QueryResponse{NodeDomain: nd.domain}
		res := &rpc.QueryMsg{Type: rpc.QueryMsg_FINISH, Prefix: nd.prefix, Addr: nd.addr}
		msg.ResCh <- res
        return
    }

    // ndParts := strings.Split(nd.prefix, "/")
    // targetParts := strings.Split(req.Prefix, "/")
    firstMismatchIndex := calFirstMismatchIndex(req.Prefix, nd.prefix)
	log.Info().Msgf("ywb prefix: %s, req.prefix: %s, firstMismatchIndex: %d", nd.prefix, req.Prefix, firstMismatchIndex)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
    // 根据层级关系选择向父节点或子节点请求
	nextReq := &rpc.QueryMsg{Type: rpc.QueryMsg_RELAY, Prefix: req.Prefix}
	var response *rpc.QueryMsg
    if firstMismatchIndex <= int(nd.level) {
        // 向父节点请求
		if len(nd.parentAddr) > 0 {
			resp, err := nd.parentClient.Send(ctx, nextReq)
			if err != nil {
				log.Error().Msgf("Error sending message: %v", err)
			}
			response = resp
		}
    } else {
        // 向子节点请求
        for prefix, _ := range nd.childrenAddr {
			mismatchIndex := calFirstMismatchIndex(req.Prefix, prefix)
			log.Info().Msgf("child prefix: %s, req.prefix: %s, mismatchIndex: %d", prefix, req.Prefix, mismatchIndex)
			if(mismatchIndex > firstMismatchIndex) {
            	resp, err := nd.childClients[prefix].Send(ctx, nextReq)
				if err != nil {
					log.Error().Msgf("Error sending message: %v", err)
				}
				response = resp
				break;
			}
        } 
    }
	log.Info().Msgf("response: %v", response)
    // req.ResponseCh <- nil

	for {
		switch response.Type {
		case rpc.QueryMsg_FINISH: {
			msg.ResCh <- response
			return
		}
		case rpc.QueryMsg_RELAY: {
			creds := insecure.NewCredentials()
			conn, err := grpc.NewClient(response.Addr, grpc.WithTransportCredentials(creds))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to dial")
			}
			client := rpc.NewIndexerClient(conn)
			log.Info().Msgf("loop send relay to %s", response.Addr)
			response, _ = client.Send(ctx, nextReq)
			log.Info().Msgf("loop relay response %v", response)
		}
		case rpc.QueryMsg_ERROR: {
			response.Addr = "not found error"
			msg.ResCh <- response
			return
		}
		}
	}
}

func (nd *Node) handleRelayQuery(msg *common.ReqWithCh) {
	req := msg.Req
	targetLevel := calculateLevel(req.Prefix)
	if targetLevel == int(nd.level) && req.Prefix == nd.prefix {
        // req.ResponseCh <- &QueryResponse{NodeDomain: nd.domain}
		res := &rpc.QueryMsg{Type: rpc.QueryMsg_FINISH, Prefix: nd.prefix, Addr: nd.addr}
		msg.ResCh <- res
        return
    }

    ndParts := strings.Split(nd.prefix, "/")
    // targetParts := strings.Split(req.Prefix, "/")
    firstMismatchIndex := calFirstMismatchIndex(req.Prefix, nd.prefix)
	log.Info().Msgf("prefix: %s, req.prefix: %s, firstMismatchIndex: %d", nd.prefix, req.Prefix, firstMismatchIndex)

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
    // 根据层级关系选择向父节点或子节点请求
	var response *rpc.QueryMsg
    if firstMismatchIndex < len(ndParts) {
        // 父节点
		if len(nd.parentAddr) > 0 {
			for _, addr := range nd.parentAddr {
				response = &rpc.QueryMsg{Type: rpc.QueryMsg_RELAY, Prefix: req.Prefix, Addr: addr}
			}
		}
    } else {
        // 向子节点请求
        for prefix, _ := range nd.childrenAddr {
			mismatchIndex := calFirstMismatchIndex(req.Prefix, prefix)
			log.Info().Msgf("child prefix: %s, req.prefix: %s, mismatchIndex: %d", prefix, req.Prefix, mismatchIndex)
			if(mismatchIndex > firstMismatchIndex) {
            	// resp, err := nd.childClients[prefix].Send(ctx, req)
				// if err != nil {
				// 	log.Error().Msgf("Error sending message: %v", err)
				// }
				// response = resp
				response = &rpc.QueryMsg{Type: rpc.QueryMsg_RELAY, Prefix: req.Prefix, Addr: nd.childrenAddr[prefix]}
				break;
			}
        } 
    }
	log.Info().Msgf("response: %v", response)
	msg.ResCh <- response
}

func calculateLevel(prefix string) int {
    parts := strings.Split(prefix, "/")
    level := 0
    for _, part := range parts {
        if part != "" {
            level++
        }
    }
    return level
}

func calFirstMismatchIndex(reqPrefix string, ndPrefix string) int {
	ndParts := strings.Split(ndPrefix, "/")
	targetParts := strings.Split(reqPrefix, "/")
	log.Info().Msgf("ndParts: %v, targetParts: %v", ndParts, targetParts)

	// log.Info().Msgf("i: %d, ndParts[%d]: %s, targetParts[%d]: %s", i, i, ndParts[i], i, targetParts[i])

	i, j := 0, 0
    for i < len(ndParts) && j < len(targetParts) {
        if ndParts[i] != targetParts[j] {
            return i
        }
        i++
        j++
    }
    if i < len(ndParts) {
        return i
    }
    return j
	// log.Info().Msgf("ywb firstMismatchIndex: %d", firstMismatchIndex)
	
}