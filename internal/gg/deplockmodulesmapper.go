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
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

// ModulesFromDepLock converts a DepLock model to GG's internal Modules model.
func ModulesFromDepLock(lock *DepLock) (Modules, error) {
	modules := make(Modules, 0, len(lock.Projects))
	for _, project := range lock.Projects {
		module, err := moduleFromDepProject(project)
		if err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}
	return modules, nil
}

func moduleFromDepProject(project DepProject) (Module, error) {
	ref := ""
	if project.Version != "" {
		ref = "tags/" + project.Version
	} else if project.Branch != "" {
		ref = "heads/" + project.Branch
	}
	return Module{
		Name:    project.Name,
		Version: ParseVersion(project.Version),
		Ref:     ref,
		Hash:    plumbing.NewHash(project.Revision),
		Remote:  project.Source,
	}, nil
}

// DepLockFromModules converts GG's internal Modules model into a DepLock
// model, losing a great deal of information that would be cached in the
// current GlideLock model.
func DepLockFromModules(modules Modules) *DepLock {
	projects := make([]DepProject, 0, len(modules))
	for _, module := range modules {
		projects = append(projects, depProjectFromModule(module))
	}
	return &DepLock{
		Projects: projects,
	}
}

func depProjectFromModule(module Module) DepProject {
	branch := ""
	version := ""
	if strings.HasPrefix(module.Ref, "heads/") {
		branch = strings.TrimPrefix(module.Ref, "heads/")
	} else if strings.HasPrefix(module.Ref, "tags/") {
		version = strings.TrimPrefix(module.Ref, "tags/")
	} else if module.Version != NoVersion {
		version = "v" + module.Version.String()
	}
	return DepProject{
		Name:     module.Name,
		Version:  version,
		Branch:   branch,
		Revision: HashString(module.Hash),
		Source:   module.Remote,
	}
}
