package cdrom

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/structs"
)

func (plugin *Plugin) fingerprinter(ctx context.Context, updates chan<- *device.FingerprintResponse) {
	defer close(updates)

	// Create a timer that will fire immediately for the first detection
	ticker := time.NewTimer(0)

	for {
		select {
		case <-ctx.Done():
			if !ticker.Stop() {
				<-ticker.C
			}

			return
		case <-ticker.C:
			plugin.fingerprint(updates)
			ticker.Reset(plugin.FingerprintInterval)
		}
	}
}

func (plugin *Plugin) fingerprint(updates chan<- *device.FingerprintResponse) {
	plugin.Debug("fingerprinting devices", plugin.InfoPath)
	info, err := os.Open(plugin.InfoPath)
	if err != nil {
		updates <- &device.FingerprintResponse{Error: err}
		return
	}

	defer info.Close()

	drives, fingerprint, err := LoadTable(info)
	if err != nil {
		updates <- &device.FingerprintResponse{Error: err}
		return
	}

	plugin.mu.Lock()
	defer plugin.mu.Unlock()

	if fingerprint == plugin.state {
		// Don't send an update if devices haven't changed
		plugin.Debug("cached devices are up to date", "fingerprint", fingerprint)
		return
	}

	plugin.state = fingerprint
	plugin.cache = make(map[string]Device)

	plugin.Info("updating devices", "fingerprint", fingerprint)
	var groups []*device.DeviceGroup

	for _, dev := range drives {
		plugin.Info("using CDROM device", "dev", dev.ID)

		// Create a read/write group with one device
		plugin.cache[dev.ID] = Device{
			ID:   dev.ID,
			Path: filepath.Join("/dev", dev.ID),
			Perm: "rw",
		}

		groups = append(groups, &device.DeviceGroup{
			Vendor:  "generic",
			Type:    "cdrom",
			Name:    "read_write",
			Devices: []*device.Device{{ID: dev.ID, Healthy: true}},
			Attributes: map[string]*structs.Attribute{
				"can_write_cdr":  structs.NewBoolAttribute(dev.CanWriteCDR),
				"can_write_cdrw": structs.NewBoolAttribute(dev.CanWriteCDRW),
				"can_read_dvd":   structs.NewBoolAttribute(dev.CanReadDVD),
				"can_write_dvdr": structs.NewBoolAttribute(dev.CanWriteDVDR),
			},
		})

		// Create a read-only group that offers multiple devices if configured
		if plugin.ReadonlySeats > 0 {
			group := &device.DeviceGroup{
				Vendor: "generic",
				Type:   "cdrom",
				Name:   "readonly",
				Attributes: map[string]*structs.Attribute{
					"can_read_dvd": structs.NewBoolAttribute(dev.CanReadDVD),
				},
			}

			for i := uint8(0); i < plugin.ReadonlySeats; i++ {
				id := fmt.Sprintf("%s.%d", dev.ID, i)

				plugin.cache[id] = Device{
					ID:   id,
					Path: filepath.Join("/dev", dev.ID),
					Perm: "ro",
				}
				group.Devices = append(group.Devices, &device.Device{ID: id, Healthy: true})
			}

			groups = append(groups, group)
		}
	}

	updates <- &device.FingerprintResponse{Devices: groups}
}
