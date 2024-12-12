package main

import (
	"math"
	"math/rand/v2"
	"time"
)

// 生成正态分布随机数
func normalDistribution(mean, stddev float64) float64 {
	// 使用 Box-Muller 变换生成标准正态分布的随机数
	u1 := rand.Float64()
	u2 := rand.Float64()
	z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
	return z0*stddev + mean
}

func sleepForLatencyMock() {
	// 模拟延迟
	mean := 40.0
	stddev := 2.5
	latency := normalDistribution(mean, stddev)
	if latency < 0 {
		latency = 0
	}
	time.Sleep(time.Duration(latency) * time.Millisecond)
}
