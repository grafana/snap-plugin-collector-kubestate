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

package str

// StringSet is set implementation on top of map
type StringSet struct {
	set map[string]bool
}

// Add adds new item to set
func (set *StringSet) Add(element string) bool {
	_, found := set.set[element]
	set.set[element] = true
	return !found
}

// Delete removes element from set
func (set *StringSet) Delete(element string) bool {
	_, found := set.set[element]
	if found {
		delete(set.set, element)
	}
	return found
}

// Elements returns list of set elements
func (set *StringSet) Elements() []string {
	iter := []string{}
	for k := range set.set {
		iter = append(iter, k)
	}
	return iter
}

// Size returns number of elements in set
func (set *StringSet) Size() int {
	return len(set.set)
}

// InitSet initializes sets internal map
func InitSet() StringSet {
	set := StringSet{}
	set.set = map[string]bool{}
	return set
}
