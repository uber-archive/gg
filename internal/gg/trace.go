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
	"fmt"
	"io"

	"github.com/RyanCarrier/dijkstra"
	"go.uber.org/multierr"
)

// Trace writes a report of the shortest chain of imports that lead from a
// command or test in the working copy to the given package.
func Trace(out io.Writer, pkg string, ownPackages, packages Packages, from string) error {
	var err error

	packages = packages.Clone()
	packages.Include(ownPackages)

	graph := dijkstra.NewGraph()

	pkgs := []string{
		"root",
		pkg,
		pkg + "_test",
	}
	indexes := map[string]int{
		"root":        0,
		pkg:           1,
		pkg + "_test": 2,
	}

	graph.AddVertex(0)
	graph.AddVertex(1)
	graph.AddVertex(2)
	err = multierr.Append(err, graph.AddArc(2, 0, 0))
	err = multierr.Append(err, graph.AddArc(1, 0, 0))

	add := func(pkg string) {
		if _, ok := indexes[pkg]; ok {
			return
		}
		index := len(pkgs)
		indexes[pkg] = index
		pkgs = append(pkgs, pkg)
		graph.AddVertex(index)
	}

	for pkg := range packages.All {
		add(pkg)
	}
	// All edge weights are similar (base of 64), but package path length is a
	// factor, so the tracer will favor two edges with short names over one
	// edge with a long name.
	for pkg, imps := range packages.Imports {
		pkgIndex := indexes[pkg]
		for imp := range imps {
			impIndex := indexes[imp]
			err = multierr.Append(err, graph.AddArc(impIndex, pkgIndex, 64+int64(len(imp))))
		}
	}
	for pkg, imps := range packages.TestImports {
		pkgIndex := indexes[pkg]
		for imp := range imps {
			impIndex := indexes[imp]
			err = multierr.Append(err, graph.AddArc(impIndex, pkgIndex, 64+int64(len(imp))))
		}
	}

	// Populate the frontier
	for pkg := range ownPackages.Commands {
		index := indexes[pkg]
		err = multierr.Append(err, graph.AddArc(index, 1, 0))
	}
	for pkg := range ownPackages.CoTestImports {
		index := indexes[pkg]
		err = multierr.Append(err, graph.AddArc(index, 2, 0))
	}

	index, ok := indexes[from]
	if !ok {
		fmt.Fprintf(out, "There is no package named %s\n", from)
		return err
	}

	bestPath, shortestErr := graph.Shortest(index, 0)
	if shortestErr != nil {
		fmt.Fprintf(out, "There is no chain of imports that leads from a command or test in the working copy to %s: %s\n", from, err)
		return multierr.Append(err, shortestErr)
	}

	fmt.Fprintf(out, "Shortest import path (each package imports the previous):\n")
	for _, index := range bestPath.Path {
		if index != 0 {
			fmt.Fprintf(out, "* %s\n", pkgs[index])
		}
	}
	return err
}
