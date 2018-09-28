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
	"io"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-billy.v4/osfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

func testWalker(t *testing.T, entry TreeEntry) {
	var paths []string
	var names []string
	walker := Walk("example.com", entry)
	for {
		path, entry, err := walker.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if entry.Name() == "testdata" {
			walker.Skip()
			continue
		}
		paths = append(paths, path)
		names = append(names, entry.Name())
	}
	sort.Strings(paths)
	assert.Equal(t, []string{
		"example.com/example",
		"example.com/example/CHANGELOG.md",
		"example.com/example/Gopkg.lock",
		"example.com/example/glide.lock",
		"example.com/example/internal",
		"example.com/example/internal/exampleutil",
		"example.com/example/internal/exampleutil/exampleutil.go",
		"example.com/example/main.go",
		"example.com/example/main_test.go",
		"example.com/example/mainx_test.go",
	}, paths)
	sort.Strings(names)
	assert.Equal(t, []string{
		"CHANGELOG.md",
		"Gopkg.lock",
		"example",
		"exampleutil",
		"exampleutil.go",
		"glide.lock",
		"internal",
		"main.go",
		"main_test.go",
		"mainx_test.go",
	}, names)
}

func TestFSEntryWalk(t *testing.T) {
	testWalker(t, FSEntry{
		path:  "testdata/src/example.com/example",
		isDir: true,
	})
}

func TestGitEntryWalk(t *testing.T) {
	path := "testdata/src/example.com/example"

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

	testWalker(t, GitEntry{
		name: filepath.Base("example.com/example"),
		mode: filemode.Dir,
		hash: tree.Hash,
		repo: repo,
	})
}
