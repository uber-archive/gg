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

func newTestGraph() StringGraph {
	g := NewStringGraph()
	g.Add("dca", "irn")
	g.Add("dca", "sfo")
	g.Add("phx", "sfo")
	g.Add("sfo", "dca")
	g.Add("sfo", "phx")
	g.Add("sfo", "sea")
	return g
}

func TestStringGraphBasics(t *testing.T) {
	g := newTestGraph()
	assert.True(t, g.Has("sfo", "phx"))
	assert.False(t, g.Has("phx", "dca"))
	assert.False(t, g.Has("sea", "dca"))
	assert.True(t, g.HasSource("phx"))
	assert.False(t, g.HasSource("sea"))
	assert.Equal(t, []string{"dca", "phx", "sfo"}, g.Sources())
	assert.Equal(t, []string{"dca", "phx", "sfo"}, g.StringSet().Keys())
	assert.Equal(t, []string{"dca", "phx", "sea"}, g.Targets("sfo").Keys())
}

func TestStringGraphSourcesIntoStringSet(t *testing.T) {
	g := newTestGraph()
	ss := make(StringSet)
	g.SourcesIntoStringSet(ss)
	assert.Equal(t, []string{"dca", "phx", "sfo"}, ss.Keys())
}

func TestStringGraphInclude(t *testing.T) {
	h := NewStringGraph()
	h.Add("sea", "sfo")
	g := newTestGraph()
	g.Include(h)
	assert.True(t, g.Has("sea", "sfo"))
	assert.True(t, g.Has("sfo", "sea"))
}

func TestStringGraphExclude(t *testing.T) {
	h := NewStringGraph()
	h.Add("sea", "sfo")
	h.Add("sfo", "sea")
	g := newTestGraph()
	g.Exclude(h)
	assert.False(t, g.Has("sea", "sfo"))
	assert.False(t, g.Has("sfo", "sea"))
	assert.True(t, g.Has("phx", "sfo"))
}

func TestStringGraphTransitive(t *testing.T) {
	g := newTestGraph()
	assert.Equal(t, []string{
		"dca",
		"irn",
		"phx",
		"sea",
		"sfo",
	}, g.Transitive(StringSet{"sfo": {}}).Keys())
}

func TestStringGraphIntersects(t *testing.T) {
	g := newTestGraph()
	has := NewStringSet([]string{"sfo"})
	hasnt := NewStringSet([]string{"sea"})
	assert.True(t, g.Intersects(has))
	assert.False(t, g.Intersects(hasnt))
}
