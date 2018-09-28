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

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSolver(t *testing.T) {
	loader := NewFakeLoader(Modules{
		{
			Name: "avery",
		},
		{
			Name: "blake",
			Modules: Modules{
				Module{
					Name:    "carey",
					Version: Version{1, 0, 0},
				},
			},
		},
		{
			Name:    "carey",
			Version: Version{1, 0, 0},
		},
		{
			Name:    "carey",
			Version: Version{2, 0, 0},
		},
		{
			Name: "drew",
			Modules: Modules{
				Module{
					Name:    "carey",
					Version: Version{2, 0, 0},
				},
			},
		},
		{
			Name: "evelyn",
			Modules: Modules{
				Module{Name: "finley"},
				Module{Name: "gayle", Test: true},
			},
		},
		{
			Name: "finley",
			Modules: Modules{
				Module{Name: "gayle"},
			},
		},
		{
			Name: "gayle",
		},
		{
			Name: "gayle",
			Test: true,
		},
		{
			Name: "harley",
			Modules: Modules{
				Module{Name: "carey", Version: Version{1, 0, 0}},
				Module{Name: "carey", Version: Version{2, 0, 0}},
			},
		},
		{
			Name: "irene",
			Modules: Modules{
				Module{Name: "carey", Version: Version{1, 0, 0}},
				Module{Name: "jamie"},
				Module{Name: "kelly"},
			},
		},
		{
			Name: "jamie",
			Modules: Modules{
				Module{Name: "carey", Version: Version{1, 0, 0}},
			},
		},
		{
			Name: "kelly",
			Modules: Modules{
				Module{Name: "carey", Version: Version{2, 0, 0}},
			},
		},
		{
			Name: "leslie",
			Modules: Modules{
				Module{Name: "morgan", Test: true},
			},
		},
		{
			Name: "morgan",
			Modules: Modules{
				Module{Name: "carey", Version: Version{1, 0, 0}},
			},
		},
		{
			Name: "nancy",
			Modules: Modules{
				Module{Name: "leslie"},
				Module{Name: "carey", Version: Version{1, 0, 0}},
			},
		},
	})

	tests := []struct {
		name string
		give Modules
		want Modules
	}{
		{
			name: "ashes to ashes",
		},
		{
			name: "dust to dust",
			give: Modules{
				{Name: "avery"},
			},
			want: Modules{
				{Name: "avery"},
			},
		},
		{
			name: "entrain a dependency",
			give: Modules{
				{Name: "blake"},
			},
			want: Modules{
				{Name: "blake"},
				{Name: "carey", Version: Version{1, 0, 0}},
			},
		},
		{
			name: "use the newest of shared dependencies",
			give: Modules{
				{Name: "blake"},
				{Name: "drew"},
			},
			want: Modules{
				{Name: "blake"},
				{Name: "carey", Version: Version{2, 0, 0}},
				{Name: "drew"},
			},
		},
		{
			name: "retain a test dependency",
			give: Modules{
				{Name: "gayle", Test: true},
			},
			want: Modules{
				{Name: "gayle", Test: true},
			},
		},
		{
			name: "promote a test dependency",
			give: Modules{
				{Name: "gayle", Test: true},
				{Name: "gayle"},
			},
			want: Modules{
				{Name: "gayle"},
			},
		},
		{
			name: "promote test dependencies",
			give: Modules{
				{Name: "evelyn"},
			},
			want: Modules{
				{Name: "evelyn"},
				{Name: "finley"},
				{Name: "gayle"},
			},
		},
		{
			name: "reconstrain double dependendcy",
			give: Modules{
				{Name: "harley"},
			},
			want: Modules{
				{Name: "carey", Version: Version{2, 0, 0}},
				{Name: "harley"},
			},
		},
		{
			name: "back track to upgrade",
			give: Modules{
				{Name: "irene"},
			},
			want: Modules{
				{Name: "carey", Version: Version{2, 0, 0}},
				{Name: "irene"},
				{Name: "jamie"},
				{Name: "kelly"},
			},
		},
		{
			name: "dependency of test is a test dependency",
			give: Modules{
				{Name: "leslie"},
			},
			want: Modules{
				{Name: "carey", Version: Version{1, 0, 0}, Test: true},
				{Name: "leslie"},
				{Name: "morgan", Test: true},
			},
		},
		{
			name: "dependency of a test dependency promoted to a normal dependency",
			give: Modules{
				{Name: "nancy"},
			},
			want: Modules{
				{Name: "carey", Version: Version{1, 0, 0}},
				{Name: "leslie"},
				{Name: "morgan", Test: true},
				{Name: "nancy"},
			},
		},
	}

	progress := &LogSolverProgress{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			state := NewState()
			var err error
			state, err = state.Constrain(ctx, loader, progress, tt.give, false)
			require.NoError(t, err)
			state, err = state.Solve(ctx, loader, progress)
			require.NoError(t, err)
			err = loader.FinishModules(ctx, progress, tt.want)
			require.NoError(t, err)
			assert.True(t, tt.want.Equal(state.Modules()))
			_ = state
		})
	}
}

