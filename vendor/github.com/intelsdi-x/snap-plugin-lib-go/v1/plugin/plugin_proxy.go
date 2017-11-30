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
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
)

// TODO(danielscottt): logging
// TODO(danielscottt): plugin panics

var (
	// Timeout settings

	// PingTimeoutLimit is the number of successively missed ping health
	// checks which must occur before the plugin is stopped
	PingTimeoutLimit = 3
	// PingTimeoutDuration is the duration during which a ping healthcheck
	// should be received
	PingTimeoutDuration = 3 * time.Second
)

type pluginProxyConstructor func(Plugin) *pluginProxy

type pluginProxy struct {
	plugin              Plugin
	LastPing            time.Time
	PingTimeoutDuration time.Duration
	halt                chan struct{}
}

// pluginProxyCtor refers to function creating a new plugin proxy instance,
// should never be exposed or used outside of tests
var pluginProxyCtor pluginProxyConstructor = defaultPluginProxyCtor

// newPluginProxy delivers a new plugin proxy instance using a constructor
func newPluginProxy(plugin Plugin) *pluginProxy {
	return pluginProxyCtor(plugin)
}

// defaultPluginProxyCtor delivers new plugin instance using default setup
// (e.g.: default plugin timeout)
func defaultPluginProxyCtor(plugin Plugin) *pluginProxy {
	return &pluginProxy{
		plugin:              plugin,
		PingTimeoutDuration: PingTimeoutDuration,
		halt:                make(chan struct{}),
	}
}

func (p *pluginProxy) Ping(ctx context.Context, arg *rpc.Empty) (*rpc.ErrReply, error) {
	p.LastPing = time.Now()
	//Change to log
	fmt.Println("Heartbeat received at:", p.LastPing)
	return &rpc.ErrReply{}, nil
}

func (p *pluginProxy) Kill(ctx context.Context, arg *rpc.KillArg) (*rpc.ErrReply, error) {
	// TODO(CDR) log kill reason
	p.halt <- struct{}{}
	return &rpc.ErrReply{}, nil
}

func (p *pluginProxy) GetConfigPolicy(ctx context.Context, arg *rpc.Empty) (*rpc.GetConfigPolicyReply, error) {
	policy, err := p.plugin.GetConfigPolicy()
	if err != nil {
		return nil, err
	}
	return newGetConfigPolicyReply(policy), nil
}

func (p *pluginProxy) HeartbeatWatch() {
	p.LastPing = time.Now()
	fmt.Println("Heartbeat started")
	count := 0
	for {
		if time.Since(p.LastPing) >= p.PingTimeoutDuration {
			count++
			fmt.Printf("Heartbeat timeout %v of %v.  (Duration between checks %v)", count, PingTimeoutLimit, p.PingTimeoutDuration)
			if count >= PingTimeoutLimit {
				fmt.Println("Heartbeat timeout expired!")
				defer close(p.halt)
				return
			}
		} else {
			fmt.Println("Heartbeat timeout reset")
			// Reset count
			count = 0
		}
		time.Sleep(p.PingTimeoutDuration)
	}

}
