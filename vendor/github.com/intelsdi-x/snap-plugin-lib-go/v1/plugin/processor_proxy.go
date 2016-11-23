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
	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
)

//TODO(danielscottt): plugin panics

type processorProxy struct {
	pluginProxy

	plugin Processor
}

func (p *processorProxy) Process(ctx context.Context, arg *rpc.PubProcArg) (*rpc.MetricsReply, error) {
	metrics := []Metric{}
	for _, mt := range arg.Metrics {
		metric := fromProtoMetric(mt)
		metrics = append(metrics, metric)
	}
	cfg := fromProtoConfig(arg.Config)
	r, err := p.plugin.Process(metrics, cfg)
	if err != nil {
		return nil, err
	}

	mts := []*rpc.Metric{}
	for _, mt := range r {
		metric, err := toProtoMetric(mt)
		if err != nil {
			return nil, err
		}
		mts = append(mts, metric)
	}
	reply := &rpc.MetricsReply{Metrics: mts}
	return reply, nil
}
