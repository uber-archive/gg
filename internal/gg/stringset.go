// Copyright (c) 2018 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// Copyright (c) 2018 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gg

import "sort"

// StringSet is a set of strings.
type StringSet map[string]struct{}

// NewStringSet constructs a string set from a slice of strings.
func NewStringSet(strings []string) StringSet {
	ss := make(StringSet, len(strings))
	for _, str := range strings {
		ss[str] = struct{}{}
	}
	return ss
}

// Clone duplicates a set of strings.
func (s StringSet) Clone() StringSet {
	t := make(StringSet, len(s))
	for str := range s {
		t[str] = struct{}{}
	}
	return t
}

// Equal returns whether two sets of strings are equivalent: that they contain
// the same number of strings and contain all of the same strings.
func (s StringSet) Equal(t StringSet) bool {
	if len(s) != len(t) {
		return false
	}
	for str := range s {
		if _, ok := t[str]; !ok {
			return false
		}
	}
	return true
}

// Has returns whether a set contains a string.
func (s StringSet) Has(str string) bool {
	_, ok := s[str]
	return ok
}

// Add adds a string to the set.
func (s StringSet) Add(str string) {
	s[str] = struct{}{}
}

// Include subsumes another set of strings into itself.
func (s StringSet) Include(t StringSet) {
	for key := range t {
		s[key] = struct{}{}
	}
}

// Exclude excludes another set of strings from itself.
func (s StringSet) Exclude(t StringSet) {
	for key := range t {
		delete(s, key)
	}
}

// Union produces a new set that contains all of the strings in this and
// another string set.
func (s StringSet) Union(t StringSet) StringSet {
	u := make(StringSet, len(s)+len(t))
	u.Include(s)
	u.Include(t)
	return u
}

// Intersects returns whether this and that string set have any common strings.
func (s StringSet) Intersects(t StringSet) bool {
	for key := range s {
		if _, ok := t[key]; ok {
			return true
		}
	}
	return false
}

// Keys returns a sorted list of all the strings in the set.
func (s StringSet) Keys() []string {
	keys := make([]string, 0, len(s))
	for key := range s {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
