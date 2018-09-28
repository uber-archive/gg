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

// GlideManifestFromModules converts modules to the glide.yaml manifest model.
func GlideManifestFromModules(modules Modules) *GlideManifest {
	// length of modules is the upper bound on both imports and test imports,
	// effectively 2x necessary allocation to avoid reallocation.
	imports := make([]GlideManifestImport, 0, len(modules))
	testImports := make([]GlideManifestImport, 0, len(modules))
	for _, module := range modules {
		version := ""
		if module.Version != NoVersion {
			if module.Version[0] == 0 {
				version = "~" + module.Version.String()
			} else {
				version = "^" + module.Version.String()
			}
		} else if strings.HasPrefix(module.Ref, "heads/") {
			version = strings.TrimPrefix(module.Ref, "heads/")
		}
		imp := GlideManifestImport{
			Package: module.Name,
			Version: version,
			Repo:    module.Remote,
		}
		if module.Test {
			testImports = append(testImports, imp)
		} else {
			imports = append(imports, imp)
		}
	}
	return &GlideManifest{
		Imports:     imports,
		TestImports: testImports,
	}
}
