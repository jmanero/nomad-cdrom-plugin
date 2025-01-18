package main

import (
	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/device"

	"github.com/jmanero/nomad-cdrom-plugin/cdrom"
)

// Version build-time constant
var Version = "undefined"

func main() {
	plugins.Serve(factory)
}

func factory(log log.Logger) interface{} {
	return cdrom.NewPlugin(log, &base.PluginInfoResponse{
		Type:              base.PluginTypeDevice,
		PluginApiVersions: []string{device.ApiVersion010},
		PluginVersion:     Version,
		Name:              "cdrom",
	})
}
