/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"text/tabwriter"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	log "github.com/sirupsen/logrus"
)

var (
	app     *cli.App
	appArgs struct {
		plugin  Plugin
		name    string
		version int
		opts    []MetaOpt
	}
	// Flags required by the plugin lib flags - plugin authors can provide their
	// own flags.  Checkout https://github.com/intelsdi-x/snap-plugin-lib-go/blob/master/examples/snap-plugin-collector-rand/rand/rand.go
	// for an example of a plugin adding a custom flag.
	Flags []cli.Flag = []cli.Flag{
		flConfig,
		flAddr,
		flPort,
		flPprof,
		flTLS,
		flCertPath,
		flKeyPath,
		flRootCertPaths,
		flStandAlone,
		flHTTPPort,
		flLogLevel,
		flMaxCollectDuration,
		flMaxMetricsBuffer,
	}
)

// Plugin is the base plugin type. All plugins must implement GetConfigPolicy.
type Plugin interface {
	GetConfigPolicy() (ConfigPolicy, error)
}

// Collector is a plugin which is the source of new data in the Snap pipeline.
type Collector interface {
	Plugin

	GetMetricTypes(Config) ([]Metric, error)
	CollectMetrics([]Metric) ([]Metric, error)
}

// Processor is a plugin which filters, aggregates, or decorates data in the
// Snap pipeline.
type Processor interface {
	Plugin

	Process([]Metric, Config) ([]Metric, error)
}

// Publisher is a sink in the Snap pipeline.  It publishes data into another
// System, completing a Workflow path.
type Publisher interface {
	Plugin

	Publish([]Metric, Config) error
}

var App *cli.App

/*
 StreamCollector is a Collector that can send back metrics within configurable limits defined in task manifest.
 These limits might be determined by user by set a value of:
  - `max-metrics-buffer`, default to 0 what means no buffering and sending reply with streaming metrics immediately
  - `max-collect-duration`, default to 10s what means after 10s no new metrics are received, send a reply whatever data it has
  in buffer instead of waiting longer
*/
type StreamCollector interface {
	Plugin

	// StreamMetrics allows the plugin to send/receive metrics on a channel
	// Arguments are (in order):
	//
	// A channel for metrics into the plugin from Snap -- which
	// are the metric types snap is requesting the plugin to collect.
	//
	// A channel for metrics from the plugin to Snap -- the actual
	// collected metrics from the plugin.
	//
	// A channel for error strings that the library will report to snap
	// as task errors.
	StreamMetrics(context.Context, chan []Metric, chan []Metric, chan string) error
	GetMetricTypes(Config) ([]Metric, error)
}

var getOSArgs = func() []string { return os.Args }

// tlsServerSetup offers functions supporting TLS server setup
type tlsServerSetup interface {
	// makeTLSConfig delivers TLS config suitable to use for plugins, excluding
	// setup of certificates (either subject or root CA certificates).
	makeTLSConfig() *tls.Config
	// readRootCAs is a function that delivers root CA certificates for the purpose
	// of TLS initialization
	readRootCAs(rootCertPaths string) (*x509.CertPool, error)
	// updateServerOptions configures any additional options for GRPC server
	updateServerOptions(options ...grpc.ServerOption) []grpc.ServerOption
}

// osInputOutput supports interactions with OS for the plugin lib
type OSInputOutput interface {
	// readOSArgs gets command line arguments passed to application
	readOSArg() string
	// printOut outputs given data to application standard output
	printOut(data string)

	args() int

	setContext(c *cli.Context)
}

// standardInputOutput delivers standard implementation for OS
// interactions
type standardInputOutput struct {
	context *cli.Context
}

// libInputOutput holds utility used for OS interactions
var libInputOutput OSInputOutput = &standardInputOutput{}

// readOSArgs implementation that returns application args passed by OS
func (io *standardInputOutput) readOSArg() string {
	if io.context != nil {
		return io.context.Args().First()
	}
	if len(os.Args) > 0 {
		return os.Args[0]
	}
	return ""
}

// printOut implementation that emits data into standard output
func (io *standardInputOutput) printOut(data string) {
	fmt.Println(data)
}

func (io *standardInputOutput) setContext(c *cli.Context) {
	io.context = c
}

func (io *standardInputOutput) args() int {
	if io.context != nil {
		return io.context.NArg()
	}
	return 0
}

