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
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func (module Module) LoaderHash() plumbing.Hash {
	return plumbing.ComputeHash(
		plumbing.CommitObject,
		[]byte(
			module.Name+"@"+
				module.Version.String()+"@"+
				strconv.Itoa(int(module.Time.Unix())),
		),
	)
}

var alphaRevisions = Modules{
	{
		Name: "example.com/alpha",
		Ref:  "heads/feature",
		Hash: oneHash,
	},
	{
		Name: "example.com/alpha",
		Ref:  "heads/master",
		Hash: twoHash,
	},
	{
		Name:    "example.com/alpha",
		Version: Version{0, 0, 1},
		Ref:     "tags/v0.0.1",
		Time:    time.Unix(0, 0),
	},
	{
		Name:    "example.com/alpha",
		Version: Version{0, 0, 1},
		Ref:     "tags/v0.0.1",
		Time:    time.Unix(0, 1),
	},
	{
		Name:    "example.com/alpha",
		Version: Version{1, 0, 0},
		Ref:     "tags/v1.0.0",
		Hash:    averyHash,
	},
	{
		Name:    "example.com/alpha",
		Version: Version{2, 0, 0},
		Ref:     "tags/v2.0.0",
		Hash:    blakeHash,
	},
}

func TestModuleString(t *testing.T) {
	tests := []struct {
		msg  string
		give Module
		want string
	}{
		{
			msg:  "empty",
			want: "########                       ?     -",
		},
		{
			msg: "full",
			give: Module{
				Hash:      averyHash,
				Name:      "example.com/example",
				Version:   Version{1234, 567, 890},
				Time:      time.Unix(0, 0),
				Glidelock: averyHash,
				Deplock:   averyHash,
				Changelog: averyHash,
				Test:      true,
			},
			want: "a28ced3c 1969-12-31 1234.567.890 TGC example.com/example",
		},
		{
			msg: "tag",
			give: Module{
				Hash:      averyHash,
				Name:      "example.com/example",
				Ref:       "tags/v1.0.0-rc01",
				Time:      time.Unix(0, 0),
				Glidelock: averyHash,
				Deplock:   averyHash,
				Changelog: averyHash,
				Test:      true,
			},
			want: "a28ced3c 1969-12-31  v1.0.0-rc01 TGC example.com/example",
		},
		{
			msg: "tag overflowing",
			give: Module{
				Hash:      averyHash,
				Name:      "example.com/example",
				Ref:       "tags/v1234.567.890-rc01",
				Time:      time.Unix(0, 0),
				Glidelock: averyHash,
				Deplock:   averyHash,
				Changelog: averyHash,
				Test:      true,
			},
			want: "a28ced3c 1969-12-31 v1234.567.89 TGC example.com/example",
		},
		{
			msg: "head",
			give: Module{
				Hash:      averyHash,
				Name:      "example.com/example",
				Ref:       "heads/master",
				Time:      time.Unix(0, 0),
				Glidelock: averyHash,
				Deplock:   averyHash,
				Changelog: averyHash,
				Test:      true,
			},
			want: "a28ced3c 1969-12-31       master TGC example.com/example",
		},
		{
			msg: "refs",
			give: Module{
				Hash:      averyHash,
				Name:      "example.com/example",
				Version:   Version{1234, 567, 890},
				Time:      time.Unix(0, 0),
				Glidelock: averyHash,
				Deplock:   averyHash,
				Changelog: averyHash,
				Test:      true,
				Refs:      []string{"a", "b", "c"},
			},
			want: "a28ced3c 1969-12-31 1234.567.890 TGC example.com/example (a b c)",
		},
		{
			msg: "deplock",
			give: Module{
				Hash:      averyHash,
				Name:      "example.com/example",
				Version:   Version{1234, 567, 890},
				Time:      time.Unix(0, 0),
				Deplock:   careyHash,
				Changelog: averyHash,
			},
			want: "a28ced3c 1969-12-31 1234.567.890  DC example.com/example",
		},
		{
			msg: "warning",
			give: Module{
				Hash:    averyHash,
				Name:    "example.com/example",
				Version: Version{1234, 567, 890},
				Time:    time.Unix(0, 0),
				Warnings: []string{
					"warning!",
				},
			},
			want: "a28ced3c 1969-12-31 1234.567.890     example.com/example " + yellow + "(warning)" + clear,
		},
		{
			msg: "warnings",
			give: Module{
				Hash:    averyHash,
				Name:    "example.com/example",
				Version: Version{1234, 567, 890},
				Time:    time.Unix(0, 0),
				Warnings: []string{
					"warning!",
					"danger!",
				},
			},
			want: "a28ced3c 1969-12-31 1234.567.890     example.com/example " + yellow + "(2 warnings)" + clear,
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.give.String())
		})
	}
}

