package collector

import (
	"log/slog"

	"github.com/containers/prometheus-podman-exporter/pdcs"
	"github.com/prometheus/client_golang/prometheus"
)

type diskUsageCollector struct {
	imagesSize typedDesc
	logger     *slog.Logger
}

func init() {
	registerCollector("disk_usage", defaultDisabled, NewDiskUsageCollector)
}

// NewDiskUsageCollector returns a Collector exposing podman disk usage information.
func NewDiskUsageCollector(logger *slog.Logger) (Collector, error) {
	return &diskUsageCollector{
		imagesSize: typedDesc{
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disk_usage", "images_size_bytes"),
				"Podman disk usage of images (sum of all image layers).",
				nil, nil,
			), prometheus.GaugeValue,
		},
		logger: logger,
	}, nil
}

// Update reads and exposes podman disk usage information.
func (c *diskUsageCollector) Update(ch chan<- prometheus.Metric) error {
	diskUsage, err := pdcs.DiskUsageSummary()
	if err != nil {
		return err
	}

	ch <- c.imagesSize.mustNewConstMetric(float64(diskUsage.ImagesSize))

	return nil
}