// tlsServerDefaultSetup provides default implementation for TLS setup routines
type tlsServerDefaultSetup struct {
}

// tlsSetup holds TLS setup utility for plugin lib
var tlsSetup tlsServerSetup = tlsServerDefaultSetup{}

// makeTLSConfig provides TLS configuration template for plugins, setting
// required verification of client cert and preferred server suites.
func (ts tlsServerDefaultSetup) makeTLSConfig() *tls.Config {
	config := tls.Config{
		ClientAuth:               tls.RequireAndVerifyClientCert,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
	}
	return &config
}

// readRootCAs delivers a standard source of root CAs from system
func (ts tlsServerDefaultSetup) readRootCAs(rootCertPaths string) (*x509.CertPool, error) {
	if rootCertPaths == "" {
		return x509.SystemCertPool()
	}
	certPaths := filepath.SplitList(rootCertPaths)
	return ts.loadRootCerts(certPaths)
}

// updateServerOptions a standard implementation delivers no additional options
func (ts tlsServerDefaultSetup) updateServerOptions(options ...grpc.ServerOption) []grpc.ServerOption {
	return options
}

func (ts tlsServerDefaultSetup) loadRootCerts(certPaths []string) (rootCAs *x509.CertPool, err error) {
	var path string
	var filepaths []string
	// list potential certificate files
	for _, path := range certPaths {
		var stat os.FileInfo
		if stat, err = os.Stat(path); err != nil {
			return nil, fmt.Errorf("unable to process CA cert source path %s: %v", path, err)
		}
		if !stat.IsDir() {
			filepaths = append(filepaths, path)
			continue
		}
		var subfiles []os.FileInfo
		if subfiles, err = ioutil.ReadDir(path); err != nil {
			return nil, fmt.Errorf("unable to process CA cert source directory %s: %v", path, err)
		}
		for _, subfile := range subfiles {
			subpath := filepath.Join(path, subfile.Name())
			if subfile.IsDir() {
				log.WithField("path", subpath).Debug("Skipping second level directory found among certificate files")
				continue
			}
			filepaths = append(filepaths, subpath)
		}
	}
	rootCAs = x509.NewCertPool()
	numread := 0
	for _, path = range filepaths {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.WithFields(log.Fields{"path": path, "error": err}).Debug("Unable to read cert file")
			continue
		}
		if !rootCAs.AppendCertsFromPEM(b) {
			log.WithField("path", path).Debug("Didn't find any usable certificates in cert file")
			continue
		}
		numread++
	}
	if numread == 0 {
		return nil, fmt.Errorf("found no usable certificates in given locations")
	}
	return rootCAs, nil
}

// makeGRPCCredentials delivers credentials object suitable for setting up gRPC
// server, with TLS optionally turned on.
func makeGRPCCredentials(m *meta) (creds credentials.TransportCredentials, err error) {
	var config *tls.Config
	if !m.TLSEnabled {
		config = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		cert, err := tls.LoadX509KeyPair(m.CertPath, m.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to setup credentials for plugin - loading key pair failed: %v", err.Error())
		}
		config = tlsSetup.makeTLSConfig()
		config.Certificates = []tls.Certificate{cert}
		if config.ClientCAs, err = tlsSetup.readRootCAs(m.RootCertPaths); err != nil {
			return nil, fmt.Errorf("unable to read root CAs: %v", err.Error())
		}
	}
	creds = credentials.NewTLS(config)
	return creds, nil
}

// applySecurityArgsToMeta validates plugin runtime arguments from OS, focusing on
// TLS functionality.
func applySecurityArgsToMeta(m *meta, args *Arg) error {
	if !args.TLSEnabled {
		if args.CertPath != "" || args.KeyPath != "" {
			return fmt.Errorf("excessive arguments given - CertPath and KeyPath are unused with TLS not enabled")
		}
		return nil
	}
	if args.CertPath == "" || args.KeyPath == "" {
		return fmt.Errorf("failed to enable TLS for plugin - need both CertPath and KeyPath")
	}
	m.CertPath = args.CertPath
	m.KeyPath = args.KeyPath
	m.TLSEnabled = true
	m.RootCertPaths = args.RootCertPaths
	return nil
}

