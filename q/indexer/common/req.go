package common

import "q/rpc"

type ReqWithCh struct {
	Req *rpc.QueryMsg
	ResCh chan *rpc.QueryMsg
}