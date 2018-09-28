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
	"os"
	"sort"
	"time"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

type FakeLoader map[plumbing.Hash]Module

func NewFakeLoader(modules Modules) FakeLoader {
	loader := make(FakeLoader)
	for _, module := range modules {
		module.Hash = module.LoaderHash()
		loader[module.Hash] = module
	}
	return loader
}

func (l FakeLoader) MustGetVersion(name string, version Version) Module {
	module, ok := l[Module{Name: name, Version: version}.LoaderHash()]
	if !ok {
		return Module{}
	}
	module.Finished = true
	return module
}

func (l FakeLoader) MustGetTestVersion(name string, version Version) Module {
	module, ok := l[Module{Name: name, Version: version}.LoaderHash()]
	if !ok {
		return Module{}
	}
	module.Test = true
	module.Finished = true
	return module
}

func (l FakeLoader) MustGetTime(name string, time time.Time) Module {
	for _, module := range l {
		if module.Name == name && module.Time.Equal(time) {
			return module
		}
	}
	return Module{}
}

func (l FakeLoader) Write(b []byte) (int, error) {
	return os.Stderr.Write(b)
}

func (l FakeLoader) FinishRemote(context.Context, ProgressWriter, *Module) error { return nil }

func (l FakeLoader) FinishModules(ctx context.Context, out ProgressWriter, modules Modules) error {
	for i, module := range modules {
		if module.Finished {
			continue
		}
		if err := l.FinishModule(ctx, out, &modules[i]); err != nil {
			return err
		}
	}
	return nil
}

func (l FakeLoader) FinishModule(ctx context.Context, out ProgressWriter, module *Module) error {
	module.Hash = module.LoaderHash()
	known, ok := l[module.Hash]
	if !ok {
		return fmt.Errorf("could not fetch module %s", module.Summary())
	}
	test := module.Test
	*module = known
	module.Test = test
	module.Finished = true
	return l.FinishModules(ctx, out, module.Modules)
}

func (l FakeLoader) FinishPackages(context.Context, ProgressWriter, Modules) error { return nil }

func (l FakeLoader) Fetch(context.Context, ProgressWriter, *Module, int) error { return nil }

func (l FakeLoader) DigestRefs(context.Context, ProgressWriter, Module) error { return nil }

func (l FakeLoader) ReadVersions(_ context.Context, _ ProgressWriter, module Module) (Modules, error) {
	versions := make(Modules, 0, len(l))
	for _, version := range l {
		if module.Name == version.Name {
			if module.Test {
				version.Test = true
			}
			versions = append(versions, version)
		}
	}
	return versions, nil
}

type LogSolverProgress struct{}

func (p *LogSolverProgress) Write(b []byte) (int, error) {
	return os.Stderr.Write(b)
}

func (p *LogSolverProgress) Start(msg string) {}

func (p *LogSolverProgress) Stop(msg string) {}

func (p *LogSolverProgress) Progress(msg string, num, tot int, start, now time.Time) {}

func (p *LogSolverProgress) ShowState(state *State) {
	var locked, unlocked Modules
	for _, module := range state.Frontier {
		unlocked = append(unlocked, module)
	}
	for _, partial := range state.Solution {
		locked = append(locked, partial.Module)
	}
	sort.Sort(locked)
	fmt.Fprintf(p, "Unlocked %v Locked %v\n", unlocked, locked)
}

func (p *LogSolverProgress) Consider(state *State, module Module) {
	fmt.Fprintf(p, "Consider +%s\n", module.Summary())
}

func (p *LogSolverProgress) Constrain(state *State, module Module) {
	fmt.Fprintf(p, "Constrain +%s\n", module.Summary())
}

func (p *LogSolverProgress) Backtrack(state *State, prev, next Module) {
	fmt.Fprintf(p, "Backtrack -%s +%s\n", prev.Summary(), next.Summary())
}
