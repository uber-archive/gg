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

// This file contains the dependency constraint solver algorithm and the data
// structures for representing its states.  The algorithm works like a
// proof-by-contradiction.
//
// The initial state has a "problem", also called the "frontier", and an empty
// "solution".  The "frontier" is an ordered list of modules that must be in
// the solution at some minimum version.  The "problem" is an unordered index
// of the "frontier".
//
// To use the solver, you need to prime the problem with the shallow
// dependencies of the packages in the working copy, either explicitly from a
// lock file or implicitly by scanning the packages imported by Go files, using
// heuristics to infer which module repositories export those imported
// packages.
//
// In each state, the solver "considers" moving one of the modules in the
// "problem" to the solution, and keeps track of the previous state so it can
// back-track if it finds a constraint that invalidates the assumption that the
// version is suitable for the solution.
//
// The solver tracks which dependencies are only needed for tests, and any
// dependency of a test dependency is implicitily also a test dependency.  The
// solver promotes test dependencies to non-test dependencies as if they were
// an older version.  The solver will back-track to ensure that the transitive
// dependencies of the old test dependency also get promoted to non-test
// dependencies.
//
// When the solver considers a version, it examines each of that version's
// minimum version constraints.  It may add a new constraint to the problem,
// upgrade a version in the existing problem, or invalidate a constraint
// already in the solution.
// In this last case, it will back-track to the state before it considered that
// version, then reapply all of the other constraints accumulated up to this
// point, strictly advancing the solution to newer versions of packages already
// explored.
//
// The solver is finished when it finds a state where the frontier is empty.

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// SolverProgress handles progress notifications from the constraint solver.
type SolverProgress interface {
	ProgressWriter

	ShowState(*State)
	Consider(*State, Module)
	Constrain(*State, Module)
	Backtrack(*State, Module, Module)
}

// SolverLoader ensures that each module has been fetched, then normalizes each
// module by filling in missing fields.
type SolverLoader interface {
	FinishModules(context.Context, ProgressWriter, Modules) error
	FinishModule(context.Context, ProgressWriter, *Module) error
}

// State represents a state of the constraint solver.
// The solver may back-track to a state.
// The state captures a frontier of modules that have not yet been visited to
// ensure that the solution captures their transitive dependencies, a solution
// that captures the modules that have been considered, and the import graph of
// the aggregate solution.
type State struct {
	// Frontier is a slice of modules that should be visited in order until the
	// frontier is empty, indicating that the solution is complete.
	Frontier Modules
	// Problem is an index of the Frontier, mapping package names to indexes in
	// the frontier.
	Problem Problem
	// Solution is a map of package names that we are considering for the
	// completed solution and the state we must return to if the package needs
	// an upgrade.
	// The shallow constraints of every module in the solution are already
	// accounted for either in the frontier or the solution.
	Solution Solution
	// Dependees is a reverse-lookup from module to the modules that depend
	// upon it, necessary for removing a package and its transitive dependees.
	// The Remove method can't just remove one package, since the solver would
	// just add them back if any module in the solution depends on it.
	Dependees StringGraph
}

// NewState returns an empty initial state.
func NewState() *State {
	return &State{
		Problem:   make(Problem),
		Solution:  make(Solution),
		Dependees: make(StringGraph),
	}
}

// Problem is a map of package names to indexes of modules in the frontier.
type Problem map[string]int

// Solution is a map of package names to a module and the state to back-track
// to if that module needs to be upgraded to a newer version, invalidating any
// of the transitive dependencies that have already been discovered.
type Solution map[string]Partial

// Partial indicates the module and the state to return to if the module needs
// an upgrade.
type Partial struct {
	Module Module
	Back   *State
}

// Backtrack walks back to before a module was considered and considers a newer
// version.
func (partial Partial) Backtrack(module Module) *State {
	return partial.Back.Consider(module)
}

// Modules returns all of the modules in the solution.
func (solution Solution) Modules() Modules {
	modules := make(Modules, 0, len(solution))
	for _, partial := range solution {
		modules = append(modules, partial.Module)
	}
	return modules
}

