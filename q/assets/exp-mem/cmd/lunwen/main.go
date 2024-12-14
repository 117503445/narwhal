package main

import "github.com/117503445/goutils"

func main() {
	goutils.InitZeroLog()
	h1 := make(map[string]interface{})
	for i := 0; i < 50000; i++ {
		h1[goutils.UUID4()] = nil
	}
	// sleep forever
	select {}

}
