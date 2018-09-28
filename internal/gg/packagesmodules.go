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

// ExtraModules takes a modules solution, the import graph of the working copy,
// and the import graph of the solution, then computes based on what commands
// and tests are in the working copy, which modules are unnecessary.
func ExtraModules(ownPackages, packages Packages, modules Modules) Modules {
	packages = packages.Clone()
	packages.Include(ownPackages)
	imports, testImports := NecessaryPackages(ownPackages, packages)
	imports.Include(testImports)
	var extra Modules
	for _, module := range modules {
		if !module.Packages.Exports.Intersects(imports) {
			extra = append(extra, module)
		}
	}
	return extra
}

// ShallowSolution takes the import graph of the working copy and a modules
// solution, then computes the subset of the module solution that the working
// copy's commands and tests import directly.
// The result is the list of modules that would be specified in a manifest file
// like glide.yaml.
func ShallowSolution(ownPackages Packages, modules Modules) Modules {
	deps := make(Modules, 0, len(modules))
	for _, module := range modules {
		if ownPackages.CoImports.Intersects(module.Packages.Exports) {
			deps = append(deps, module)
		} else if ownPackages.CoTestImports.Intersects(module.Packages.Exports) {
			deps = append(deps, module)
		}
	}
	return deps
}
