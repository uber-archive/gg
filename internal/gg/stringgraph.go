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

// StringGraph represents a directed graph that addresses vertexes with
// strings.
type StringGraph map[string]StringSet

// NewStringGraph returns an empty string-addressed directed graph.
func NewStringGraph() StringGraph {
	return make(StringGraph)
}

// Clone duplicates a StringGraph.
func (g StringGraph) Clone() StringGraph {
	q := make(StringGraph, len(g))
	for key, value := range g {
		q[key] = value.Clone()
	}
	return q
}

// Targets returns the set of nodes that are reachable by a directed edge from a
// given node.
func (g StringGraph) Targets(src string) StringSet {
	ss, exists := g[src]
	if !exists {
		ss = make(StringSet)
		g[src] = ss
	}
	return ss
}

// Add adds an edge from one node address to another.
// Has no effect if the edge already exists.
func (g StringGraph) Add(src, tgt string) {
	g.Targets(src).Add(tgt)
}

// Has returns whether there is an edge from one node to another.
func (g StringGraph) Has(src, tgt string) bool {
	tgts, exists := g[src]
	if exists {
		_, exists = tgts[tgt]
	}
	return exists
}

// HasSource returns whether there are any edges departing from the addressed
// node.
func (g StringGraph) HasSource(src string) bool {
	_, exists := g[src]
	return exists
}

// Include adds all of the edges of another graph into this graph, if not
// already present.
func (g StringGraph) Include(q StringGraph) {
	for src, tgts := range q {
		g.Targets(src).Include(tgts)
	}
}

// Exclude removes all edges from another graph from this graph, if present.
func (g StringGraph) Exclude(q StringGraph) {
	for src, tgts := range q {
		g.Targets(src).Exclude(tgts)
	}
}

// StringSet returns a string set of all the source node addresses in this
// graph.
func (g StringGraph) StringSet() StringSet {
	ss := make(StringSet, len(g))
	for key := range g {
		ss.Add(key)
	}
	return ss
}

// Sources returns a sorted slice of node addresses for source nodes in this
// graph.
func (g StringGraph) Sources() []string {
	sources := make([]string, 0, len(g))
	for src := range g {
		sources = append(sources, src)
	}
	sort.Strings(sources)
	return sources
}

// Transitive captures the set of transitively reachable destinations from the
// given source, including itself.
func (g StringGraph) Transitive(frontier StringSet) StringSet {
	frontier = frontier.Clone()
	collection := frontier.Clone()
	for len(frontier) > 0 {
		for name := range frontier {
			delete(frontier, name)
			for imp := range g[name] {
				if collection.Has(imp) || frontier.Has(imp) {
					continue
				}
				collection.Add(imp)
				frontier.Add(imp)
			}
		}
	}
	return collection
}

// SourcesIntoStringSet accumulates all of the source nodes in this graph into
// a string set.
func (g StringGraph) SourcesIntoStringSet(q StringSet) {
	for key := range g {
		q.Add(key)
	}
}

// Intersects returns whether this string set contains any of the
// source nodes in a string graph.
func (g StringGraph) Intersects(s StringSet) bool {
	for key := range s {
		if _, exists := g[key]; exists {
			return true
		}
	}
	return false
}
