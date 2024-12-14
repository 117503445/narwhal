package command

import "github.com/117503445/goutils"

func Exp4RunOnce() {
	goutils.Exec("go build -o exp-map ./cmd/map/main.go", goutils.WithCwd("./assets/exp-mem"))
	goutils.Exec("go build -o exp-bloom ./cmd/bloom/main.go", goutils.WithCwd("./assets/exp-mem"))
	goutils.Exec("go build -o exp-lunwen ./cmd/lunwen/main.go", goutils.WithCwd("./assets/exp-mem"))

	bins := []string{"exp-map", "exp-bloom", "exp-lunwen"}
	for _, bin := range bins {
		
	}
}
