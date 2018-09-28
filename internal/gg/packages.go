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
	"strings"
)

// Packages tracks the import and export paths of all the Go packages in a
// module or in an entire solution.
type Packages struct {
	// All is the union of all imported and exported packages.
	All StringSet
	// Commands that exist. The Imports of commands retain packages for main
	// binaries.
	Commands StringSet
	// Exports are packages that can be imported.
	Exports StringSet
	// Imports is a graph of the imports of packages.
	Imports StringGraph
	// TestImports is a graph of imports from test files from packages.
	TestImports StringGraph
	// CoImports is a graph of what packages import a package.
	CoImports StringGraph
	// CoTestImports is a graph of what packages import a package for tests.
	CoTestImports StringGraph
}

// NewPackages returns Packages for tracking imports and exports.
func NewPackages() Packages {
	return Packages{
		All:           make(StringSet),
		Commands:      make(StringSet),
		Exports:       make(StringSet),
		Imports:       NewStringGraph(),
		TestImports:   NewStringGraph(),
		CoImports:     NewStringGraph(),
		CoTestImports: NewStringGraph(),
	}
}

// Defined returns whether the package solution has any packages.
// When we read a lockfile, the packages section might not be populated.
// This indicates that the module loader needs to read packages.
func (p Packages) Defined() bool {
	return len(p.Exports) > 0
}

// Clone creates a deep copy of Packages.
func (p Packages) Clone() Packages {
	return Packages{
		All:           p.All.Clone(),
		Commands:      p.Commands.Clone(),
		Exports:       p.Exports.Clone(),
		Imports:       p.Imports.Clone(),
		TestImports:   p.TestImports.Clone(),
		CoImports:     p.CoImports.Clone(),
		CoTestImports: p.CoTestImports.Clone(),
	}
}

// Command adds a command (a package with a "main") package to the graph.
func (p Packages) Command(exp string) {
	if isBuiltin(exp) {
		return
	}
	p.Commands.Add(exp)
	p.All.Add(exp)
}

// Export adds a non-main package to the graph.
// This indicates that the package is importable in the graph.
func (p Packages) Export(exp string) {
	if isBuiltin(exp) {
		return
	}
	p.Exports.Add(exp)
	p.All.Add(exp)
}

// Import adds a normal dependency to the package graph,
// indicating that a package imports another.
func (p Packages) Import(exp, imp string) {
	if isBuiltin(imp) {
		return
	}
	p.Imports.Add(exp, imp)
	p.CoImports.Add(imp, exp)
	p.All.Add(imp)
	p.All.Add(exp)
}

// TestImport adds a test dependency to the package graph,
// indicating that a package imports another in its tests.
func (p Packages) TestImport(exp, imp string) {
	if isBuiltin(imp) {
		return
	}
	p.TestImports.Add(exp, imp)
	p.CoTestImports.Add(imp, exp)
	p.All.Add(imp)
	p.All.Add(exp)
}

// Include subsumes the imports and exports of another packages collection.
func (p Packages) Include(q Packages) {
	p.All.Include(q.All)
	p.Commands.Include(q.Commands)
	p.Exports.Include(q.Exports)
	p.Imports.Include(q.Imports)
	p.TestImports.Include(q.TestImports)
	p.CoImports.Include(q.CoImports)
	p.CoTestImports.Include(q.CoTestImports)
}

// isBuiltin indicates whether a package is the purview of the language to
// provide, as opposed to packages that need to be vendored.
func isBuiltin(pkg string) bool {
	parts := strings.Split(pkg, "/")
	if len(parts) > 0 {
		if strings.Contains(parts[0], ".") {
			return false
		}
	}
	return true
}

// NecessaryPackages computes the sets of packages that are transitively
// imported by commands and tests respectively from the working copy.
// NecessaryPackages assumes that packages are a supergraph of ownPackages.
func NecessaryPackages(ownPackages, packages Packages) (StringSet, StringSet) {
	// Collect imports of commands
	imports := packages.Imports.Transitive(ownPackages.Commands)

	// Collect imports of own test imports
	testImporters := ownPackages.CoTestImports.StringSet()
	testImports := packages.Imports.Transitive(testImporters)

	return imports, testImports
}

// MissingPackages computes the sets of packages that are imported but absent
// from a set of packages, for commands and tests respectively, given the
// package graph of the working copy, and the aggregated package graph in a
// proposed vendor solution.
func MissingPackages(ownPackages, packages Packages) (StringSet, StringSet) {
	packages = packages.Clone()
	packages.Include(ownPackages)
	imports, testImports := NecessaryPackages(ownPackages, packages)
	imports.Exclude(packages.Exports)
	imports.Exclude(packages.Commands)
	testImports.Exclude(packages.Exports)
	testImports.Exclude(imports)
	return imports, testImports
}