// buildGRPCServer configures and builds GRPC server ready to server a plugin
// instance
func buildGRPCServer(typeOfPlugin pluginType, name string, version int, arg *Arg, opts ...MetaOpt) (server *grpc.Server, m *meta, err error) {
	m = newMeta(typeOfPlugin, name, version, opts...)

	if err := applySecurityArgsToMeta(m, arg); err != nil {
		return nil, nil, err
	}
	creds, err := makeGRPCCredentials(m)
	if err != nil {
		return nil, nil, err
	}
	if m.TLSEnabled {
		server = grpc.NewServer(tlsSetup.updateServerOptions(grpc.Creds(creds))...)
	} else {
		server = grpc.NewServer(tlsSetup.updateServerOptions()...)
	}
	return server, m, nil
}

// StartCollector is given a Collector implementation and its metadata,
// generates a response for the initial stdin / stdout handshake, and starts
// the plugin's gRPC server.
func StartCollector(plugin Collector, name string, version int, opts ...MetaOpt) int {
	app = cli.NewApp()
	app.Flags = Flags
	app.Action = startPlugin

	appArgs.plugin = plugin
	appArgs.name = name
	appArgs.version = version
	appArgs.opts = opts
	app.Version = strconv.Itoa(version)
	app.Usage = "a Snap collector"
	err := app.Run(getOSArgs())
	if err != nil {
		log.WithFields(log.Fields{
			"_block": "StartCollector",
		}).Error(err)
		return 1
	}
	return 0
}

// StartProcessor is given a Processor implementation and its metadata,
// generates a response for the initial stdin / stdout handshake, and starts
// the plugin's gRPC server.
func StartProcessor(plugin Processor, name string, version int, opts ...MetaOpt) int {
	app = cli.NewApp()
	app.Flags = Flags
	app.Action = startPlugin

	appArgs.plugin = plugin
	appArgs.name = name
	appArgs.version = version
	appArgs.opts = opts
	app.Version = strconv.Itoa(version)
	app.Usage = "a Snap processor"
	err := app.Run(getOSArgs())
	if err != nil {
		log.WithFields(log.Fields{
			"_block": "StartProcessor",
		}).Error(err)
		return 1
	}
	return 0
}

// StartPublisher is given a Publisher implementation and its metadata,
// generates a response for the initial stdin / stdout handshake, and starts
// the plugin's gRPC server.
func StartPublisher(plugin Publisher, name string, version int, opts ...MetaOpt) int {
	app = cli.NewApp()
	app.Flags = Flags
	app.Action = startPlugin

	appArgs.plugin = plugin
	appArgs.name = name
	appArgs.version = version
	appArgs.opts = opts
	app.Version = strconv.Itoa(version)
	app.Usage = "a Snap publisher"
	err := app.Run(getOSArgs())
	if err != nil {
		log.WithFields(log.Fields{
			"_block": "StartPublisher",
		}).Error(err)
		return 1
	}
	return 0
}

// StartStreamCollector is given a StreamCollector implementation and its metadata,
// generates a response for the initial stdin / stdout handshake, and starts
// the plugin's gRPC server.
func StartStreamCollector(plugin StreamCollector, name string, version int, opts ...MetaOpt) int {
	app = cli.NewApp()
	app.Flags = Flags
	app.Action = startPlugin

	appArgs.plugin = plugin
	appArgs.name = name
	appArgs.version = version
	//set gRPCStream as RPC type
	opts = append(opts, rpcType(gRPCStream))
	appArgs.opts = opts
	app.Version = strconv.Itoa(version)
	app.Usage = "a Snap collector"
	err := app.Run(getOSArgs())
	if err != nil {
		log.WithFields(log.Fields{
			"_block": "StartStreamCollector",
		}).Error(err)
		return 1
	}
	return 0
}

type server interface {
	Serve(net.Listener) error
}

type preamble struct {
	Meta          meta
	ListenAddress string
	PprofAddress  string
	Type          pluginType
	State         int
	ErrorMessage  string
}

