package device

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/structs"
)

// CheckDrive is a helper to check if media is present in an optical drive
func CheckDrive(path string) bool {
	dev, err := os.OpenFile(path, 0, 0)
	if err != nil {
		return false
	}

	defer dev.Close()

	var buf [64]byte
	_, err = dev.Read(buf[:])
	if err != nil {
		return false
	}

	return true
}

// doStats is the long running goroutine that streams device statistics
func (plugin *Plugin) stater(ctx context.Context, stats chan<- *device.StatsResponse, interval time.Duration) {
	defer close(stats)

	// Create a timer that will fire immediately for the first detection
	ticker := time.NewTimer(0)

	for {
		select {
		case <-ctx.Done():
			if !ticker.Stop() {
				<-ticker.C
			}

			return
		case now := <-ticker.C:
			plugin.stats(stats, now)
			ticker.Reset(interval)
		}
	}
}

func (plugin *Plugin) stats(stats chan<- *device.StatsResponse, timestamp time.Time) {
	plugin.Debug("collecting statistics")

	plugin.mu.RLock()
	defer plugin.mu.RUnlock()

	instances := make(map[string]*device.DeviceStats)

	for _, drive := range plugin.cache {
		plugin.Debug("checking for media presence", drive.ID, drive.Path)
		loaded := CheckDrive(drive.Path)

		instances[drive.ID] = &device.DeviceStats{
			Summary: &structs.StatValue{
				Desc:    "Check if media is present in the optical drive",
				BoolVal: &loaded,
			},
			Timestamp: timestamp,
		}
	}

	stats <- &device.StatsResponse{
		Groups: []*device.DeviceGroupStats{
			{Type: "cdrom", Vendor: "generic", Name: "cdrom", InstanceStats: instances},
		},
	}
}
