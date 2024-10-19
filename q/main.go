package main

import (
	"github.com/117503445/goutils"
	"github.com/alecthomas/kong"
	dev "q/dev/command"
	executor "q/executor/command"
	indexer "q/indexer/command"
	indexerClient "q/indexer-client/command"
)

type DefaultCmd struct {
}

func (d *DefaultCmd) Run() error {
	return new(dev.BuildCmd).Run()
}

var cli struct {
	DefaultCmd DefaultCmd           `cmd:"" hidden:"" default:"1"`
	Build      dev.BuildCmd         `cmd:""`
	Req        dev.ReqCMD           `cmd:""`
	IndexReq   dev.IndexReqCMD      `cmd:""`
	Dev0       dev.Dev0CMD          `cmd:""`
	Executor   executor.ExecutorCmd `cmd:""`
	Indexer	   indexer.IndexerCmd   `cmd:""` 
	IndexerClient indexerClient.IndexerClientCmd `cmd:""`
}

func main() {
	goutils.InitZeroLog()

	goutils.ExecOpt.DumpOutput = true

	ctx := kong.Parse(&cli)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
