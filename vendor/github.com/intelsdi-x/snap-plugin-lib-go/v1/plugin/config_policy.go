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
	"strings"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
)

type ConfigPolicy struct {
	boolRules    map[string]*rpc.BoolPolicy
	floatRules   map[string]*rpc.FloatPolicy
	integerRules map[string]*rpc.IntegerPolicy
	stringRules  map[string]*rpc.StringPolicy
}

func NewConfigPolicy() *ConfigPolicy {
	return &ConfigPolicy{
		boolRules:    map[string]*rpc.BoolPolicy{},
		floatRules:   map[string]*rpc.FloatPolicy{},
		integerRules: map[string]*rpc.IntegerPolicy{},
		stringRules:  map[string]*rpc.StringPolicy{},
	}
}

// AddNewIntegerRule adds a new integerRule with the specified args to the integerRules map.
// The required arguments are ns([]string), key(string), req(bool)
// and optionally:
//		plugin.SetDefaultInt(int64),
//		plugin.SetMinInt(int64),
//		plugin.SetMaxInt(int64),
func (c *ConfigPolicy) AddNewIntRule(ns []string, key string, req bool, opts ...integerRuleOpt) error {
	if key == "" {
		return ErrEmptyKey
	}
	rule := rpc.IntegerRule{
		Required: req,
	}

	for _, opt := range opts {
		opt(&rule)
	}
	k := strings.Join(ns, ".")
	if c.integerRules[k] == nil {
		c.integerRules[k] = &rpc.IntegerPolicy{
			Rules: map[string]*rpc.IntegerRule{},
			Key:   ns,
		}
	}
	c.integerRules[k].Rules[key] = &rule
	return nil
}

// AddNewBoolRule adds a new boolRule with the specified args to the boolRules map.
// The required arguments are ns([]string), key(string), req(bool)
// and optionally:
//		plugin.SetDefaultBool(bool)
func (c *ConfigPolicy) AddNewBoolRule(ns []string, key string, req bool, opts ...boolRuleOpt) error {
	if key == "" {
		return ErrEmptyKey
	}
	rule := &rpc.BoolRule{
		Required: req,
	}

	for _, opt := range opts {
		opt(rule)
	}
	k := strings.Join(ns, ".") // Method used in daemon/ctree
	if c.boolRules[k] == nil {
		c.boolRules[k] = &rpc.BoolPolicy{
			Rules: map[string]*rpc.BoolRule{},
			Key:   ns,
		}
	}
	c.boolRules[k].Rules[key] = rule
	return nil
}

// AddNewFloatRule adds a new floatRule with the specified args to the floatRules map.
// The required arguments are ns([]string), key(string), req(bool)
// and optionally:
//		plugin.SetDefaultFloat(float64),
//		plugin.SetMinFloat(float64),
//		plugin.SetMaxFloat(float64),
func (c *ConfigPolicy) AddNewFloatRule(ns []string, key string, req bool, opts ...floatRuleOpt) error {
	if key == "" {
		return ErrEmptyKey
	}
	rule := &rpc.FloatRule{
		Required: req,
	}

	for _, opt := range opts {
		opt(rule)
	}
	k := strings.Join(ns, ".")
	if c.floatRules[k] == nil {
		c.floatRules[k] = &rpc.FloatPolicy{
			Rules: map[string]*rpc.FloatRule{},
			Key:   ns,
		}
	}
	c.floatRules[k].Rules[key] = rule
	return nil
}

// AddNewStringRule adds a new stringRule with the specified args to the stringRules map.
// The required arguments are ns([]string), key(string), req(bool)
// and optionally:
//		plugin.SetDefaultString(string)
func (c *ConfigPolicy) AddNewStringRule(ns []string, key string, req bool, opts ...stringRuleOpt) error {
	if key == "" {
		return ErrEmptyKey
	}
	rule := &rpc.StringRule{
		Required: req,
	}

	for _, opt := range opts {
		opt(rule)
	}
	k := strings.Join(ns, ".") // Method used in daemon/ctree
	if c.stringRules[k] == nil {
		c.stringRules[k] = &rpc.StringPolicy{
			Rules: map[string]*rpc.StringRule{},
			Key:   ns,
		}
	}
	c.stringRules[k].Rules[key] = rule
	return nil
}

func newGetConfigPolicyReply(cfg ConfigPolicy) *rpc.GetConfigPolicyReply {
	return &rpc.GetConfigPolicyReply{
		BoolPolicy:    cfg.boolRules,
		FloatPolicy:   cfg.floatRules,
		IntegerPolicy: cfg.integerRules,
		StringPolicy:  cfg.stringRules,
	}
}