func TestMonotonicModules(t *testing.T) {
	order := Modules{
		// Next two differ by hash only
		{
			Name:    "example.com/alpha",
			Version: Version{0, 0, 1},
			Time:    time.Unix(1, 0),
			Hash:    oneHash,
		},
		// Next two differ by timestamp and hash, timestamp preceeds.
		{
			Name:    "example.com/alpha",
			Version: Version{0, 0, 1},
			Time:    time.Unix(1, 0),
			Hash:    twoHash,
		},
		// Next two differ by timestamp and hash, timestamp preceeds.
		// Negative is very positive unsigned.
		{
			Name:    "example.com/alpha",
			Version: Version{0, 0, 1},
			Time:    time.Unix(1, 0),
			Hash:    minusOneHash, // ffff
		},
		{
			Name:    "example.com/alpha",
			Version: Version{0, 0, 1},
			Time:    time.Unix(2, 0),
		},
		// The next three are sequential by version, but not by time.
		{
			Name:    "example.com/alpha",
			Version: Version{1, 0, 0},
			Time:    time.Unix(10, 0),
		},
		{
			Name:    "example.com/alpha",
			Version: Version{1, 0, 1},
			Time:    time.Unix(30, 0),
		},
		{
			Name:    "example.com/alpha",
			Version: Version{2, 0, 0},
			Time:    time.Unix(20, 0),
		},
		// The next two differ by package name, even though time is earlier.
		{
			Name:    "example.com/beta",
			Version: Version{0, 0, 1},
			Time:    time.Unix(2, 0),
		},
		{
			Name:    "example.com/gamma",
			Version: Version{0, 0, 1},
			Time:    time.Unix(1, 0),
		},
	}
	for _, module := range order {
		assert.False(t, module.Before(module))
	}
	for i := 0; i < len(order)-1; i++ {
		module := order[i]
		nodule := order[i+1]
		assert.True(t, module.Before(nodule))
		assert.False(t, nodule.Before(module))
	}

	modules := make(Modules, len(order))
	copy(modules, order)
	sort.Sort(modules)
	assert.Equal(t, modules, order)
}

func TestModulesEqual(t *testing.T) {
	modules := Modules{
		{
			Name: "example.com/alpha",
		},
		{
			Name: "example.com/beta",
		},
	}
	assert.True(t, modules.Equal(modules))

	nodules := append(modules, Module{
		Name: "example.com/gamma",
	})
	assert.False(t, modules.Equal(nodules))

	nodules = nodules[1:]
	assert.False(t, modules.Equal(nodules))
}

