package behavior

import (
	"runtime"
	"time"

	"github.com/LiusCraft/x-mindflow/mock-service/pkg/logger"
)

type cpuWorker struct {
	logger logger.Logger
}

func newCPUWorker(log logger.Logger) *cpuWorker {
	return &cpuWorker{
		logger: log,
	}
}

func (w *cpuWorker) Simulate(targetPercent int, duration time.Duration) {
	numCPU := runtime.NumCPU()
	w.logger.Info("Starting CPU simulation",
		"num_cpu", numCPU,
		"target_percent", targetPercent,
		"duration", duration)

	done := make(chan struct{})

	for i := 0; i < numCPU; i++ {
		go func(core int) {
			start := time.Now()
			cycleTime := 100 * time.Millisecond
			workTime := time.Duration(targetPercent) * time.Millisecond
			sleepTime := cycleTime - workTime

			w.logger.Debug("CPU worker started", "core", core)

			for time.Since(start) < duration {
				workEnd := time.Now().Add(workTime)
				for time.Now().Before(workEnd) {
					for j := 0; j < 1000000; j++ {
						_ = j * j
					}
				}
				time.Sleep(sleepTime)
			}

			done <- struct{}{}
		}(i)
	}

	go func() {
		for i := 0; i < numCPU; i++ {
			<-done
		}
		w.logger.Info("CPU simulation completed")
	}()
}
