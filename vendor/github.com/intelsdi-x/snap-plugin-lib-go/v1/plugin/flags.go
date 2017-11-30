package plugin

import (
	"fmt"
	"path/filepath"

	"github.com/urfave/cli"
)

var (
	flConfig = cli.StringFlag{
		Name:  "config",
		Usage: "config to use in JSON format",
	}
	// If no port was provided we let the OS select a port for us.
	// This is safe as address is returned in the Response and keep
	// alive prevents unattended plugins.
	flPort = cli.StringFlag{
		Name:  "port",
		Usage: "port GRPC will listen on",
	}
	// ListenAddr the address that GRPC will listen on.  Plugin authors can also
	// use this address if their plugin binds to a local port as it's sometimes
	// needed to bind to a public interface.
	ListenAddr = "127.0.0.1"
	flAddr     = cli.StringFlag{
		Name:        "addr",
		Usage:       "addr GRPC will listen on",
		Value:       ListenAddr,
		Destination: &ListenAddr,
	}
	LogLevel   = 2
	flLogLevel = cli.IntFlag{
		Name:        "log-level",
		Usage:       "log level - 0:panic 1:fatal 2:error 3:warn 4:info 5:debug",
		Value:       LogLevel,
		Destination: &LogLevel,
	}
	flPprof = cli.BoolFlag{
		Name:  "pprof",
		Usage: "enable pprof",
	}
	flTLS = cli.BoolFlag{
		Name:  "tls",
		Usage: "enable TLS",
	}
	flCertPath = cli.StringFlag{
		Name:  "cert-path",
		Usage: "necessary to provide when TLS enabled",
	}
	flKeyPath = cli.StringFlag{
		Name:  "key-path",
		Usage: "necessary to provide when TLS enabled",
	}
	flRootCertPaths = cli.StringFlag{
		Name:  "root-cert-paths",
		Usage: fmt.Sprintf("root paths separated by '%c'", filepath.ListSeparator),
	}
	flStandAlone = cli.BoolFlag{
		Name:  "stand-alone",
		Usage: "enable stand alone plugin",
	}
	flHTTPPort = cli.IntFlag{
		Name:  "stand-alone-port",
		Usage: "specify http port when stand-alone is set",
		Value: 8182,
	}
	collectDurationStr   = "5s"
	flMaxCollectDuration = cli.StringFlag{
		Name:        "max-collect-duration",
		Usage:       "sets the maximum duration (always greater than 0s) between collections before metrics are sent. Defaults to 10s what means that after 10 seconds no new metrics are received, the plugin should send whatever data it has in the buffer instead of waiting longer. (e.g. 5s)",
		Value:       collectDurationStr,
		Destination: &collectDurationStr,
	}

	flMaxMetricsBuffer = cli.Int64Flag{
		Name:  "max-metrics-buffer",
		Usage: "maximum number of metrics the plugin is buffering before sending metrics. Defaults to zero what means send metrics immediately.",
	}
)