// Names returns all of the module package names in the solution.
func (solution Solution) Names() []string {
	packages := make([]string, 0, len(solution))
	for name := range solution {
		packages = append(packages, name)
	}
	sort.Strings(packages)
	return packages
}

// Has returns whether the solution has a module with the given package name.
func (state *State) Has(name string) bool {
	if _, ok := state.Problem[name]; ok {
		return true
	}
	if _, ok := state.Solution[name]; ok {
		return true
	}
	return false
}

// Solve is a state trampoline that moves modules from the problem to the
// solution, considering each module and adding its constraints, until the
// it exhausts the frontier.
func (state *State) Solve(ctx context.Context, loader SolverLoader, out SolverProgress) (*State, error) {
	start := time.Now()

	var err error
	for len(state.Frontier) > 0 {
		out.ShowState(state)
		consider := state.Frontier[0]

		status := fmt.Sprintf("Considering %s", consider.Summary())
		out.Start(status)
		state = state.Consider(consider)
		state, err = state.Constrain(ctx, loader, out, consider.Modules, consider.Test)
		out.Stop(status)

		// Progress indicator
		num := len(state.Solution)
		den := len(state.Frontier) + num
		now := time.Now()
		out.Progress("Solving dependency graph", num, den, start, now)

		if err != nil {
			return state, err
		}
	}
	out.ShowState(state)
	return state, nil
}

// Add is a shorthand for adding a single constraint and re-running the solver
// to completion.
func (state *State) Add(ctx context.Context, loader SolverLoader, out SolverProgress, module Module) (*State, error) {
	if err := loader.FinishModule(ctx, out, &module); err != nil {
		return nil, err
	}
	return state.lock(loader, out, module).Solve(ctx, loader, out)
}

// lock not only adds a module to the state, but retroactively adds it to the
// original frontier, ensuring that a version of the module exists even if we
// back-track to a state prior to the lock.
func (state *State) lock(loader SolverLoader, out SolverProgress, choice Module) *State {
	frontier := make(Modules, 0, len(state.Frontier))
	problem := make(Problem, len(state.Problem))
	solution := make(Solution, len(state.Solution))
	dependees := state.Dependees.Clone()

	added := false

	for _, module := range state.Frontier {
		if module.Name == choice.Name {
			if choice.Before(module) {
				problem[module.Name] = len(frontier)
				frontier = append(frontier, module)
				added = true
			}
		} else {
			problem[module.Name] = len(frontier)
			frontier = append(frontier, module)
		}
	}

	for name, partial := range state.Solution {
		module := partial.Module
		if module.Name == choice.Name {
			if choice.Before(module) {
				problem[module.Name] = len(frontier)
				frontier = append(frontier, module)
				added = true
			}
		} else {
			solution[name] = Partial{
				Module: partial.Module,
				Back:   partial.Back.Consider(choice),
			}
		}
	}

	if !added {
		problem[choice.Name] = len(frontier)
		frontier = append(frontier, choice)
	}

	return &State{
		Frontier:  frontier,
		Problem:   problem,
		Solution:  solution,
		Dependees: dependees,
	}
}

