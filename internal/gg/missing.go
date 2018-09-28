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
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	// AddMissingFetchMaxAttempts is the number of attempts to perform a git
	// fetch while running the add-missing workflow.
	// This is lower than the default because we expect to fail when we
	// misguess the repository remote location.
	AddMissingFetchMaxAttempts = 2
)

// AddMissingLoader finds and loads modules that satisfy missing imported
// packages.
type AddMissingLoader interface {
	SolverLoader

	Fetch(context.Context, ProgressWriter, *Module, int) error
	FinishRemote(context.Context, ProgressWriter, *Module) error
	FinishPackages(context.Context, ProgressWriter, Modules) error
	DigestRefs(context.Context, ProgressWriter, Module) error
	ReadVersions(context.Context, ProgressWriter, Module) (Modules, error)
}

// AddMissingProgress provides progress notifications while adding missing
// modules.
type AddMissingProgress interface {
	SolverProgress
}

// AddMissing takes a state and returns a new state that makes progress toward
// ensuring that all packages imported in a dependency solution are exported by
// a module in the solution.
// The adder accepts the name of the working copy's root package, to avoid
// adding itself to the dependency graph, and the import graph of the packages
// in the working copy.
// The adder will use the most recent semantic version of any module it can
// find that will satisfy an import.
// The adder does not attempt to isolate the semantic version that provides
// compatible types for the packages that import a package.
// Otherwise, the adder will use the master branch.
// The adder will visit every transitive dependency in the solution, even
// as it adds modules to the solution.
func AddMissing(ctx context.Context, loader AddMissingLoader, out AddMissingProgress, state *State, name string, packages Packages, recommended map[string]Version) (*State, error) {
	tried := make(StringSet)
	out.Start("Adding modules for missing packages")

	modules := state.Modules()
	if err := loader.FinishPackages(ctx, out, modules); err != nil {
		return nil, err
	}
	max := maxExports(modules, 0)
	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		modules := state.Modules()
		if err := loader.FinishPackages(ctx, out, modules); err != nil {
			return nil, err
		}
		max = maxExports(modules, max)
		imports, testImports := MissingPackages(packages, modules.Packages())
		missingProgress(out, state, modules, imports, testImports, max, start)
		if next, ok := addOneMissingModule(ctx, loader, out, state, tried, name, imports, false, recommended); ok {
			state = next
			continue
		}

		modules = state.Modules()
		if err := loader.FinishPackages(ctx, out, modules); err != nil {
			return nil, err
		}
		max = maxExports(modules, max)
		imports, testImports = MissingPackages(packages, modules.Packages())
		missingProgress(out, state, modules, imports, testImports, max, start)
		if next, ok := addOneMissingModule(ctx, loader, out, state, tried, name, testImports, true, recommended); ok {
			state = next
			continue
		}
		break
	}

	out.Stop("Adding modules for missing packages")
	return state, nil
}

func maxExports(modules Modules, max int) int {
	count := len(modules.Packages().Exports)
	if count > max {
		return count
	}
	return max
}

func missingProgress(out AddMissingProgress, state *State, modules Modules, imports, testImports StringSet, max int, start time.Time) {
	// Compute progress indicator
	union := imports.Clone()
	union.Include(testImports)
	tot := len(modules.Packages().Exports) - max
	num := tot - len(union)
	now := time.Now()
	out.Progress("Adding modules for missing packages", num, tot, start, now)
}

func addOneMissingModule(ctx context.Context, loader AddMissingLoader, out AddMissingProgress, state *State, tried StringSet, ownPackage string, packages StringSet, test bool, recommended map[string]Version) (*State, bool) {
Scan:
	for _, name := range packages.Keys() {
		if tried.Has(name) {
			continue Scan
		}
		tried.Add(name)
		fmt.Fprintf(out, "Searching for module to export package %s.\n", name)

		// Skip packages underlying the working copy package.
		if name == ownPackage || strings.HasPrefix(name, ownPackage+"/") {
			continue Scan
		}

		parts := strings.Split(name, "/")

		// Skip any package that *should* be exported by a module in the
		// solution.
		for i := len(parts) - 1; i >= 0; i-- {
			short := strings.Join(parts[0:i], "/")
			if index, ok := state.Problem[short]; ok {
				fmt.Fprintf(out, "* Package %s should be exported by module %s.\n", name, state.Frontier[index].Name)
				continue Scan
			}
			if solution, ok := state.Solution[short]; ok {
				fmt.Fprintf(out, "* Package %s should be exported by module %s.\n", name, solution.Module.Name)
				continue Scan
			}
		}

		for len(parts) >= 2 {
			name = strings.Join(parts, "/")

			module := Module{
				Name: name,
				Test: test,
			}
			if err := loader.FinishRemote(ctx, out, &module); err != nil {
				fmt.Fprintf(out, "Error while finding remote for %s: %s\n", module.Name, err)
			}
			if err := loader.Fetch(ctx, out, &module, AddMissingFetchMaxAttempts); err != nil {
				fmt.Fprintf(out, "Error while fetching versions for %s: %s\n", module.Name, err)
			}
			if err := loader.DigestRefs(ctx, out, module); err != nil {
				fmt.Fprintf(out, "Error while digesting reference for %s: %s\n", module.Name, err)
			}
			versions, err := loader.ReadVersions(ctx, out, module)
			if err != nil {
				fmt.Fprintf(out, "Error while reading versions for %s: %s\n", module.Name, err)
				continue Scan
			}

			var ok bool
			var add Module
			if version := recommended[module.Name]; version != NoVersion {
				add, ok = versions.FindVersion(version)
			} else {
				add, ok = versions.FindBestVersion()
			}

			if ok {
				if next, err := state.Add(ctx, loader, out, add); err != nil {
					fmt.Fprintf(out, "%s\n", err)
				} else {
					fmt.Fprintf(out, "+ %s\n", add.String())
					state = next
				}
				// We return instead of continue because this function should
				// only advance one package forward from the set of missing
				// packages, so we can provide progress notifications for every
				// added module.
				return state, true
			}

			if module.ExactRemote {
				continue Scan
			}

			fmt.Fprintf(out, "Could not find a suitable version for %s.\n", name)
			parts = parts[:len(parts)-1]
			if len(parts) < 2 {
				continue Scan
			}
			fmt.Fprintf(out, "Trying a shorter package name: %s.\n", strings.Join(parts, "/"))
		}
	}
	return state, false
}
