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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddMissing(t *testing.T) {
	ctx := context.Background()

	loader := NewFakeLoader(Modules{
		{
			Name:     "example.com/blake",
			Version:  Version{1, 0, 0},
			Packages: blakePackages(),
		},
		{
			Name:     "example.com/blake",
			Version:  Version{2, 0, 0},
			Packages: blakePackages(),
		},
		{
			Name:     "example.com/carey",
			Version:  Version{1, 0, 0},
			Packages: careyPackages(),
		},
		{
			Name:     "example.com/drew",
			Ref:      "heads/master",
			Packages: drewPackages(),
		},
	})

	progress := &LogSolverProgress{}
	state := NewState()

	name := "example.com/avery"
	packages := averyPackages()

	recommended := map[string]Version{
		"example.com/blake": Version{1, 0, 0},
	}
	next, err := AddMissing(ctx, loader, progress, state, name, packages, recommended)
	require.NoError(t, err)

	modules := Modules{
		{Name: "example.com/blake", Version: Version{1, 0, 0}},
		{Name: "example.com/carey", Version: Version{1, 0, 0}},
		{Name: "example.com/drew", Ref: "heads/master", Test: true},
	}
	err = loader.FinishModules(ctx, progress, modules)
	require.NoError(t, err)

	assert.Equal(t, modules, next.Modules())
}

func TestAddMissingFavorNonTest(t *testing.T) {
	ctx := context.Background()

	loader := NewFakeLoader(Modules{
		{Name: "example.com/blake", Ref: "heads/master"},
		{Name: "example.com/carey", Ref: "heads/master"},
	})
	progress := &LogSolverProgress{}
	state := NewState()

	name := "example.com/avery"
	packages := NewPackages()
	packages.Command("example.com/avery")
	packages.Import("example.com/avery", "example.com/blake")
	packages.Import("example.com/avery", "example.com/carey/command")
	packages.TestImport("example.com/avery", "example.com/carey/test")

	var recommended map[string]Version
	next, err := AddMissing(ctx, loader, progress, state, name, packages, recommended)
	require.NoError(t, err)

	modules := Modules{
		{Name: "example.com/blake"},
		{Name: "example.com/carey"},
	}
	err = loader.FinishModules(ctx, progress, modules)
	require.NoError(t, err)

	assert.Equal(t, modules, next.Modules())
}
