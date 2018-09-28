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
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-billy.v4/osfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

func TestReadOwnPackagesExample(t *testing.T) {
	name, packages, err := ReadOwnPackages(ioutil.Discard, "testdata/src/example.com/example", []string{"testdata"}, nil)
	require.NoError(t, err)

	assert.Equal(t, "example.com/example", name)

	commands := make(StringSet)
	commands.Add("example.com/example")
	assert.Equal(t, commands, packages.Commands)

	exports := make(StringSet)
	exports.Add("example.com/example/internal/exampleutil")
	assert.Equal(t, exports, packages.Exports)

	imports := NewStringGraph()
	imports.Add("example.com/example", "example.com/example/internal/exampleutil")
	assert.Equal(t, imports, packages.Imports)

	coImports := NewStringGraph()
	coImports.Add("example.com/example/internal/exampleutil", "example.com/example")
	assert.Equal(t, coImports, packages.CoImports)

	testImports := NewStringGraph()
	testImports.Add("example.com/example", "example.com/exampletest")
	testImports.Add("example.com/example", "example.com/examplextest")
	assert.Equal(t, testImports, packages.TestImports)

	coTestImports := NewStringGraph()
	coTestImports.Add("example.com/exampletest", "example.com/example")
	coTestImports.Add("example.com/examplextest", "example.com/example")
	assert.Equal(t, coTestImports, packages.CoTestImports)
}

func TestReadOwnPackagesExampleTest(t *testing.T) {
	name, packages, err := ReadOwnPackages(ioutil.Discard, "testdata/src/example.com/examplextest", []string{"testdata"}, nil)
	require.NoError(t, err)

	assert.Equal(t, "example.com/examplextest", name)

	exports := make(StringSet)
	exports.Add("example.com/examplextest")
	assert.Equal(t, exports, packages.Exports)

	commands := make(StringSet)
	assert.Equal(t, commands, packages.Commands)

	imports := NewStringGraph()
	imports.Add("example.com/examplextest", "example.com/exampletest")
	assert.Equal(t, imports, packages.Imports)

	coImports := NewStringGraph()
	coImports.Add("example.com/exampletest", "example.com/examplextest")
	assert.Equal(t, coImports, packages.CoImports)
}

func TestReadOwnPackagesAllArchitectures(t *testing.T) {
	// This test validates that the UseAllFiles directive to the go build context
	// does in fact collect imports from all files in a package, regardless of
	// their build tags.
	name, packages, err := ReadOwnPackages(ioutil.Discard, "testdata/src/example.com/examplearch", []string{"testdata"}, nil)
	require.NoError(t, err)

	assert.Equal(t, "example.com/examplearch", name)

	exports := make(StringSet)
	exports.Add("example.com/examplearch")
	assert.Equal(t, exports, packages.Exports)

	commands := make(StringSet)
	assert.Equal(t, commands, packages.Commands)

	imports := NewStringGraph()
	imports.Add("example.com/examplearch", "example.com/exampleunix")
	assert.Equal(t, imports, packages.Imports)

	coImports := NewStringGraph()
	coImports.Add("example.com/exampleunix", "example.com/examplearch")
	assert.Equal(t, coImports, packages.CoImports)
}

func TestReadOwnPackagesNoGo(t *testing.T) {
	name, packages, err := ReadOwnPackages(ioutil.Discard, "testdata/src/example.com", []string{"testdata"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "example.com", name)
	_ = packages
}

func testReadGitPackages(t *testing.T, name, path string) (Packages, Module) {
	store := memory.NewStorage()
	fs := osfs.New(path)
	repo, err := git.Init(store, fs)
	require.NoError(t, err)

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add(".")
	require.NoError(t, err)
	hash, err := worktree.Commit("First", &git.CommitOptions{
		Author: &object.Signature{
			Name: "Robotto Botdroid",
		},
	})
	require.NoError(t, err)

	commit, err := repo.CommitObject(hash)
	require.NoError(t, err)

	tree, err := repo.TreeObject(commit.TreeHash)
	require.NoError(t, err)

	module := Module{
		Name: name,
	}
	err = ReadGitPackages(ioutil.Discard, repo, tree, &module)
	require.NoError(t, err)

	return module.Packages, module
}

func TestReadGitPackages(t *testing.T) {
	packages, module := testReadGitPackages(t, "example.com/example", "testdata/src/example.com/example")
	_ = module

	commands := make(StringSet)
	commands.Add("example.com/example")
	assert.Equal(t, commands, packages.Commands)

	exports := make(StringSet)
	exports.Add("example.com/example/internal/exampleutil")
	assert.Equal(t, exports, packages.Exports)

	imports := NewStringGraph()
	imports.Add("example.com/example", "example.com/example/internal/exampleutil")
	assert.Equal(t, imports, packages.Imports)

	coImports := NewStringGraph()
	coImports.Add("example.com/example/internal/exampleutil", "example.com/example")
	assert.Equal(t, coImports, packages.CoImports)
}

func TestReadGitPackagesWrongName(t *testing.T) {
	packages, module := testReadGitPackages(t, "example.com/bogus", "testdata/src/example.com/exampletest")
	_ = packages

	assert.Equal(t, []string{
		`The package "example.com/bogus" must be imported as "example.com/exampletest" according to its package import comment.`,
	}, module.Warnings)
}
