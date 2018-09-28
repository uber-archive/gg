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

// The go-git tree walker implementation is buggy.
// This implementation is less buggy, and works for both OS and Git trees.

import (
	"io"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

// TreeEntry is an entry in a directory, regardless of whether it comes from
// the file system or a Git repository.
type TreeEntry interface {
	Name() string
	IsDir() bool
	Hash() plumbing.Hash
	Reader() (io.ReadCloser, error)
	List() ([]TreeEntry, error)
}

// Walk returns a tree walker that will iterate the given tree and all of its
// descendants in depth-first order.  Dir is the name of the directory
// containing the entry.
func Walk(dir string, entry TreeEntry) *Walker {
	return &Walker{
		stack: walkerStack{
			walkerFrame{
				Prefix:  dir + "/",
				Entries: []TreeEntry{entry},
				Index:   0,
			},
		},
	}
}

// Walker captures the state of a depth-first git tree traversal as a stack
// of frames.
// Each frame corresponds to the next tree entry when the walker returns to
// that path depth.
type Walker struct {
	stack walkerStack
}

// Next advances the walker and returns the next tree entry and its full path
// in the depth first traversal.
// Next returns an io.EOF error to indicate the end of the traversal.
func (w *Walker) Next() (string, TreeEntry, error) {
	stack, path, entry, err := w.stack.Next()
	w.stack = stack
	return path, entry, err
}

// Skip advances the walker to the next child of the parent entry.
// Calling skip after processing a directory entry will skip that directory's
// descendants.
func (w *Walker) Skip() {
	w.stack = w.stack.Skip()
}

// walkerFrame represents the next entry to reveal, or whether the current
// branch has no further children.
type walkerFrame struct {
	// Prefix is the full path leading to current tree.
	Prefix string
	// Entries are the entries of the current tree.
	Entries []TreeEntry
	// Index is the index of the next entry in the current tree.
	// If Index is the length of the current tree's entries,
	// it indicates that the next entry to reveal is up the stack or we have
	// reached the end of the traversal.
	Index int
}

// walkerStack is a slice of frames.
// Operations on this slice return a new slice.
// Walker trampoulines these slice mutations.
type walkerStack []walkerFrame

func (w walkerStack) Next() (walkerStack, string, TreeEntry, error) {
	// An empty stack indicates that we have reached the end of the traversal.
	if len(w) == 0 {
		return nil, "", nil, io.EOF
	}
	top := w[len(w)-1]

	// If there are no further child entries at this frame, we recurse into the
	// parent frame until we find a parent that has further children, or
	// terminate at the base of the stack.
	if top.Index >= len(top.Entries) {
		return w[:len(w)-1].Next()
	}

	// Otherwise, we advance the frame at the current stack depth to the next
	// child of this branch of the tree.
	entry := top.Entries[top.Index]
	path := top.Prefix + entry.Name()
	next := append(w[:len(w)-1], walkerFrame{
		Prefix:  top.Prefix,
		Entries: top.Entries,
		Index:   top.Index + 1,
	})
	// If this child is a subtree, we will push it onto the stack for the next
	// iteration.
	if entry.IsDir() {
		entries, err := entry.List()
		if err != nil {
			return next, path, entry, err
		}
		next = append(next, walkerFrame{
			Prefix:  path + "/",
			Entries: entries,
			Index:   0,
		})
	}
	// We then present the entry for the subtree or leaf entry.
	// The user may elect to skip the next subtree by calling Skip().
	return next, path, entry, nil
}

// Skip advances the walker past a subtree (to the next child of the parent).
func (w walkerStack) Skip() walkerStack {
	if len(w) > 0 {
		return w[:len(w)-1]
	}
	return w
}
