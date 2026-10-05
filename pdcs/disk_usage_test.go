package pdcs_test

import (
	"encoding/json"
	"os/exec"

	"github.com/containers/prometheus-podman-exporter/pdcs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pdcs/DiskUsage", func() {
	It("DiskUsageSummary", func() {
		testImage := "quay.io/quay/busybox"

		_, err := exec.Command("podman", "image", "pull", testImage).Output()
		Expect(err).To(BeNil())

		pdcs.UpdateDiskUsage()

		diskUsage, err := pdcs.DiskUsageSummary()
		Expect(err).To(BeNil())

		output, err := exec.Command("podman", "system", "df", "--format", "json").Output()
		Expect(err).To(BeNil())

		var summaries []struct {
			Type    string
			RawSize int64
		}

		Expect(json.Unmarshal(output, &summaries)).To(Succeed())

		imagesSize := int64(-1)
		for _, summary := range summaries {
			if summary.Type == "Images" {
				imagesSize = summary.RawSize

				break
			}
		}

		Expect(diskUsage.ImagesSize).To(BeNumerically(">", 0))
		Expect(diskUsage.ImagesSize).To(Equal(imagesSize))
	})
})
