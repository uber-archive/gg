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

func TestExtraModules(t *testing.T) {
	// Define working copy
	ownPackages := NewPackages()
	ownPackages.Command("example.com/zed")
	ownPackages.Import("example.com/zed", "example.com/avery")

	// Define example.com/avery
	averyModule := Module{
		Name:     "example.com/avery",
		Packages: NewPackages(),
	}
	averyModule.Packages.Export("example.com/avery")
	averyModule.Packages.Import("example.com/avery", "example.com/blake")

	// Define example.com/blake
	blakeModule := Module{
		Name:     "example.com/blake",
		Packages: NewPackages(),
	}
	blakeModule.Packages.Export("example.com/blake")

	// Define example.com/carey
	careyModule := Module{
		Name:     "example.com/carey",
		Packages: NewPackages(),
	}
	careyModule.Packages.Export("example.com/carey")

	// Whole solution packages
	packages := NewPackages()
	packages.Include(ownPackages)
	packages.Include(averyModule.Packages)
	packages.Include(blakeModule.Packages)
	packages.Include(careyModule.Packages)

	assert.Equal(t, Modules{careyModule}, ExtraModules(ownPackages, packages, Modules{
		averyModule,
		blakeModule,
		careyModule,
	}))
}

func TestShallowSolution(t *testing.T) {
	// Define working copy
	ownPackages := NewPackages()
	ownPackages.Command("example.com/zed")
	ownPackages.Import("example.com/zed", "example.com/avery")
	ownPackages.TestImport("example.com/zed", "example.com/blake")

	// Define example.com/avery
	averyModule := Module{
		Name:     "example.com/avery",
		Packages: NewPackages(),
	}
	averyModule.Packages.Export("example.com/avery")
	averyModule.Packages.Import("example.com/avery", "example.com/blake")

	// Define example.com/blake
	blakeModule := Module{
		Name:     "example.com/blake",
		Packages: NewPackages(),
	}
	blakeModule.Packages.Export("example.com/blake")
	blakeModule.Packages.Import("example.com/blake", "example.com/carey")

	// Define example.com/carey
	careyModule := Module{
		Name:     "example.com/carey",
		Packages: NewPackages(),
	}
	careyModule.Packages.Export("example.com/carey")

	assert.Equal(t, Modules{
		averyModule,
		blakeModule,
	}, ShallowSolution(ownPackages, Modules{
		averyModule,
		blakeModule,
		careyModule,
	}))
}
