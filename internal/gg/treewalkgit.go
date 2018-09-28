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

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
)

// GitEntry represents an entry in a git tree, suitable for reading as a
// directory or a file.
type GitEntry struct {
	name string
	mode filemode.FileMode
	repo *git.Repository
	hash plumbing.Hash
}

// Name returns the entry's name.
func (g GitEntry) Name() string {
	return g.name
}

// IsDir returns whether the entry is a directory.
func (g GitEntry) IsDir() bool {
	return g.mode == filemode.Dir
}

// Hash returns the hash of the git object this entry addresses.
func (g GitEntry) Hash() plumbing.Hash {
	return g.hash
}

// Reader reads a blob entry.
func (g GitEntry) Reader() (io.ReadCloser, error) {
	blob, err := g.repo.BlobObject(g.hash)
	if err != nil {
		return nil, err
	}
	return blob.Reader()
}

// List reads a tree entry.
func (g GitEntry) List() ([]TreeEntry, error) {
	tree, err := g.repo.TreeObject(g.hash)
	if err != nil {
		return nil, err
	}
	entries := make([]TreeEntry, 0, len(tree.Entries))
	for _, entry := range tree.Entries {
		entries = append(entries, GitEntry{
			name: entry.Name,
			mode: entry.Mode,
			hash: entry.Hash,
			repo: g.repo,
		})
	}
	return entries, nil
}
