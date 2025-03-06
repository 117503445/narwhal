package command

import (
	"fmt"
	"os"
	"sync"
	"text/template"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

func DeployFC(wg *sync.WaitGroup) {
	var err error
	// 创建 "q/assets/fc-worker/data"
	// if err = os.MkdirAll("../q/assets/fc-worker/data", 0755); err != nil {
	// 	log.Fatal().Err(err).Msg("failed to create dir")
	// }

	tmpl, err := goutils.ReadText("../q/assets/fc-worker/s.yaml.tmpl")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read file")
	}
	t := template.Must(template.New("s").Parse(tmpl))

	for nodeIndex := 0; nodeIndex < 4; nodeIndex++ {
		i := 0

		// 复制 "q/assets/fc-worker/s.yaml" 到 "q/assets/fc-worker/data/nodeIndex_i/s.yaml"
		// src := "../q/assets/fc-worker/s.yaml"
		// dst := fmt.Sprintf("../q/assets/fc-worker/data/%d_%d/s.yaml", nodeIndex, i)

		// if err = goutils.CopyFile(src, dst); err != nil {
		// 	log.Fatal().Err(err).Msg("failed to copy file")
		// }

		dstDir := fmt.Sprintf("../q/assets/fc-worker/data/%d_%d", nodeIndex, i)
		dstDirInDocker := fmt.Sprintf("/workspace/q/assets/fc-worker/data/%d_%d", nodeIndex, i)

		// 创建 q/assets/fc-worker/data/nodeIndex_i
		if err = os.MkdirAll(dstDir, 0755); err != nil {
			log.Fatal().Err(err).Msg("failed to create dir")
		}

		f, err := os.Create(dstDir + "/s.yaml")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create file")
		}
		defer f.Close()

		functionName := fmt.Sprintf("biye-%d-%d", nodeIndex, i)
		err = t.Execute(f, map[string]string{"functionName": functionName, "slaveId": fmt.Sprintf("%d", i)})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to execute template")
		}
		wg.Add(1)
		go func(dstDir string) {
			defer wg.Done()
			cmd := fmt.Sprintf("docker compose exec -T --workdir %v fc s deploy -y", dstDirInDocker)
			goutils.Exec(cmd, goutils.WithCwd("../"))
		}(dstDirInDocker)
	}
}
