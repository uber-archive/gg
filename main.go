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

// Package gg contains the gg command.
//
// The gg dependency management tool is a suite of commands built on the
// premise that the data in the glide.lock and Gopkg.lock formats, in
// combination with the metadata in git repositories, is sufficient to create a
// deterministic total order for all of a project's dependencies.
// With git, that total order, and insights from vgo's dependency resolver,
// which only takes into account the exact versions in a projects' transitive
// dependency's lockfiles, we can greatly improve the speed and predictability
// of common dependency management tasks.
//
// We can also provide a richer experience for dependency management that
// provides tools for introspecting a dependency solution and alternatives,
// heuristics for adding missing modules, pruning extra modules, and upgrading
// known dependencies.
//
// Lockfiles address git revisions with their commit or tag hash.
// From that hash we can infer a commit timestamp, a branch name or version by
// a reverse-lookup of known references in the package's git repository, and
// even read its entire import graph.  We can also read its own lockfiles and
// changelog.  From this we can generate a richer lockfile that caches all
// of this metadata and load it all very quickly for later dependency
// management tasks.
package gg

import (
	"fmt"
	"os"

	"github.com/uber/gg/internal/gg"
)

func main() {
	err := gg.Main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