func TestModuleCanUpgradeTo(t *testing.T) {
	tests := []struct {
		msg        string
		this, that Module
		want       bool
	}{
		{
			msg:  "degenerate case",
			this: Module{},
			that: Module{},
			want: false,
		},
		{
			msg: "upgrade in time, but no common ref",
			this: Module{
				Time: time.Unix(1, 0),
			},
			that: Module{
				Time: time.Unix(2, 0),
			},
			want: false,
		},
		{
			msg: "upgrade in time, no ref to master",
			this: Module{
				Time: time.Unix(1, 0),
			},
			that: Module{
				Time: time.Unix(2, 0),
				Ref:  "heads/master",
			},
			want: true,
		},
		{
			msg: "downgrade in time on same branch",
			this: Module{
				Time: time.Unix(2, 0),
				Ref:  "heads/master",
			},
			that: Module{
				Time: time.Unix(1, 0),
				Ref:  "heads/master",
			},
			want: false,
		},
		{
			msg: "upgrade in time, refs do not match",
			this: Module{
				Time: time.Unix(1, 0),
				Ref:  "heads/feature",
			},
			that: Module{
				Time: time.Unix(2, 0),
				Ref:  "heads/phitur",
			},
			want: false,
		},
		{
			msg: "upgrade in time, downgrade by version",
			this: Module{
				Version: Version{0, 0, 2},
				Time:    time.Unix(1, 0),
			},
			that: Module{
				Version: Version{0, 0, 1},
				Time:    time.Unix(2, 0),
			},
			want: false,
		},
		{
			msg: "upgrade major",
			this: Module{
				Version: Version{1, 2, 3},
			},
			that: Module{
				Version: Version{2, 0, 0},
			},
			want: false,
		},
		{
			msg: "upgrade patch in same major",
			this: Module{
				Version: Version{1, 2, 3},
			},
			that: Module{
				Version: Version{1, 2, 4},
			},
			want: true,
		},
		{
			msg: "upgrade minor in zero major",
			this: Module{
				Version: Version{0, 1, 2},
			},
			that: Module{
				Version: Version{0, 2, 1},
			},
			want: false,
		},
		{
			msg: "upgrade patch in zero major",
			this: Module{
				Version: Version{0, 1, 2},
			},
			that: Module{
				Version: Version{0, 1, 3},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			assert.Equal(t, tt.this.CanUpgradeTo(tt.that), tt.want)
		})
	}
}

func TestIndexModules(t *testing.T) {
	modules := Modules{
		{
			Name: "example.com/alpha",
		},
		{
			Name: "example.com/beta",
		},
	}
	byName := modules.Index()
	assert.Equal(t, map[string]Module{
		"example.com/alpha": modules[0],
		"example.com/beta":  modules[1],
	}, byName)
}

func TestFindReference(t *testing.T) {
	modules := alphaRevisions

	module, ok := modules.FindReference("heads/master")
	require.True(t, ok)
	assert.Equal(t, "heads/master", module.Ref)

	module, ok = modules.FindReference("heads/bogus")
	assert.False(t, ok)
}

func TestFindHash(t *testing.T) {
	modules := alphaRevisions

	module, ok := modules.FindHash(twoHash, twoHash)
	require.True(t, ok)
	assert.Equal(t, twoHash, module.Hash)

	module, ok = modules.FindHash(careyHash, careyHash)
	assert.False(t, ok)
}

func TestFindVersion(t *testing.T) {
	modules := alphaRevisions

	module, ok := modules.FindVersion(Version{2, 0, 0})
	require.True(t, ok)
	assert.Equal(t, Version{2, 0, 0}, module.Version)

	module, ok = modules.FindVersion(Version{3, 0, 0})
	assert.False(t, ok)
}

func TestFindBestVersion(t *testing.T) {
	modules := alphaRevisions

	module, ok := modules.FindBestVersion()
	require.True(t, ok)
	assert.Equal(t, Version{2, 0, 0}, module.Version)

	module, ok = modules[0:5].FindBestVersion()
	require.True(t, ok)
	assert.Equal(t, Version{1, 0, 0}, module.Version)

	module, ok = modules[0:4].FindBestVersion()
	require.True(t, ok)
	assert.Equal(t, time.Unix(0, 1), module.Time)

	module, ok = modules[0:2].FindBestVersion()
	require.True(t, ok)
	assert.Equal(t, "heads/master", module.Ref)

	// This last module does not have a version and its branch name is not
	// master, so it is inelligible to fill a missing dependency automatically.
	// It would be an arbitrary choice among given versions.
	module, ok = modules[0:1].FindBestVersion()
	require.False(t, ok)

	module, ok = modules[0:0].FindBestVersion()
	require.False(t, ok)

	module, ok = Modules(nil).FindBestVersion()
	require.False(t, ok)
}

func TestFilterNumberedVersions(t *testing.T) {
	versioned := alphaRevisions.FilterNumberedVersions()
	assert.Len(t, versioned, 4)
}
