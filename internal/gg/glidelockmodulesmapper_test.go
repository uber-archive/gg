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
	"github.com/stretchr/testify/require"
)

func TestGlideLockModulesMapper(t *testing.T) {
	input := Modules{
		{
			Name:     "avery",
			Hash:     averyHash,
			Version:  Version{1, 0, 0},
			Packages: averyPackages(),
			NoLock:   true,
		},
		{
			Name: "blake",
			Hash: blakeHash,
			Ref:  "heads/master",
			Modules: Modules{
				{Name: "carey"},
			},
			Packages: blakePackages(),
		},
		{
			Name:     "carey",
			Hash:     careyHash,
			Packages: careyPackages(),
		},
		{
			Name:                  "drew",
			Hash:                  drewHash,
			Packages:              drewPackages(),
			GitoliteMirror:        true,
			GitoliteMirrorCreated: true,
			Test: true,
		},
	}
	lock := GlideLockFromModules(input)
	output, err := ModulesFromGlideLock(lock)
	require.NoError(t, err)
	assert.Equal(t, output, input)
}
