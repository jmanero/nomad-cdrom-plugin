package cdrom

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ConfigSpec is the specification of the schema for this plugin's config.
	// this is used to validate the HCL for the plugin provided
	// as part of the client config:
	//   https://www.nomadproject.io/docs/configuration/plugin.html
	// options are here:
	//   https://github.com/hashicorp/nomad/blob/v0.10.0/plugins/shared/hclspec/hcl_spec.proto
	ConfigSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		"fingerprint_interval": hclspec.NewDefault(
			hclspec.NewAttr("fingerprint_interval", "string", false),
			hclspec.NewLiteral(`"5m"`),
		),
		"cdrom_info_path": hclspec.NewDefault(
			hclspec.NewAttr("cdrom_info_path", "string", false),
			hclspec.NewLiteral(`"/proc/sys/dev/cdrom/info"`),
		),
		"readonly_seats": hclspec.NewDefault(
			hclspec.NewAttr("readonly_seats", "number", false),
			hclspec.NewLiteral("0"),
		),
		"default_vendor": hclspec.NewDefault(
			hclspec.NewAttr("default_vendor", "string", false),
			hclspec.NewLiteral(`"generic"`),
		),
	})
)

// Config contains configuration information for the plugin.
type Config struct {
	FingerprintInterval string `codec:"fingerprint_interval"`
	InfoPath            string `codec:"cdrom_info_path"`
	ReadonlySeats       uint8  `codec:"readonly_seats"`
	DefaultVendor       string `codec:"default_vendor"`
}

// Device describes an available device/seat
type Device struct {
	ID   string
	Path string
	Perm string
}

// Plugin fingerprints optical media devices available on a linux host
type Plugin struct {
	log.Logger

	// Configured
	FingerprintInterval time.Duration
	InfoPath            string
	ReadonlySeats       uint8
	DefaultVendor       string

	info  *base.PluginInfoResponse
	state string
	cache map[string]Device
	mu    sync.RWMutex
}

// NewPlugin returns a device plugin, used primarily by the main wrapper
//
// Plugin configuration isn't available yet, so there will typically be
// a limit to the initialization that can be performed at this point.
func NewPlugin(log log.Logger, info *base.PluginInfoResponse) *Plugin {
	return &Plugin{
		Logger: log.Named("cdrom"),
		info:   info,
		cache:  make(map[string]Device),
	}
}

// PluginInfo returns information describing the plugin.
//
// This is called during Nomad client startup, while discovering and loading
// plugins.
func (plugin *Plugin) PluginInfo() (*base.PluginInfoResponse, error) {
	return plugin.info, nil
}

// ConfigSchema returns the configuration schema for the plugin.
//
// This is called during Nomad client startup, immediately before parsing
// plugin config and calling SetConfig
func (*Plugin) ConfigSchema() (*hclspec.Spec, error) {
	return ConfigSpec, nil
}

// SetConfig is called by the client to pass the configuration for the plugin.
func (plugin *Plugin) SetConfig(c *base.Config) error {
	var config Config

	// decode the plugin config
	if err := base.MsgPackDecode(c.PluginConfig, &config); err != nil {
		return err
	}

	// for example, convert the poll period from an HCL string into a time.Duration
	interval, err := time.ParseDuration(config.FingerprintInterval)
	if err != nil {
		return fmt.Errorf("failed to parse fingerprint_interval %q: %w", config.FingerprintInterval, err)
	}

	plugin.FingerprintInterval = interval

	_, err = os.Stat(config.InfoPath)
	if err != nil {
		return fmt.Errorf("unable to read cdrom_info_path at %q: %w", config.FingerprintInterval, err)
	}

	plugin.InfoPath = config.InfoPath

	if config.ReadonlySeats == 0 {
		plugin.Warn("no readonly devices will be allocated")
	}

	plugin.ReadonlySeats = config.ReadonlySeats
	plugin.DefaultVendor = config.DefaultVendor

	plugin.Info("plugin configured", "fingerprint_interval", interval, "cdrom_info_path", config.InfoPath, "readonly_seats", config.ReadonlySeats)
	return nil
}

// Fingerprint streams detected devices
func (plugin *Plugin) Fingerprint(ctx context.Context) (<-chan *device.FingerprintResponse, error) {
	updates := make(chan *device.FingerprintResponse)

	go plugin.fingerprinter(ctx, updates)
	return updates, nil
}

// Stats streams statistics for the detected devices
func (plugin *Plugin) Stats(ctx context.Context, interval time.Duration) (<-chan *device.StatsResponse, error) {
	stats := make(chan *device.StatsResponse)

	go plugin.stater(ctx, stats, interval)
	return stats, nil
}

// Reserve returns information to the task driver on on how to mount the given devices.
// It may also perform any device-specific orchestration necessary to prepare the device
// for use. This is called in a pre-start hook on the client, before starting the workload.
func (plugin *Plugin) Reserve(ids []string) (*device.ContainerReservation, error) {
	if len(ids) == 0 {
		return &device.ContainerReservation{}, nil
	}

	plugin.mu.RLock()
	defer plugin.mu.RUnlock()

	var missing, matched []string

	reservation := &device.ContainerReservation{
		Envs:    map[string]string{},
		Devices: []*device.DeviceSpec{},
	}

	for _, id := range ids {
		// Check if the device is known
		entry, exist := plugin.cache[id]

		if !exist {
			missing = append(missing, id)
			continue
		}

		matched = append(matched, entry.Path)

		// Devices are the set of devices to mount into the container.
		reservation.Devices = append(reservation.Devices, &device.DeviceSpec{
			HostPath:    entry.Path,
			TaskPath:    entry.Path,
			CgroupPerms: entry.Perm,
		})
	}

	if len(missing) != 0 {
		return nil, status.Newf(codes.InvalidArgument, "unknown devices: %v", missing).Err()
	}

	// Tell the task which devices have been provided
	reservation.Envs["NOMAD_CDROM_DEVICES"] = strings.Join(matched, string(filepath.ListSeparator))

	return reservation, nil
}
