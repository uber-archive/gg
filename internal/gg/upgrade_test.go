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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LogUpdaterProgress struct {
	LogSolverProgress
}

func TestUpgrade(t *testing.T) {
	loader := NewFakeLoader(Modules{
		{
			Name: "avery",
		},
		{
			Name:    "avery",
			Version: Version{1, 0, 0},
		},
		{
			Name:    "avery",
			Version: Version{1, 1, 0},
		},
		{
			Name:    "avery",
			Version: Version{2, 0, 0},
		},
		{
			Name: "blake",
			Ref:  "heads/master",
			Time: time.Unix(0, 0),
		},
		{
			Name: "blake",
			Ref:  "heads/master",
			Time: time.Unix(0, 0).Add(24 * time.Hour),
		},
	})

	tests := []struct {
		name string
		give Modules
		want Modules
	}{
		{
			name: "ex nihilo nihil fit",
		},
		{
			name: "avery 1.0 to 1.1 but not 2.0",
			give: Modules{
				{Name: "avery", Version: Version{1, 0, 0}},
			},
			want: Modules{
				loader.MustGetVersion("avery", Version{1, 1, 0}),
			},
		},
		{
			name: "avery remains avery",
			give: Modules{
				{Name: "avery"},
			},
			want: Modules{
				loader.MustGetVersion("avery", NoVersion),
			},
		},
		{
			name: "master branch upgrades based on timestamp",
			give: Modules{
				loader.MustGetTime("blake", time.Unix(0, 0)),
			},
			want: Modules{
				loader.MustGetTime("blake", time.Unix(0, 0).Add(24*time.Hour)),
			},
		},
		{
			name: "master branch does not downgrade based on timestamp",
			give: Modules{
				loader.MustGetTime("blake", time.Unix(0, 0).Add(24*time.Hour)),
			},
			want: Modules{
				loader.MustGetTime("blake", time.Unix(0, 0).Add(24*time.Hour)),
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
			state, err = Upgrade(ctx, loader, progress, state)
			require.NoError(t, err)
			assert.True(t, tt.want.Equal(state.Modules()))
		})
	}
}
