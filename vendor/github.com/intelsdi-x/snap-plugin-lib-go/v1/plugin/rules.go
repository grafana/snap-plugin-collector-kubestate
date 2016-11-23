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

import "github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"

type boolRuleOpt func(*rpc.BoolRule)

// SetDefaultBool Allows easy setting of the Default value for an rpc.BoolRule.
// Usage:
//		AddNewBoolRule(ns, key, req, config.SetDefaultBool(default))
func SetDefaultBool(in bool) boolRuleOpt {
	return func(i *rpc.BoolRule) {
		i.Default = in
		i.HasDefault = true
	}
}

type floatRuleOpt func(*rpc.FloatRule)

// SetDefaultFloat Allows easy setting of the Default value for an rpc.FloatRule.
// Usage:
//		AddNewFloatRule(ns, key, req, config.SetDefaultFloat(default))
func SetDefaultFloat(in float64) floatRuleOpt {
	return func(i *rpc.FloatRule) {
		i.Default = in
		i.HasDefault = true
	}
}

// SetMaxFloat Allows easy setting of the Max value for an rpc.FloatRule.
// Usage:
//		AddNewFloatRule(ns, key, req, config.SetMaxFloat(max))
func SetMaxFloat(max float64) floatRuleOpt {
	return func(i *rpc.FloatRule) {
		i.Maximum = max
		i.HasMax = true
	}
}

// SetMinFloat Allows easy setting of the Min value for an rpc.FloatRule.
// Usage:
//		AddNewFloatRule(ns, key, req, config.SetMinFloat(min))
func SetMinFloat(min float64) floatRuleOpt {
	return func(i *rpc.FloatRule) {
		i.Minimum = min
		i.HasMin = true
	}
}

type integerRuleOpt func(*rpc.IntegerRule)

// SetDefaultInt Allows easy setting of the Default value for an rpc.IntegerRule.
// Usage:
//		AddNewIntegerRule(ns, key, req, config.SetDefaultInt(default))
func SetDefaultInt(in int64) integerRuleOpt {
	return func(i *rpc.IntegerRule) {
		i.Default = in
		i.HasDefault = true
	}
}

// SetMaxInt Allows easy setting of the Max value for an rpc.IntegerRule.
// Usage:
//		AddNewIntegerRule(ns, key, req, config.SetMaxInt(max))
func SetMaxInt(max int64) integerRuleOpt {
	return func(i *rpc.IntegerRule) {
		i.Maximum = max
		i.HasMax = true
	}
}

// SetMinInt Allows easy setting of the Min value for an rpc.IntegerRule.
// Usage:
//		AddNewIntegerRule(ns, key, req, config.SetMinInt(min))
func SetMinInt(min int64) integerRuleOpt {
	return func(i *rpc.IntegerRule) {
		i.Minimum = min
		i.HasMin = true
	}
}

type stringRuleOpt func(*rpc.StringRule)

// SetDefaultString Allows easy setting of the Default value for an rpc.StringRule.
// Usage:
//		AddNewStringRule(ns, key, req, config.SetDefaultString(default))
func SetDefaultString(in string) stringRuleOpt {
	return func(i *rpc.StringRule) {
		i.Default = in
		i.HasDefault = true
	}
}
