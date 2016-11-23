package main

import (
	// Import the snap plugin library
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	// Import our collector plugin implementation
	. "github.com/intelsdi-x/snap-plugin-utilities/logger"
	"github.com/raintank/snap-plugin-collector-kubestate/kubestate"
)

const (
	pluginName    = "kubestate"
	pluginVersion = 1
)

func main() {
	LogDebug("Starting kubestate collector")

	plugin.StartCollector(new(kubestate.Kubestate), pluginName, pluginVersion, plugin.ConcurrencyCount(1000))
}
