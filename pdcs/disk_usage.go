package pdcs

import (
	"log/slog"
	"sync"
	"time"

	"github.com/containers/podman/v5/cmd/podman/registry"
	"github.com/containers/podman/v5/pkg/domain/entities"
)

var diskUsageRep diskUsageReport

type diskUsageReport struct {
	diskUsage DiskUsage
	updateErr error
	repLock   sync.Mutex
}

// DiskUsage implements podman disk usage (podman system df) summary information.
type DiskUsage struct {
	ImagesSize int64
}

// DiskUsageSummary returns cached podman disk usage summary (DiskUsage).
func DiskUsageSummary() (DiskUsage, error) {
	diskUsageRep.repLock.Lock()
	defer diskUsageRep.repLock.Unlock()

	if diskUsageRep.updateErr != nil {
		return DiskUsage{}, diskUsageRep.updateErr
	}

	return diskUsageRep.diskUsage, nil
}

func updateDiskUsage() {
	report, err := registry.ContainerEngine().SystemDf(registry.Context(), entities.SystemDfOptions{})

	diskUsageRep.repLock.Lock()
	defer diskUsageRep.repLock.Unlock()

	if err != nil {
		diskUsageRep.updateErr = err
		diskUsageRep.diskUsage = DiskUsage{}

		return
	}

	diskUsageRep.updateErr = nil
	diskUsageRep.diskUsage = DiskUsage{
		ImagesSize: report.ImagesSize,
	}
}

// StartDiskUsageTicker starts disk usage cache refresh routine.
func StartDiskUsageTicker(logger *slog.Logger, duration int64) {
	logger.Info("starting disk usage cache ticker", "duration", duration)
	logger.Info("update disk usage cache")

	updateDiskUsage()

	ticker := time.NewTicker(time.Duration(duration) * time.Second)

	go func() {
		for {
			<-ticker.C
			logger.Info("update disk usage cache")
			updateDiskUsage()
		}
	}()
}
