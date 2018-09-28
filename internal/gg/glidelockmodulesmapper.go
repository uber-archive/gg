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

	"gopkg.in/src-d/go-git.v4/plumbing"
)

// ModulesFromGlideLock converts a GlideLock model (as read from a glide.lock
// file) to the GG internal Modules model.
func ModulesFromGlideLock(lock *GlideLock) (Modules, error) {
	// Use an empty slice instead of nil to distinguish absent from empty.
	modules := make(Modules, 0, 5)
	if lock == nil {
		return modules, nil
	}
	for _, imp := range lock.Imports {
		module, err := moduleFromGlideLockImport(imp, false)
		if err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}
	for _, imp := range lock.TestImports {
		module, err := moduleFromGlideLockImport(imp, true)
		if err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}
	return modules, nil
}

// GlideLockFromModules converts the GG internal Modules model into the
// GlideLock model, suitable for writing to glide.lock.
func GlideLockFromModules(modules Modules) *GlideLock {
	// length of modules is the upper bound on both imports and test imports,
	// effectively 2x necessary allocation to avoid reallocation.
	imports := make([]GlideLockImport, 0, len(modules))
	testImports := make([]GlideLockImport, 0, len(modules))
	for _, module := range modules {
		imp := glideLockImportFromModule(module)
		if module.Test {
			testImports = append(testImports, imp)
		} else {
			imports = append(imports, imp)
		}
	}
	return &GlideLock{
		Generator:   Stamp,
		Imports:     imports,
		TestImports: testImports,
	}
}

func moduleFromGlideLockImport(imp GlideLockImport, test bool) (Module, error) {
	if imp.VCS != "" && imp.VCS != "git" {
		return Module{}, fmt.Errorf("VCS must be empty (or git) on all imports for gg")
	}

	commands := NewStringSet(imp.Commands)
	exports := NewStringSet(imp.Exports)
	imports, coImports := yamlToStringGraphs(imp.Imports)
	testImports, coTestImports := yamlToStringGraphs(imp.TestImports)

	all := make(StringSet)
	all.Include(commands)
	all.Include(exports)
	imports.SourcesIntoStringSet(all)
	coImports.SourcesIntoStringSet(all)
	testImports.SourcesIntoStringSet(all)
	coTestImports.SourcesIntoStringSet(all)

	return Module{
		Name:                  imp.Name,
		Version:               ParseVersion(imp.Revision),
		Hash:                  plumbing.NewHash(imp.Version),
		Time:                  imp.Time,
		Remote:                imp.Repo,
		Root:                  imp.Root,
		Ref:                   imp.Ref,
		Test:                  test,
		Modules:               modulesFromGlideLockRequirements(imp.Requirements),
		NoLock:                imp.NoRequirements,
		Warnings:              imp.Warnings,
		Changelog:             plumbing.NewHash(imp.Changelog),
		Glidelock:             plumbing.NewHash(imp.Glidelock),
		GitoliteMirror:        imp.GitoliteMirror,
		GitoliteMirrorCreated: imp.GitoliteMirrorCreated,
		Packages: Packages{
			All:           all,
			Commands:      commands,
			Exports:       exports,
			Imports:       imports,
			CoImports:     coImports,
			TestImports:   testImports,
			CoTestImports: coTestImports,
		},
	}, nil
}

func glideLockImportFromModule(module Module) GlideLockImport {
	return GlideLockImport{
		GlideLockRequirement:  glideLockRequirementFromModule(module),
		Warnings:              module.Warnings,
		Changelog:             HashString(module.Changelog),
		Glidelock:             HashString(module.Glidelock),
		GitoliteMirror:        module.GitoliteMirror,
		GitoliteMirrorCreated: module.GitoliteMirrorCreated,
		Requirements:          glideLockRequirementsFromModules(module.Modules),
		NoRequirements:        module.NoLock,
		Commands:              module.Packages.Commands.Keys(),
		Exports:               module.Packages.Exports.Keys(),
		Imports:               stringGraphToYAML(module.Packages.Imports),
		TestImports:           stringGraphToYAML(module.Packages.TestImports),
	}
}

func moduleFromGlideLockRequirement(requirement GlideLockRequirement) Module {
	return Module{
		Name:    requirement.Name,
		Version: ParseVersion(requirement.Revision),
		Hash:    plumbing.NewHash(requirement.Version),
		Time:    requirement.Time,
		Remote:  requirement.Repo,
		Root:    requirement.Root,
		Ref:     requirement.Ref,
	}
}

func glideLockRequirementFromModule(module Module) GlideLockRequirement {
	return GlideLockRequirement{
		Name:     module.Name,
		Revision: module.Version.String(),
		Version:  HashString(module.Hash),
		Time:     module.Time,
		Repo:     module.Remote,
		Root:     module.Root,
		Ref:      module.Ref,
	}
}

func modulesFromGlideLockRequirements(requirements []GlideLockRequirement) Modules {
	var modules Modules
	for _, requirement := range requirements {
		modules = append(modules, moduleFromGlideLockRequirement(requirement))
	}
	return modules
}

func glideLockRequirementsFromModules(modules Modules) []GlideLockRequirement {
	var requirements []GlideLockRequirement
	for _, module := range modules {
		if !module.Test {
			requirements = append(requirements, glideLockRequirementFromModule(module))
		}
	}
	return requirements
}

func stringGraphToYAML(g StringGraph) map[string][]string {
	yaml := make(map[string][]string)
	for key, values := range g {
		yaml[key] = values.Keys()
	}
	return yaml
}

func yamlToStringGraphs(yaml map[string][]string) (StringGraph, StringGraph) {
	graph := NewStringGraph()
	coGraph := NewStringGraph()
	for key, values := range yaml {
		for _, value := range values {
			graph.Add(key, value)
			coGraph.Add(value, key)
		}
	}
	return graph, coGraph
}
