package plugin

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"net/http/pprof"

	"github.com/julienschmidt/httprouter"
)

var (
	listenPort = "0"
	LogLevel   = uint8(2)
	pprofPort  = "0"
)

// Arguments passed to startup of Plugin
type Arg struct {
	// Plugin log level, see logrus.Loglevel
	LogLevel uint8
	// Ping timeout duration
	PingTimeoutDuration time.Duration

	// The listen port
	ListenPort string

	// enable pprof
	Pprof bool
}

// getArgs returns plugin args or default ones
func getArgs() error {
	pluginArg := &Arg{}
	if os.Args[1] == "" {
		return nil
	}
	err := json.Unmarshal([]byte(os.Args[1]), pluginArg)
	if err != nil {
		return err
	}

	// If no port was provided we let the OS select a port for us.
	// This is safe as address is returned in the Response and keep
	// alive prevents unattended plugins.
	if pluginArg.ListenPort != "" {
		listenPort = pluginArg.ListenPort
	}

	// If PingTimeoutDuration was provided we set it
	if pluginArg.PingTimeoutDuration != 0 {
		PingTimeoutDurationDefault = pluginArg.PingTimeoutDuration
	}
	if pluginArg.Pprof {
		return getPort()
	}

	return nil
}

func getPort() error {
	router := httprouter.New()
	router.GET("/debug/pprof/", index)
	router.GET("/debug/pprof/block", index)
	router.GET("/debug/pprof/goroutine", index)
	router.GET("/debug/pprof/heap", index)
	router.GET("/debug/pprof/threadcreate", index)
	router.GET("/debug/pprof/cmdline", cmdline)
	router.GET("/debug/pprof/profile", profile)
	router.GET("/debug/pprof/symbol", symbol)
	router.GET("/debug/pprof/trace", trace)
	addr, err := net.ResolveTCPAddr("tcp", ":0")
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	pprofPort = fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port)

	go func() {
		log.Fatal(http.Serve(l, router))
	}()

	return nil
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Index(w, r)
}

func cmdline(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Cmdline(w, r)
}

func profile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Profile(w, r)
}

func symbol(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Symbol(w, r)
}

func trace(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Trace(w, r)
}