func startPlugin(c *cli.Context) error {
	var (
		server           *grpc.Server
		meta             *meta
		pluginProxy      *pluginProxy
		MaxMetricsBuffer int64
	)
	libInputOutput.setContext(c)
	arg, err := processInput(c)
	if err != nil {
		return err
	}
	if lvl := c.Int("log-level"); lvl > arg.LogLevel {
		log.SetLevel(log.Level(lvl))
	} else {
		log.SetLevel(log.Level(arg.LogLevel))
	}
	logger := log.WithFields(
		log.Fields{
			"_block": "startPlugin",
		})
	switch plugin := appArgs.plugin.(type) {
	case Collector:
		proxy := &collectorProxy{
			plugin:      plugin,
			pluginProxy: *newPluginProxy(plugin),
		}
		pluginProxy = &proxy.pluginProxy
		server, meta, err = buildGRPCServer(collectorType, appArgs.name, appArgs.version, arg, appArgs.opts...)
		if err != nil {
			return cli.NewExitError(err, 2)
		}
		rpc.RegisterCollectorServer(server, proxy)
	case Processor:
		proxy := &processorProxy{
			plugin:      plugin,
			pluginProxy: *newPluginProxy(plugin),
		}
		pluginProxy = &proxy.pluginProxy
		server, meta, err = buildGRPCServer(processorType, appArgs.name, appArgs.version, arg, appArgs.opts...)
		if err != nil {
			return cli.NewExitError(err, 2)
		}
		rpc.RegisterProcessorServer(server, proxy)
	case Publisher:
		proxy := &publisherProxy{
			plugin:      plugin,
			pluginProxy: *newPluginProxy(plugin),
		}
		pluginProxy = &proxy.pluginProxy
		server, meta, err = buildGRPCServer(publisherType, appArgs.name, appArgs.version, arg, appArgs.opts...)
		if err != nil {
			return cli.NewExitError(err, 2)
		}
		rpc.RegisterPublisherServer(server, proxy)
	case StreamCollector:
		if c.IsSet("max-metrics-buffer") {
			MaxMetricsBuffer = c.Int64("max-metrics-buffer")
		} else {
			MaxMetricsBuffer = defaultMaxMetricsBuffer
		}

		logger.WithFields(log.Fields{
			"option": "max-metrics-buffer",
			"value":  MaxMetricsBuffer,
		}).Debug("setting max metrics buffer")

		maxCollectDuration, err := time.ParseDuration(collectDurationStr)
		if err != nil {
			return err
		}

		logger.WithFields(log.Fields{
			"option": "max-collect-duration",
			"value":  maxCollectDuration,
		}).Debug("setting max collect duration")

		proxy := &StreamProxy{
			plugin:             plugin,
			ctx:                context.Background(),
			pluginProxy:        *newPluginProxy(plugin),
			maxCollectDuration: maxCollectDuration,
			maxMetricsBuffer:   MaxMetricsBuffer,
		}

		pluginProxy = &proxy.pluginProxy
		server, meta, err = buildGRPCServer(streamCollectorType, appArgs.name, appArgs.version, arg, appArgs.opts...)
		if err != nil {
			return cli.NewExitError(err, 2)
		}
		rpc.RegisterStreamCollectorServer(server, proxy)
	default:
		logger.WithField("type", fmt.Sprintf("%T", plugin)).Fatal("Unknown plugin type")
	}

	if c.Bool("stand-alone") {
		httpPort := c.Int("stand-alone-port")
		preamble, err := printPreambleAndServe(server, meta, pluginProxy, arg.ListenPort, arg.Pprof)
		if err != nil {
			return err
		}

		go func() {
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, preamble)
			})

			listener, err := net.Listen("tcp", fmt.Sprintf(":%d", httpPort))
			if err != nil {
				log.WithFields(
					log.Fields{
						"port": httpPort,
					},
				).Fatal("Unable to get open port")
			}
			defer listener.Close()
			fmt.Printf("Preamble URL: %v\n", listener.Addr().String())
			err = http.Serve(listener, nil)
			if err != nil {
				log.Fatal(err)
			}
		}()
		<-pluginProxy.halt

	} else if libInputOutput.args() > 0 {
		// snapteld is starting the plugin
		// presumably with a single arg (valid json)
		preamble, err := printPreambleAndServe(server, meta, pluginProxy, arg.ListenPort, arg.Pprof)
		if err != nil {
			log.Fatal(err)
		}
		libInputOutput.printOut(preamble)
		go pluginProxy.HeartbeatWatch()
		<-pluginProxy.halt

	} else {
		// no arguments provided - run and display diagnostics to the user
		config := NewConfig()
		if c.IsSet("config") {
			err := json.Unmarshal([]byte(c.String("config")), &config)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("unable to parse config")
				return err
			}
		}
		// Get plugin config policy
		cPolicy, err := pluginProxy.plugin.GetConfigPolicy()
		if err != nil {
			logger.WithFields(log.Fields{
				"err": err,
			}).Errorf("cannot get config policy")
			return err
		}
		// Update config with defaults from config policy
		config.applyDefaults(cPolicy)

		switch pluginProxy.plugin.(type) {
		case Collector:
			return showDiagnostics(*meta, pluginProxy, config)
		case StreamCollector:
			fmt.Println("Diagnostics not currently available for streaming collector plugins.")
		case Processor:
			fmt.Println("Diagnostics not currently available for processor plugins.")
		case Publisher:
			fmt.Println("Diagnostics not currently available for publisher plugins.")
		}
	}
	return nil
}