// Constrain returns a new state with all of the given modules added to the
// frontier, if it is not already in the solution with a better version.
// If necessary, constrain will back-track to a prior solution to upgrade a
// module that is already in the solution.
func (state *State) Constrain(ctx context.Context, loader SolverLoader, out SolverProgress, modules Modules, test bool) (*State, error) {
	var err error
	if err = loader.FinishModules(ctx, out, modules); err != nil {
		return state, err
	}

	// Back-track for any module that we have already considered for the
	// solution but need to upgrade.
	for _, module := range modules {
		out.Constrain(state, module)
		if partial, ok := state.Solution[module.Name]; ok {
			if partial.Module.Before(module) {
				out.Backtrack(state, partial.Module, module)
				// Back-track to before the prior version was locked,
				// so we can forget modules and packages that might not exist
				// any longer.
				// Replace the old module with the new.
				state = partial.Backtrack(module)
			}
		}
	}

	// Add all modules to the frontier.
	state = state.Clone()
	for _, module := range modules {
		if test {
			module.Test = true
		}
		if i, ok := state.Problem[module.Name]; ok {
			problem := state.Frontier[i]
			// Promote test modules to normal modules if they are depended upon
			// by a non-test module.
			if !module.Test {
				problem.Test = false
			}
			// Promote to the higher or newer of two versions.
			if problem.Before(module) {
				problem = module
			}
			state.Frontier[i] = problem
		} else {
			solution, ok := state.Solution[module.Name]
			if ok && !module.Test {
				solution.Module.Test = false
				state.Solution[module.Name] = solution
			}
			if !ok || solution.Module.Before(module) {
				// Append modules to the frontier either if we have not already
				// seen them (not in the existing problem or solution), or if they
				// are already in the solution but at an earlier version.
				delete(state.Solution, module.Name)
				state.Problem[module.Name] = len(state.Frontier)
				state.Frontier = append(state.Frontier, module)
			}
		}
	}

	return state, nil
}

// Clone duplicates the state.
func (state *State) Clone() *State {
	return state.Consider(Module{})
}

// Consider creates a new state assuming that the given module will be in the
// solution (removes it from the frontier and problem), and hooks it up to
// backtrack to this state if it's replaced later.
func (state *State) Consider(choice Module) *State {
	frontier := make(Modules, 0, len(state.Frontier))
	problem := make(Problem, len(state.Problem))
	solution := make(Solution, len(state.Solution))
	dependees := state.Dependees.Clone()

	for _, module := range state.Frontier {
		if choice.Name != module.Name {
			problem[module.Name] = len(frontier)
			frontier = append(frontier, module)
		}
	}

	for name, partial := range state.Solution {
		if name != choice.Name {
			solution[name] = Partial{
				Module: partial.Module,
				Back:   partial.Back,
			}
		}
	}

	if choice.Name != "" {
		solution[choice.Name] = Partial{
			Module: choice,
			Back:   state,
		}
		for _, module := range choice.Modules {
			dependees.Add(module.Name, choice.Name)
		}
	}

	return &State{
		Frontier:  frontier,
		Problem:   problem,
		Solution:  solution,
		Dependees: dependees,
	}
}

// Remove removes a module from the solution or the problem.
// If the module is in the solution, back-tracks to before
// the module was added so that we can reconstruct the
// retained packages from before the module was added.
// If the module is in the problem, reconstructs the frontier.
// Removing a module does *not* release its transitive dependencies.
// We do not have enough information to be absolutely certain that nothing in
// the working copy also retains these modules.
func (state *State) Remove(ctx context.Context, loader SolverLoader, out SolverProgress, name string) (*State, error) {
	// Pre-solve to ensure that the dependees table is full and the frontier is
	// empty.
	state, err := state.Solve(ctx, loader, out)
	if err != nil {
		return nil, err
	}

	// Create a list of existing constraints, less those that are transitive
	// dependees of the package to be removed, including itself.
	constraints := make(Modules, 0, len(state.Frontier))
	dependees := state.Dependees.Transitive(StringSet{name: {}})
	for _, solution := range state.Solution {
		if !dependees.Has(solution.Module.Name) {
			constraints = append(constraints, solution.Module)
		}
	}

	state = NewState()

	state, err = state.Constrain(ctx, loader, out, constraints, false)
	if err != nil {
		return nil, err
	}

	// Move all entries from frontier to solution.
	state, err = state.Solve(ctx, loader, out)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// Modules returns a slice of Modules from both the solution and the unsolved
// modules in the problem, suitable for reconstructing a new lockfile even with
// a partial solution.
func (state *State) Modules() Modules {
	modules := make(Modules, 0, len(state.Frontier)+len(state.Solution))
	for _, name := range state.Solution.Names() {
		module := state.Solution[name].Module
		modules = append(modules, module)
	}
	for _, module := range state.Frontier {
		modules = append(modules, module)
	}
	sort.Sort(modules)
	return modules
}
