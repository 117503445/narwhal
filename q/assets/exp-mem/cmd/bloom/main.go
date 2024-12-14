package main

import "github.com/117503445/goutils"

func main() {
	goutils.InitZeroLog()

	filter := bloom.NewWithEstimates(1000000, 0.01) 

	println("Hello World!")
}
