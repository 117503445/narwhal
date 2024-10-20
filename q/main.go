package main

import (
	"github.com/117503445/goutils"
	"github.com/alecthomas/kong"
	dev "q/dev/command"
	executor "q/executor/command"
	sendreq "q/sendreq/command"
	worker "q/worker-master/command"
	workerSlaveClient "q/worker-slave-client/command"
)

type DefaultCmd struct {
}

func (d *DefaultCmd) Run() error {
	return new(dev.BuildCmd).Run()
}

var cli struct {
	DefaultCmd        DefaultCmd                             `cmd:"" hidden:"" default:"1"`
	Build             dev.BuildCmd                           `cmd:""`
	Req               dev.ReqCMD                             `cmd:"" help:"Call SendReq in Docker"`
	Dev0              dev.Dev0CMD                            `cmd:""`
	Executor          executor.ExecutorCmd                   `cmd:""`
	Worker            worker.WorkerCmd                       `cmd:""`
	SendReq           sendreq.SendReqCmd                     `cmd:"" help:"send a request to the worker"`
	WorkerSlaveClient workerSlaveClient.WorkerSlaveClientCmd `cmd:""`
}

func main() {
	goutils.InitZeroLog()

	goutils.ExecOpt.DumpOutput = true

	ctx := kong.Parse(&cli)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
