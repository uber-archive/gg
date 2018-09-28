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
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

// FSEntry represents an entry in a filesystem directory.
type FSEntry struct {
	path  string
	isDir bool
}

// Name returns the name of the entry.
func (n FSEntry) Name() string {
	return filepath.Base(n.path)
}

// Hash returns NoHash since we have no use for hashes in the working copy.
func (n FSEntry) Hash() plumbing.Hash {
	return NoHash
}

// IsDir returns whether the entry is for a directory.
func (n FSEntry) IsDir() bool {
	return n.isDir
}

// List returns entries of this directory entry.
func (n FSEntry) List() ([]TreeEntry, error) {
	f, err := os.Open(n.path)
	if err != nil {
		return nil, err
	}
	entries, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	list := make([]TreeEntry, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		mode := entry.Mode()
		list = append(list, &FSEntry{
			path:  filepath.Join(n.path, name),
			isDir: mode.IsDir(),
		})
	}
	return list, nil
}

// Reader opens the file addressed by this entry for reading.
func (n FSEntry) Reader() (io.ReadCloser, error) {
	return os.Open(n.path)
}
