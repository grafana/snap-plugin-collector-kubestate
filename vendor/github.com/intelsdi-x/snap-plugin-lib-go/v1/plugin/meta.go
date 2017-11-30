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

import "time"

type router int

const (
	LRURouter router = iota
	StickyRouter
	ConfigBasedRouter
)

const defaultConcurrencyCount = 5

// MetaOpt is used to apply optional metadata on a plugin
type MetaOpt func(m *meta)

// ConcurrencyCount is the max number of concurrent calls the plugin
// should take.  For example:
// If there are 5 tasks using the plugin and its concurrency count is 2,
// snapteld will keep 3 plugin instances running.
// ConcurrencyCount overwrites the default (5) for a Meta's ConcurrencyCount.
func ConcurrencyCount(cc int) MetaOpt {
	return func(m *meta) {
		m.ConcurrencyCount = cc
	}
}

// Exclusive == true results in a single instance of the plugin running
// regardless of the number of tasks using the plugin.
// Exclusive overwrites the default (false) for a Meta's Exclusive key.
func Exclusive(e bool) MetaOpt {
	return func(m *meta) {
		m.Exclusive = e
	}
}

// RoutingStrategy will override the routing strategy this plugin requires.
// The default routing strategy is Least Recently Used.
// RoutingStrategy overwrites the default (LRU) for a Meta's RoutingStrategy.
func RoutingStrategy(r router) MetaOpt {
	return func(m *meta) {
		m.RoutingStrategy = r
	}
}

// CacheTTL will override the default cache TTL for the this plugin. snapteld
// caches metrics on the daemon side for a default of 500ms.
// CacheTTL overwrites the default (500ms) for a Meta's CacheTTL.
func CacheTTL(t time.Duration) MetaOpt {
	return func(m *meta) {
		m.CacheTTL = t
	}
}

// metaRPCType sets the metaRPCType for the meta object. Used only internally.
func rpcType(typ metaRPCType) MetaOpt {
	return func(m *meta) {
		m.RPCType = typ
	}
}

type pluginType int

const (
	collectorType pluginType = iota
	processorType
	publisherType
	streamCollectorType
)

type metaRPCType int

const (
	gRPC       metaRPCType = 2
	gRPCStream             = 3
)

func (t metaRPCType) String() string {
	switch t {
	case gRPC:
		return "gRPC"
	case gRPCStream:
		return "streaming gRPC"
	default:
		return "unknown"
	}
}

// meta is the metadata for a plugin
type meta struct {
	// A plugin's unique identifier is type:name:version.
	Type       pluginType
	Name       string
	Version    int
	RPCType    metaRPCType
	RPCVersion int

	ConcurrencyCount int
	Exclusive        bool
	Unsecure         bool
	CacheTTL         time.Duration
	RoutingStrategy  router
	CertPath         string
	KeyPath          string
	TLSEnabled       bool
	RootCertPaths    string
}

// newMeta sets defaults, applies options, and then returns a meta struct
func newMeta(plType pluginType, name string, version int, opts ...MetaOpt) *meta {
	p := meta{
		Name:             name,
		Version:          version,
		Type:             plType,
		ConcurrencyCount: defaultConcurrencyCount,
		RoutingStrategy:  LRURouter,
		RPCType:          gRPC, // GRPC type
		RPCVersion:       1,    // This is v1 lib
		// Unsecure is a legacy value not used for grpc, but needed to avoid
		// calling SetKey needlessly.
		Unsecure: true,
	}

	for _, opt := range opts {
		opt(&p)
	}

	return &p
}