func TestAdd(t *testing.T) {
	ctx := context.Background()

	loader := NewFakeLoader(Modules{
		{Name: "avery"},
	})
	progress := &LogSolverProgress{}
	state := NewState()
	var err error

	state, err = state.Add(ctx, loader, progress, Module{Name: "avery", Test: true})
	require.NoError(t, err)
	assert.Equal(t, Modules{
		loader.MustGetTestVersion("avery", NoVersion),
	}, state.Modules())

	state, err = state.Add(ctx, loader, progress, Module{Name: "avery"})
	require.NoError(t, err)
	want := Modules{
		loader.MustGetVersion("avery", NoVersion),
	}
	assert.Equal(t, want, state.Modules())
}

func TestRemove(t *testing.T) {
	loader := NewFakeLoader(Modules{
		{
			Name: "avery",
		},
		{
			Name: "blake",
		},
		{
			Name: "carey",
			Modules: Modules{
				Module{Name: "drew"},
			},
		},
		{
			Name: "drew",
		},
	})

	progress := &LogSolverProgress{}

	t.Run("remove one package, keep the other", func(t *testing.T) {
		ctx := context.Background()
		state := NewState()
		var err error
		state, err = state.Constrain(ctx, loader, progress, Modules{
			loader.MustGetVersion("avery", NoVersion),
			loader.MustGetVersion("blake", NoVersion),
		}, false)
		require.NoError(t, err)
		state, err = state.Solve(ctx, loader, progress)
		require.NoError(t, err)
		state, err = state.Remove(ctx, loader, progress, "blake")
		require.NoError(t, err)
		expected := Modules{
			loader.MustGetVersion("avery", NoVersion),
		}
		if !assert.True(t, expected.Equal(state.Modules())) {
			fmt.Printf("Expected: %v\n", expected)
			fmt.Printf("Actual:   %v\n", state.Modules())
		}
	})

	t.Run("remove one package, from the frontier", func(t *testing.T) {
		ctx := context.Background()
		state := NewState()
		var err error
		state, err = state.Constrain(ctx, loader, progress, Modules{
			loader.MustGetVersion("avery", NoVersion),
			loader.MustGetVersion("blake", NoVersion),
		}, false)
		require.NoError(t, err)
		state, err = state.Remove(ctx, loader, progress, "blake")
		require.NoError(t, err)
		expected := Modules{
			loader.MustGetVersion("avery", NoVersion),
		}
		if !assert.True(t, expected.Equal(state.Modules())) {
			fmt.Printf("Expected: %v\n", expected)
			fmt.Printf("Actual:   %v\n", state.Modules())
		}
	})

	t.Run("remove one package, and all that depend upon it", func(t *testing.T) {
		ctx := context.Background()
		state := NewState()
		var err error
		state, err = state.Constrain(ctx, loader, progress, Modules{
			{Name: "avery"},
			{Name: "blake"},
			{Name: "carey"},
		}, false)
		require.NoError(t, err)
		state, err = state.Solve(ctx, loader, progress)
		require.NoError(t, err)
		state, err = state.Remove(ctx, loader, progress, "drew")
		require.NoError(t, err)
		expected := Modules{
			loader.MustGetVersion("avery", NoVersion),
			loader.MustGetVersion("blake", NoVersion),
		}
		if !assert.True(t, expected.Equal(state.Modules())) {
			fmt.Printf("Expected: %v\n", expected)
			fmt.Printf("Actual:   %v\n", state.Modules())
		}
		_ = state
	})
}
