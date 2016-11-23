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

// Config is a type alias for map[string]interface{} to allow the
// helper functions Get{String,Bool,Float,Int} to be defined.
type Config map[string]interface{}

// GetString takes a given key and checks the config for both
// that the key exists, and that it is of type string.
// Returns an error if either of these is false.
func (c Config) GetString(key string) (string, error) {
	var (
		strout string
		val    interface{}
		ok     bool
	)

	if val, ok = c[key]; !ok {
		return strout, ErrConfigNotFound
	}
	if strout, ok = val.(string); !ok {
		return strout, ErrNotAString
	}
	return strout, nil
}

// GetBool takes a given key and checks the config for both
// that the key exists, and that it is of type bool.
// Returns an error if either of these is false.
func (c Config) GetBool(key string) (bool, error) {
	var (
		bout bool
		val  interface{}
		ok   bool
	)

	if val, ok = c[key]; !ok {
		return bout, ErrConfigNotFound
	}

	if bout, ok = val.(bool); !ok {
		return bout, ErrNotABool
	}

	return bout, nil
}

// GetFloat takes a given key and checks the config for both
// that the key exists, and that it is of type float64.
// Returns an error if either of these is false.
func (c Config) GetFloat(key string) (float64, error) {
	var (
		fout float64
		val  interface{}
		ok   bool
	)

	if val, ok = c[key]; !ok {
		return fout, ErrConfigNotFound
	}

	if fout, ok = val.(float64); !ok {
		return fout, ErrNotAFloat
	}

	return fout, nil
}

// GetInt takes a given key and checks the config for both
// that the key exists, and that it is of type int64.
// Returns an error if either of these is false.
func (c Config) GetInt(key string) (int64, error) {
	var (
		iout int64
		val  interface{}
		ok   bool
	)

	if val, ok = c[key]; !ok {
		return iout, ErrConfigNotFound
	}

	if iout, ok = val.(int64); !ok {
		return iout, ErrNotAnInt
	}

	return iout, nil
}