func printPreambleAndServe(srv server, m *meta, p *pluginProxy, port string, isPprof bool) (string, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%v", ListenAddr, port))
	if err != nil {
		return "", err
	}
	l.Close()

	addr := fmt.Sprintf("%s:%v", ListenAddr, l.Addr().(*net.TCPAddr).Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return "", err
	}
	go func() {
		err := srv.Serve(lis)
		if err != nil {
			log.Fatal(err)
		}
	}()
	pprofAddr := "0"
	if isPprof {
		pprofAddr, err = startPprof()
		if err != nil {
			return "", err
		}
	}
	advertisedAddr, err := getAddr(ListenAddr)
	if err != nil {
		return "", err
	}
	resp := preamble{
		Meta:          *m,
		ListenAddress: fmt.Sprintf("%v:%v", advertisedAddr, l.Addr().(*net.TCPAddr).Port),
		Type:          m.Type,
		PprofAddress:  pprofAddr,
		State:         0, // Hardcode success since panics on err
	}
	preambleJSON, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}

	return string(preambleJSON), nil
}

// getAddr if we were provided the addr 0.0.0.0 we need to determine the
// address we will advertise to the framework in the preamble.
func getAddr(addr string) (string, error) {
	if strings.Compare(addr, "0.0.0.0") == 0 {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return "", err
		}
		for _, address := range addrs {
			// check the address type and if it is not a loopback the display it
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String(), nil
				}
			}
		}
	}
	return addr, nil
}

func showDiagnostics(m meta, p *pluginProxy, c Config) error {
	defer timeTrack(time.Now(), "showDiagnostics")
	printRuntimeDetails(m)
	err := printConfigPolicy(p, c)
	if err != nil {
		return err
	}

	met, err := printMetricTypes(p, c)
	if err != nil {
		return err
	}
	err = printCollectMetrics(p, met)
	if err != nil {
		return err
	}
	printContactUs()
	return nil

}

func printMetricTypes(p *pluginProxy, conf Config) ([]Metric, error) {
	defer timeTrack(time.Now(), "printMetricTypes")
	met, err := p.plugin.(Collector).GetMetricTypes(conf)
	if err != nil {
		return nil, fmt.Errorf("! Error in the call to GetMetricTypes: \n%v", err)
	}
	//apply any config passed in to met so that
	//CollectMetrics can see the config for each metric
	for i := range met {
		met[i].Config = conf
	}

	fmt.Println("Metric catalog will be updated to include: ")
	for _, j := range met {
		fmt.Printf("    Namespace: %v \n", j.Namespace.String())
	}
	return met, nil
}

func printConfigPolicy(p *pluginProxy, conf Config) error {
	defer timeTrack(time.Now(), "printConfigPolicy")
	requiredConfigs := ""
	cPolicy, err := p.plugin.(Collector).GetConfigPolicy()
	if err != nil {
		return err
	}

	fmt.Println("Config Policy:")
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	printFields(w, false, 0, "NAMESPACE", "KEY", "TYPE", "REQUIRED", "DEFAULT", "MINIMUM", "MAXIMUM")

	requiredConfigs += printConfigPolicyStringRules(cPolicy, conf, w)
	requiredConfigs += printConfigPolicyIntegerRules(cPolicy, conf, w)
	requiredConfigs += printConfigPolicyFloatRules(cPolicy, conf, w)
	requiredConfigs += printConfigPolicyBoolRules(cPolicy, conf, w)

	w.Flush()
	if requiredConfigs != "" {
		requiredConfigs += "! Please provide config in form of: -config '{\"key\":\"kelly\", \"spirit-animal\":\"coatimundi\"}'\n"
		err := fmt.Errorf(requiredConfigs)
		return err
	}

	return nil
}

