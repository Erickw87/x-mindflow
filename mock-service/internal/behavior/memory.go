package behavior

import (
	"runtime"
	"time"

	"github.com/LiusCraft/x-mindflow/mock-service/pkg/logger"
)

type memoryWorker struct {
	allocations [][]byte
	logger      logger.Logger
}

func newMemoryWorker(log logger.Logger) *memoryWorker {
	return &memoryWorker{
		allocations: make([][]byte, 0),
		logger:      log,
	}
}

func (w *memoryWorker) Simulate(targetMB int, duration time.Duration) {
	w.logger.Info("Starting memory simulation",
		"target_mb", targetMB,
		"duration", duration)

	data := make([]byte, targetMB*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	w.allocations = append(w.allocations, data)

	w.logger.Debug("Memory allocated", "size_mb", targetMB)

	time.Sleep(duration)

	w.allocations = w.allocations[:0]
	runtime.GC()

	w.logger.Info("Memory simulation completed and released")
}
