package command

type ExpCMD struct {
}

func runExp1() {
	// pressList := []int{1000, 1000000}
	pressList := []int{1000000}
	nList := []int{4, 8, 16, 32, 64}
	// nList := []int{64}
	// modeList := []string{ "p2p"}
	modeList := []string{"p2p", "broadcast"}
	// mode := "p2p"
	latencyMock := true

	for _, press := range pressList {
		for _, mode := range modeList {
			for _, n := range nList {
				Exp1RunOnce(&Exp1Param{
					Press: press,
					// Mode:  "p2p",
					Mode:        mode,
					N:           n,
					LatencyMock: latencyMock,
				})
			}
		}
	}
}

func runExp2() {
	pressList := []int{1000000}
	// pressList := []int{1000, 1000000}
	nList := []int{32, 64}
	// nList := []int{4, 8, 16, 32, 64}
	eciNumList := []int{2, 3}

	for _, press := range pressList {
		for _, n := range nList {
			for _, eciNum := range eciNumList {
				Exp2RunOnce(&Exp2Param{
					Press:       press,
					N:           n,
					EciNum:      eciNum,
					LatencyMock: false,
				})
			}
		}
	}
}

func runExp3() {
	pressList := []int{1000}
	nList := []int{64}

	for _, press := range pressList {
		for _, n := range nList {
			Exp3RunOnce(&Exp3Param{
				Press:       press,
				N:           n,
				LatencyMock: false,
			})
		}
	}
}

func (cmd *ExpCMD) Run() error {
	runExp2()
	return nil
}
