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
	yaml "gopkg.in/yaml.v2"
)

func TestGlideManifestModulesMapper(t *testing.T) {
	input := Modules{
		{
			Name:    "avery",
			Version: Version{1, 0, 0},
		},
		{
			Name: "blake",
			Ref:  "heads/master",
		},
		{
			Name:    "carey",
			Version: Version{0, 1, 0},
		},
		{
			Name: "drew",
			Test: true,
		},
	}
	manifest := GlideManifestFromModules(input)
	bytes, err := yaml.Marshal(&manifest)
	require.NoError(t, err)
	assert.Equal(t, `import:
- package: avery
  version: ^1.0.0
- package: blake
  version: master
- package: carey
  version: ~0.1.0
testImport:
- package: drew
`, string(bytes))
}
