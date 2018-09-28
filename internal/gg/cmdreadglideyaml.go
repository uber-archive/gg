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

	"github.com/Masterminds/semver"
)

const readGlideManifestUsage UsageError = `Usage: gg read-glide-yaml/rgy
Example: gg rgl new rgy diff

Reads the project's glide.yaml and adds the newest revision of each dependency
that satisfies the given version predicate to the solution, then runs the
constraint solver for each of these dependencies to ensure their transitive
dependencies are captured.

The log output describes out gg interprets each constraint, and which versions
were added and removed as a consequence of the additional constraint.

1. If no version is specified:
   Selecting newest version or most recent timestamp of package.

2. If the version is an exact version like 1.0, 1.2.3, or v2:
   Selecting version M.m.p of package.

3. If the version is a glide version predicate like ^1 && !1.2:
   Selecting newest version of package that satisfies constraint.

4. Otherwise if the version is a name like master:
   Selecting version of package with shortest reference ending with branch name.
`

func readGlideManifestCommand() Command {
	return Command{
		Names: []string{
			"read-glide-yaml",
			"rgy",
		},
		Usage: readGlideManifestUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			state := driver.next
			prev := state

			manifest, err := ReadOwnGlideManifest()
			if err != nil {
				return fmt.Errorf("unable to read glide.yaml: %s", err)
			}

			add := func(imp GlideManifestImport, test bool) {
				if found, ok := findGlideManifestImport(ctx, driver.err, driver.memo, imp, test); ok {
					msg := fmt.Sprintf("Adding %s", found.Summary())
					driver.err.Start(msg)
					next, err := state.Add(ctx, driver.memo, driver.err, found)
					driver.err.Stop(msg)
					if err != nil {
						fmt.Fprintf(driver.err, "Failed to add %s\n", found.Summary())
					} else {
						state = next
					}
					ShowDiff(driver.err, prev.Modules(), state.Modules())
					prev = state
				} else {
					fmt.Fprintf(driver.err, "Unable to find a version of %s@%s\n", imp.Package, imp.Version)
				}
			}

			for _, imp := range manifest.Imports {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				add(imp, false)
			}
			for _, imp := range manifest.TestImports {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				add(imp, true)
			}

			driver.next = state
			return nil
		},
	}
}

func findGlideManifestImport(ctx context.Context, out ProgressWriter, memo *Memo, imp GlideManifestImport, test bool) (Module, bool) {
	module := Module{
		Name:   imp.Package,
		Test:   test,
		Remote: imp.Repo,
	}
	if err := memo.FinishRemote(ctx, out, &module); err != nil {
		fmt.Fprintf(out, "Cannot find module to import: cannot find remote for package %s: %s", module.Summary(), err)
		return Module{}, false
	}
	versions, err := memo.ReadVersions(ctx, out, module)
	if err != nil {
		fmt.Fprintf(out, "Unable to find versions of module %s: %s\n", imp.Package, err)
	}
	if imp.Version == "" {
		fmt.Fprintf(out, "Selecting newest version or most recent timestamp of %s.\n", imp.Package)
		return versions.FindBestVersion()
	} else if version := ParseVersion(imp.Version); version != NoVersion {
		fmt.Fprintf(out, "Selecting version %s of %s.\n", imp.Version, imp.Package)
		return versions.FindVersion(version)
	} else if constraint, err := semver.NewConstraint(imp.Version); err == nil {
		fmt.Fprintf(out, "Selecting newest version of %s that satisfies constraint %q.\n", imp.Package, imp.Version)
		return versions.FindBestSemver(constraint)
	} else if min, max := ParseHashPrefix(imp.Version); min != NoHash && max != NoHash {
		fmt.Fprintf(out, "Selecting newest version of %s with hash prefix %s.\n", imp.Package, imp.Version)
		return versions.FindHash(min, max)
	} else {
		fmt.Fprintf(out, "Selecting version of %s with shortest references ending with %q.\n", imp.Package, imp.Version)
		return versions.FindReference(imp.Version)
	}
}
