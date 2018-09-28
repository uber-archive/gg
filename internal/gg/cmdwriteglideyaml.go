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

import "context"

const writeGlideManifestUsage UsageError = `Usage: gg write-glide-yaml/wgy
Example: gg wgy

Writes a glide.yaml file using a heuristic.  The glide.yaml will have an import
or testImport rule for each module in the solution if the working copy imports
one of its modules directly.  The version will be a major or minor semantic
version range if the module has a version.  Otherwise, the version will be a
branch name.  Since glide does not merge pinned hash requirements well, the
version will otherwise will be omitted.

Alone on the command line, write-glide-yaml will implicitly read your existing
solution then translate that back out to a glide manifest.
`

func writeGlideManifestCommand() Command {
	return Command{
		Names: []string{
			"write-glide-yaml",
			"wgy",
		},
		Usage: writeGlideManifestUsage,
		Read:  true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			name, packages, err := driver.memo.ReadOwnPackages(ctx, driver.err)
			if err != nil {
				return err
			}

			former, _ := ReadOwnGlideManifest()

			state := driver.next
			modules := state.Modules()
			if err := driver.memo.FinishPackages(ctx, driver.err, modules); err != nil {
				return err
			}

			shallow := ShallowSolution(packages, modules)

			msg := "Writing glide.yaml"
			driver.err.Start(msg)
			manifest := GlideManifestFromModules(shallow)
			manifest.Package = name
			manifest.Homepage = former.Homepage
			manifest.License = former.License
			err = WriteOwnGlideManifest(manifest)
			driver.err.Stop(msg)
			return err
		},
	}
}