func printFields(tw *tabwriter.Writer, indent bool, width int, fields ...interface{}) {
	var argArray []interface{}
	if indent {
		argArray = append(argArray, strings.Repeat(" ", width))
	}
	for i, field := range fields {
		if field != nil {
			argArray = append(argArray, field)
		} else {
			argArray = append(argArray, "")
		}
		if i < (len(fields) - 1) {
			argArray = append(argArray, "\t")
		}
	}
	fmt.Fprintln(tw, argArray...)
}

func stringInSlice(a string, list []string) (int, bool) {
	for i, b := range list {
		if b == a {
			return i, true
		}
	}
	return -1, false
}

func parseString(vals []string, name string, conf Config) (required string, minimum string, maximum string) {
	//check if required
	if _, okReq := stringInSlice("required:true", vals); okReq {
		required = "true"
	} else {
		required = "false"
	}

	//check if has_min:true
	if _, okMin := stringInSlice("has_min:true", vals); okMin {
		//Check if minimum is specified, if not, default = 0
		idxInArray := -1
		valueAtIndex := ""
		for i, b := range vals {
			if strings.Contains(b, "minimum:") {
				idxInArray = i
				valueAtIndex = b
			}
		}
		if idxInArray != -1 {
			//parse val[idx] to get contents after :
			idxOfColon := strings.Index(valueAtIndex, ":")
			minimum = valueAtIndex[idxOfColon+1:]
		} else {
			minimum = "0"
		}
	}

	//check if has_max:true
	if _, okMax := stringInSlice("has_max:true", vals); okMax {
		//Check if minimum is specified
		idxInArray := -1
		valueAtIndex := ""
		for i, b := range vals {
			if strings.Contains(b, "maximum:") {
				idxInArray = i
				valueAtIndex = b
			}
		}
		if idxInArray != -1 {
			//parse val[idx] to get contents after :
			idxOfColon := strings.Index(valueAtIndex, ":")
			maximum = valueAtIndex[idxOfColon+1:]
		} else {
			//No default max value
		}
	}

	return required, minimum, maximum
}

func checkForMissingRequirements(vals []string, name string, conf Config) (requiredConfigs string) {
	if _, okReq := stringInSlice("required:true", vals); okReq {
		if _, ok := conf[name]; !ok {
			requiredConfigs += "! Warning: \"" + name + "\" required by plugin and not provided in config \n"
		}
	}
	return requiredConfigs
}

func printConfigPolicyStringRules(cPolicy ConfigPolicy, conf Config, w *tabwriter.Writer) (requiredConfigs string) {
	for ns, v := range cPolicy.stringRules {
		for key, val := range v.Rules {
			defaultValue := ""
			if val.HasDefault {
				defaultValue = val.Default
			}
			if val.String() != "" {
				vals := strings.Fields(val.String())
				req, min, max := parseString(vals, ns, conf)
				printFields(w, false, 0, ns, key, "string", req, defaultValue, min, max)

				requiredConfigs += checkForMissingRequirements(vals, key, conf)
			} else {
				printFields(w, false, 0, ns, key, "string", "false", defaultValue, "", "")
			}
		}
	}
	return requiredConfigs
}

func printConfigPolicyIntegerRules(cPolicy ConfigPolicy, conf Config, w *tabwriter.Writer) (requiredConfigs string) {
	for k, v := range cPolicy.integerRules {
		for key, val := range v.Rules {
			defaultValue := ""
			if val.HasDefault {
				defaultValue = strconv.FormatInt(val.Default, 10)
			}
			if val.String() != "" {
				//parse info:
				vals := strings.Fields(val.String())
				req, min, max := parseString(vals, k, conf)
				printFields(w, false, 0, k, key, "integer", req, defaultValue, min, max)

				requiredConfigs += checkForMissingRequirements(vals, key, conf)
			} else {
				printFields(w, false, 0, k, key, "integer", "false", defaultValue, "", "")
			}
		}
	}
	return requiredConfigs
}

