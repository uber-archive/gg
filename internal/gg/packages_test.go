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

func averyPackages() Packages {
	packages := NewPackages()

	packages.Export("example.com/avery")
	packages.Import("example.com/avery", "net/http")
	packages.Import("example.com/avery", "time")
	packages.Import("example.com/avery", "example.com/blake")
	packages.Import("example.com/avery", "example.com/carey")
	packages.TestImport("example.com/avery", "example.com/carey")
	packages.TestImport("example.com/avery", "example.com/drew/drewutil")

	packages.Command("example.com/avery/cmd/avery")
	packages.Import("example.com/avery/cmd/avery", "example.com/avery")

	return packages
}

func blakePackages() Packages {
	packages := NewPackages()
	packages.Export("example.com/blake")
	packages.Command("example.com/blake/cmd/blake")
	// Command in dependency, importantly absent from missing packages.
	return packages
}

func careyPackages() Packages {
	packages := NewPackages()
	packages.Export("example.com/carey")
	packages.Import("example.com/carey", "example.com/carey/internal/carey")
	// Oddly, internal carey not included in carey package.
	packages.Import("example.com/carey/internal/carey", "example.com/carey")
	// But an import cycle?  That's upossible!
	return packages
}

func drewPackages() Packages {
	packages := NewPackages()
	packages.Export("example.com/drew/drewutil")
	packages.Import("example.com/drew/drewutil", "example.com/carey")
	packages.Import("example.com/drew/drewutil", "example.com/bogus")
	return packages
}

func TestPackagesDefined(t *testing.T) {
	packages := NewPackages()
	assert.False(t, packages.Defined())

	packages.Export("example.com/example")
	assert.True(t, packages.Defined())
}

func TestPackages(t *testing.T) {
	a := averyPackages()
	b := blakePackages()
	c := careyPackages()

	alone := a.Imports.Transitive(StringSet{"example.com/avery/cmd/avery": {}})
	assert.Equal(t, []string{
		"example.com/avery",
		"example.com/avery/cmd/avery",
		"example.com/blake",
		"example.com/carey",
	}, alone.Keys())

	a.Include(b)
	a.Include(c)

	needed := a.Imports.Transitive(StringSet{"example.com/avery/cmd/avery": {}})
	assert.Equal(t, []string{
		"example.com/avery",
		"example.com/avery/cmd/avery",
		"example.com/blake",
		"example.com/carey",
		"example.com/carey/internal/carey",
	}, needed.Keys())

}

func TestAllPackages(t *testing.T) {
	a := averyPackages()
	b := blakePackages()
	c := careyPackages()
	a.Include(b)
	a.Include(c)
	assert.Equal(t, []string{
		"example.com/avery",
		"example.com/avery/cmd/avery",
		"example.com/blake",
		"example.com/blake/cmd/blake",
		"example.com/carey",
		"example.com/carey/internal/carey",
		"example.com/drew/drewutil",
	}, a.All.Keys())
}

func TestMissingPackages(t *testing.T) {
	var imports, testImports StringSet

	// Initially empty
	ownPackages := averyPackages()
	packages := NewPackages()

	imports, testImports = MissingPackages(ownPackages, packages)
	assert.Equal(t, []string{
		"example.com/blake",
		"example.com/carey",
	}, imports.Keys())
	assert.Equal(t, []string{
		"example.com/drew/drewutil",
	}, testImports.Keys())

	// Add own packages
	packages.Include(ownPackages)

	imports, testImports = MissingPackages(ownPackages, packages)
	assert.Equal(t, []string{
		"example.com/blake",
		"example.com/carey",
	}, imports.Keys())
	assert.Equal(t, []string{
		"example.com/drew/drewutil",
	}, testImports.Keys())

	// Add blake
	packages.Include(blakePackages())

	imports, testImports = MissingPackages(ownPackages, packages)
	assert.Equal(t, []string{
		"example.com/carey",
	}, imports.Keys())
	assert.Equal(t, []string{
		"example.com/drew/drewutil",
	}, testImports.Keys())

	// Add carey
	packages.Include(careyPackages())

	imports, testImports = MissingPackages(ownPackages, packages)
	assert.Equal(t, []string{
		"example.com/carey/internal/carey",
	}, imports.Keys())
	assert.Equal(t, []string{
		"example.com/drew/drewutil",
	}, testImports.Keys())

	// Add drew
	packages.Include(drewPackages())

	imports, testImports = MissingPackages(ownPackages, packages)
	assert.Equal(t, []string{
		"example.com/carey/internal/carey",
	}, imports.Keys())
	assert.Equal(t, []string{
		"example.com/bogus",
	}, testImports.Keys())
}
