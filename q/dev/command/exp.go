package command

type ExpCMD struct {
}

func runExp1() {
	pressList := []int{1000, 1000000}
	// pressList := []int{1000000}
	nList := []int{4, 8, 16, 32, 64}
	// nList := []int{64}
	// modeList := []string{"p2p", "broadcast"}
	modeList := []string{"broadcast"}
	latencyMock := false

	ckpNList := []int{0, 1, 4, 10, 20}

	// test ckpn
	// pressList = []int{1000}
	// nList = []int{4}
	// ckpNList = []int{1, 2, 10, 50}
	// modeList = []string{"broadcast"}

	for _, press := range pressList {
		for _, ckpN := range ckpNList {
			for _, mode := range modeList {
				for _, n := range nList {
					Exp1RunOnce(&Exp1Param{
						Press:       press,
						Mode:        mode,
						N:           n,
						LatencyMock: latencyMock,
						CkpN:        ckpN,
					})
				}
			}
		}
	}
}

func runExp2() {
	pressList := []int{1000000}
	// pressList := []int{1000, 100000}
	nList := []int{4, 8, 16}
	// nList := []int{4, 8, 16, 32, 64}
	// eciNumList := []int{4}
	eciNumList := []int{2, 3}

	for _, eciNum := range eciNumList {
		for _, press := range pressList {
			for _, n := range nList {
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
	// pressList := []int{1000, 1000000}
	pressList := []int{ 1000000}

	// bandwidthList := []float64{10, 12.5, 25, 50, 100}
	bandwidthList := []float64{12.5, 25, 50, 100}
	// bandwidthList := []float64{ 25, 50, 100}
	// bandwidthList := []float64{ 80}

	// pressList := []int{1000,100000}
	nList := []int{4, 8, 16, 32, 64}
	// nList := []int{64}

	for _, press := range pressList {
		for _, bandwidth := range bandwidthList {
			for _, n := range nList {
				if bandwidth == 12.5 {
					continue
				}

				Exp3RunOnce(&Exp3Param{
					Press:       press,
					N:           n,
					LatencyMock: false,
					Bandwidth:   bandwidth,
				})
			}
		}
	}
}

func runExp4() {
	Exp4RunOnce()
}

func (cmd *ExpCMD) Run() error {
	runExp3()
	return nil
}
