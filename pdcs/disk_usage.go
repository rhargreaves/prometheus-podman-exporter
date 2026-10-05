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
	ImagesSize            int64
	ImagesReclaimable     int64
	ContainersSize        int64
	ContainersReclaimable int64
	VolumesSize           int64
	VolumesReclaimable    int64
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
	diskUsageRep.diskUsage = diskUsageSummary(report)
}

// diskUsageSummary calculates the summary in the same way as the podman system df command.
// Following code is based on printSummary from https://github.com/containers/podman/blob/v5.4.2/cmd/podman/system/df.go
func diskUsageSummary(report *entities.SystemDfReport) DiskUsage {
	diskUsage := DiskUsage{
		ImagesSize: report.ImagesSize,
	}

	var imagesUsed int64

	// an image can have multiple tags, count each image only once.
	visitedImages := make(map[string]bool)

	for _, image := range report.Images {
		if visitedImages[image.ImageID] {
			continue
		}

		visitedImages[image.ImageID] = true

		if image.Containers > 0 {
			imagesUsed += image.UniqueSize
		}
	}

	diskUsage.ImagesReclaimable = report.ImagesSize - imagesUsed

	for _, container := range report.Containers {
		if container.Status != "running" {
			diskUsage.ContainersReclaimable += container.RWSize
		}

		diskUsage.ContainersSize += container.RWSize
	}

	for _, volume := range report.Volumes {
		diskUsage.VolumesSize += volume.Size
		diskUsage.VolumesReclaimable += volume.ReclaimableSize
	}

	return diskUsage
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
