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
	"time"
)

// UpgradeLoader loads alternate versions of modules for the duration of an
// upgrade.
type UpgradeLoader interface {
	SolverLoader

	Fetch(context.Context, ProgressWriter, *Module, int) error
	DigestRefs(context.Context, ProgressWriter, Module) error
	ReadVersions(context.Context, ProgressWriter, Module) (Modules, error)
}

// UpgradeProgress provides progress notifications and warnings for the
// duration of an upgrade.
type UpgradeProgress interface {
	SolverProgress
}

// Upgrade takes a state and returns an upgraded state, using a loader to
// obtain alternate versions, and emitting progress events.
// The upgrader will promote any module to the latest version in its semantic
// version range.  Absent a version, the upgrader will promote any module to a
// revision with a newer commit timestamp and the same git reference.
// Otherwise, if the module does not have a known git reference or version, the upgrader
// will promote a revision to the latest known semantic version or any revision
// with a newer commit timestamp on the master branch.
func Upgrade(ctx context.Context, loader UpgradeLoader, out UpgradeProgress, state *State) (*State, error) {
	start := time.Now()
	reviewed := make(StringSet)
	var done bool
	for !done {
		done = true
		modules := state.Modules()
		for _, module := range modules {
			if _, ok := reviewed[module.Name]; ok {
				continue
			}
			done = false
			reviewed[module.Name] = struct{}{}

			// Report progress and ETA
			num := len(reviewed)
			tot := len(modules)
			now := time.Now()
			out.Progress("Upgrading", num, tot, start, now)

			next, err := upgradeModule(ctx, loader, out, state, module)
			if err != nil {
				return nil, err
			}
			state = next
		}
	}
	return state, nil
}

func upgradeModule(ctx context.Context, loader UpgradeLoader, out UpgradeProgress, state *State, module Module) (*State, error) {
	if err := loader.Fetch(ctx, out, &module, FetchMaxAttempts); err != nil {
		fmt.Fprintf(out, "warning while attempting to fetch %s: %s\n", module.Summary(), err)
	}
	if err := loader.DigestRefs(ctx, out, module); err != nil {
		fmt.Fprintf(out, "warning while attempting to digest references %s: %s\n", module.Summary(), err)
	}

	modules, err := loader.ReadVersions(ctx, out, module)
	if err != nil {
		return state, err
	}
	upgrade := findUpgradeModule(modules, module)
	if upgrade.Equal(module) {
		return state, nil
	}
	return state.Add(ctx, loader, out, upgrade)
}

func findUpgradeModule(modules Modules, module Module) Module {
	for _, upgrade := range modules {
		if module.CanUpgradeTo(upgrade) {
			module = upgrade
		}
	}
	return module
}
