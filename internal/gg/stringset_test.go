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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	noNames  = []string{}
	allNames = []string{
		"alice",
		"blake",
		"carey",
		"drew",
		"evelyn",
		"frank",
		"gayle",
		"herb",
		"irene",
		"james",
		"kris",
		"leslie",
		"morgan",
	}
	unisexNames = []string{
		"blake",
		"carey",
		"drew",
		"kris",
		"leslie",
	}
	likelyLadyNames = []string{
		"alice",
		"evelyn",
		"gayle",
		"irene",
		"morgan",
	}
	likelyGentNames = []string{
		"frank",
		"herb",
		"james",
	}
	maybeLadyNames = []string{
		"alice",
		"blake",
		"carey",
		"drew",
		"evelyn",
		"gayle",
		"irene",
		"kris",
		"morgan",
	}
	maybeGentNames = []string{
		"blake",
		"carey",
		"drew",
		"frank",
		"herb",
		"james",
		"kris",
		"leslie",
	}
	namesTable = [][]string{
		noNames,
		allNames,
		unisexNames,
		likelyLadyNames,
		likelyGentNames,
		maybeLadyNames,
		maybeGentNames,
	}
)

func TestStringSet(t *testing.T) {
	all := NewStringSet(allNames)
	all.Exclude(all)
	assert.Len(t, all, 0)
}

func TestStringSetClone(t *testing.T) {
	original := NewStringSet(allNames)
	all := original.Clone()
	all.Exclude(all)
	assert.Len(t, all, 0)
	assert.Len(t, original, len(allNames))
}

func TestStringSetEqual(t *testing.T) {
	for i := 0; i < len(namesTable); i++ {
		for j := 0; j < len(namesTable); j++ {
			this := NewStringSet(namesTable[i])
			that := NewStringSet(namesTable[j])
			assert.Equal(t, i == j, this.Equal(that))
			assert.Equal(t, i == j, that.Equal(this))
		}
	}
}

func TestStringSetAdd(t *testing.T) {
	names := make(StringSet)
	for _, name := range maybeLadyNames {
		names.Add(name)
	}
	for _, name := range maybeGentNames {
		names.Add(name)
	}
	assert.True(t, NewStringSet(allNames).Equal(names))
}

func TestStringSetHas(t *testing.T) {
	names := NewStringSet(allNames)
	assert.True(t, names.Has("alice"))
	assert.False(t, names.Has("zed"))
}

func TestStringSetInclude(t *testing.T) {
	names := make(StringSet)
	names.Include(NewStringSet(maybeLadyNames))
	names.Include(NewStringSet(maybeGentNames))
	assert.True(t, NewStringSet(allNames).Equal(names))
}

func TestStringSetIncludeNil(t *testing.T) {
	names := make(StringSet)
	names.Include(NewStringSet(maybeLadyNames))
	names.Include(nil)
	assert.True(t, NewStringSet(maybeLadyNames).Equal(names))
}

func TestStringSetExclude(t *testing.T) {
	names := make(StringSet)
	names.Include(NewStringSet(maybeLadyNames))
	names.Exclude(NewStringSet(maybeGentNames))
	assert.True(t, NewStringSet(likelyLadyNames).Equal(names))
}

func TestStringSetExcludeNil(t *testing.T) {
	names := make(StringSet)
	names.Include(NewStringSet(maybeLadyNames))
	names.Exclude(nil)
	assert.True(t, NewStringSet(maybeLadyNames).Equal(names))
}

func TestStringSetUnion(t *testing.T) {
	ladies := NewStringSet(maybeLadyNames)
	gents := NewStringSet(maybeGentNames)
	all := NewStringSet(allNames)
	assert.True(t, all.Equal(ladies.Union(gents)))
}

func TestStringSetIntersects(t *testing.T) {
	tests := []struct {
		msg        string
		this, that []string
		want       bool
	}{
		{
			msg:  "degenerate case",
			this: nil,
			that: nil,
			want: false,
		},
		{
			msg:  "exact overlap",
			this: allNames,
			that: allNames,
			want: true,
		},
		{
			msg:  "partial overlap",
			this: maybeLadyNames,
			that: maybeGentNames,
			want: true,
		},
		{
			msg:  "no overlap",
			this: likelyLadyNames,
			that: likelyGentNames,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			this := NewStringSet(tt.this)
			that := NewStringSet(tt.that)
			assert.Equal(t, tt.want, this.Intersects(that))
		})
	}
}

func TestStringSetKeys(t *testing.T) {
	assert.Equal(t, allNames, NewStringSet(allNames).Keys())
}