func printConfigPolicyFloatRules(cPolicy ConfigPolicy, conf Config, w *tabwriter.Writer) (requiredConfigs string) {
	for k, v := range cPolicy.floatRules {
		for key, val := range v.Rules {
			defaultValue := ""
			if val.HasDefault {
				defaultValue = strconv.FormatFloat(val.Default, 'f', -1, 64)
			}
			if val.String() != "" {
				//parse info:
				vals := strings.Fields(val.String())
				req, min, max := parseString(vals, k, conf)
				printFields(w, false, 0, k, key, "float", req, defaultValue, min, max)

				requiredConfigs += checkForMissingRequirements(vals, key, conf)
			} else {
				printFields(w, false, 0, k, key, "float", "false", defaultValue, "", "")
			}
		}
	}
	return requiredConfigs
}

func printConfigPolicyBoolRules(cPolicy ConfigPolicy, conf Config, w *tabwriter.Writer) (requiredConfigs string) {
	for k, v := range cPolicy.boolRules {
		for key, val := range v.Rules {
			defaultValue := ""
			if val.HasDefault {
				defaultValue = strconv.FormatBool(val.Default)
			}
			if val.String() != "" {
				//parse info:
				vals := strings.Fields(val.String())
				req, min, max := parseString(vals, k, conf)
				printFields(w, false, 0, k, key, "bool", req, defaultValue, min, max)

				requiredConfigs += checkForMissingRequirements(vals, key, conf)
			} else {
				printFields(w, false, 0, k, key, "bool", "false", defaultValue, "", "")
			}
		}
	}
	return requiredConfigs
}

func printCollectMetrics(p *pluginProxy, m []Metric) error {
	defer timeTrack(time.Now(), "printCollectMetrics")
	cltd, err := p.plugin.(Collector).CollectMetrics(m)
	if err != nil {
		return fmt.Errorf("! Error in the call to CollectMetrics. Please ensure your config contains any required fields mentioned in the error below. \n %v", err)
	}
	fmt.Println("Metrics that can be collected right now are: ")
	for _, j := range cltd {
		fmt.Printf("    Namespace: %-30v  Type: %-10T  Value: %v \n", j.Namespace, j.Data, j.Data)
	}
	return nil
}

func printRuntimeDetails(m meta) {
	defer timeTrack(time.Now(), "printRuntimeDetails")
	fmt.Printf("Runtime Details:\n    PluginName: %v, Version: %v \n    RPC Type: %v, RPC Version: %v \n", m.Name, m.Version, m.RPCType.String(), m.RPCVersion)
	fmt.Printf("    Operating system: %v \n    Architecture: %v \n    Go version: %v \n", runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func printContactUs() {
	fmt.Print("Thank you for using this Snap plugin. If you have questions or are running \ninto errors, please contact us on Github (github.com/intelsdi-x/snap) or \nour Slack channel (intelsdi-x.herokuapp.com). \nThe repo for this plugin can be found: github.com/intelsdi-x/<plugin-name>. \nWhen submitting a new issue on Github, please include this diagnostic \nprint out so that we have a starting point for addressing your question. \nThank you. \n\n")
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s \n\n", name, elapsed)
}

func processInput(c *cli.Context) (*Arg, error) {
	arg := &Arg{}
	if c.IsSet("log-level") {
		arg.LogLevel = c.Int("log-level")
	}
	if c.IsSet("port") {
		arg.ListenPort = c.String("port")
	}
	if c.IsSet("pprof") {
		arg.Pprof = c.Bool("pprof")
	}
	if c.IsSet("cert-path") {
		arg.CertPath = c.String("cert-path")
	}
	if c.IsSet("key-path") {
		arg.KeyPath = c.String("key-path")
	}
	if c.IsSet("root-cert-paths") {
		arg.RootCertPaths = c.String("root-cert-paths")
	}
	if c.IsSet("tls") {
		arg.TLSEnabled = true
	}

	if c.IsSet("max-collect-duration") {
		arg.MaxCollectDuration = c.String("max-collect-duration")
	}

	if c.IsSet("max-metrics-buffer") {
		arg.MaxMetricsBuffer = c.Int64("max-metrics-buffer")
	}

	return processArg(arg)
}
