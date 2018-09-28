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

import "strings"

// DepManifestFromModules converts the content of a Gopkg.toml into a slice of
// modules, using a heuristic to distinguish constraints (direct dependencies,
// those modules that contain packages imported by packages in the working
// copy) from overrides (modules from the transitive dependencies that have
// conflicts amongst themselves).
func DepManifestFromModules(manifest *DepManifest, modules Modules, ownPackages Packages) {
	// length of modules is the upper bound on both constraints and overrides,
	// effectively 2x necessary allocation to avoid reallocation.
	constraints := make([]DepManifestConstraint, 0, len(modules))
	overrides := make([]DepManifestConstraint, 0, len(modules))

	shallow := ShallowSolution(ownPackages, modules)
	index := shallow.Index()

	for _, module := range modules {
		constraint := DepManifestConstraint{
			Name: module.Name,
		}

		if module.Version != NoVersion {
			if module.Version[0] == 0 {
				constraint.Version = "~" + module.Version.String()
			} else {
				constraint.Version = "^" + module.Version.String()
			}
		} else if strings.HasPrefix(module.Ref, "heads/") {
			constraint.Branch = strings.TrimPrefix(module.Ref, "heads/")
		} else {
			constraint.Revision = module.Hash.String()
		}

		source := module.Remote
		if strings.HasPrefix(source, "http://") {
			source = strings.TrimPrefix(source, "http://")
		} else if strings.HasPrefix(source, "https://") {
			source = strings.TrimPrefix(source, "https://")
		}
		constraint.Source = source

		if modules.Conflicts(module) {
			overrides = append(overrides, constraint)
		} else if _, ok := index[module.Name]; ok {
			constraints = append(constraints, constraint)
		}
	}

	manifest.Constraints = constraints
	manifest.Overrides = overrides
}
