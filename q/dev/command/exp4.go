package command

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v4/process"
)

func Exp4RunOnce() {
	goutils.Exec("go build -o exp-map ./cmd/map/main.go", goutils.WithCwd("./assets/exp-mem"))
	goutils.Exec("go build -o exp-bloom ./cmd/bloom/main.go", goutils.WithCwd("./assets/exp-mem"))
	goutils.Exec("go build -o exp-lunwen ./cmd/lunwen/main.go", goutils.WithCwd("./assets/exp-mem"))

	bins := []string{"exp-bloom", "exp-lunwen", "exp-map"}

	press := 100000

	type record struct {
		Seconds  int
		RssBytes int
		Bin      string
	}
	result := make([]record, 0)

	for _, bin := range bins {
		cmd := exec.Command(fmt.Sprintf("./assets/exp-mem/%s", bin))
		cmd.Env = append(cmd.Env, fmt.Sprintf("PRESS=%d", press))

		// Start the command but do not wait for it to complete.
		if err := cmd.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start command")
		}

		// Get the PID of the started subprocess.
		pid := cmd.Process.Pid
		fmt.Printf("Started subprocess with PID: %d\n", pid)

		for i := 0; i < 3600 * 4; i++ {
			// Give some time for the subprocess to fully start up.
			time.Sleep(1 * time.Second) // 可能需要根据实际情况调整这个时间

			// Try to get memory information for the subprocess.
			memInfo, err := getProcessMemoryInfo(int32(pid))
			if err != nil {
				log.Fatal().Msg("Failed to get memory info for subprocess")
			}

			// Print out the memory information.
			fmt.Printf("%d - Subprocess Memory Info - RSS (Resident Set Size): %d bytes\n", i, memInfo.RSS)
			// fmt.Printf("Subprocess Memory Info - VMS (Virtual Memory Size): %d bytes\n", memInfo.VMS)
			result = append(result, record{
				Seconds:  i,
				RssBytes: int(memInfo.RSS),
				Bin:      bin,
			})
			if i%10 == 0 {
				err = goutils.WriteJSON(fmt.Sprintf("../paper-exp-data/exp4-%s.json", expID), result)
				if err != nil {
					log.Error().Err(err).Msg("WriteJSON")
				}
			}
		}

	}
}

func getProcessMemoryInfo(pid int32) (*process.MemoryInfoStat, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to create process instance: %v", err)
	}

	memInfo, err := p.MemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %v", err)
	}

	return memInfo, nil
}
