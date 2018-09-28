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
	"gopkg.in/src-d/go-git.v4/plumbing"
)

const readDepManifestUsage UsageError = `Usage: gg read-dep-toml/read-gopkg-toml/rdt

Reads the constraints and overrides in a dep Gopkg.toml and adds them to the
solution.

gg does not distinguish constraints and overrides, since it uses the
max-of-min-timestamps algorithm to settle conflicts among transitive
dependencies.
`

func readDepManifestCommand() Command {
	return Command{
		Names: []string{
			"read-dep-toml",
			"read-gopkg-toml",
			"rdt",
		},
		Usage: readDepManifestUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			state := driver.next
			prev := state

			manifest, err := ReadOwnDepManifest()
			if err != nil {
				return fmt.Errorf("unable to read Gopkg.toml: %s", err)
			}

			add := func(imp DepManifestConstraint) {
				if found, ok := findDepManifestConstraint(ctx, driver.err, driver.memo, imp); ok {
					msg := fmt.Sprintf("Adding %s.", found.Summary())
					driver.err.Start(msg)
					next, err := state.Add(ctx, driver.memo, driver.err, found)
					driver.err.Stop(msg)
					if err != nil {
						fmt.Fprintf(driver.err, "Failed to add %s.\n", found.Summary())
					} else {
						state = next
					}
					ShowDiff(driver.err, prev.Modules(), state.Modules())
					prev = state
				} else {
					fmt.Fprintf(driver.err, "Unable to find a version of %s.\n", imp.Name)
				}
			}

			for _, imp := range manifest.Overrides {
				add(imp)
			}
			for _, imp := range manifest.Constraints {
				add(imp)
			}

			driver.next = state
			return nil
		},
	}
}

func findDepManifestConstraint(ctx context.Context, out ProgressWriter, memo *Memo, imp DepManifestConstraint) (Module, bool) {
	module := Module{
		Name:   imp.Name,
		Remote: imp.Source,
	}
	if err := memo.FinishRemote(ctx, out, &module); err != nil {
		return Module{}, false
	}
	versions, err := memo.ReadVersions(ctx, out, module)
	if err != nil {
		fmt.Fprintf(out, "Unable to find versions of module %s: %s\n", imp.Name, err)
	}

	if imp.Revision != "" {
		hash := plumbing.NewHash(imp.Revision)
		return versions.FindHash(hash, hash)
	} else if imp.Version != "" {
		constraint, err := semver.NewConstraint(imp.Version)
		if err != nil {
			fmt.Fprintf(out, "Invalid predicate for constraint %s: %q", imp.Name, imp.Version)
			return Module{}, false
		}
		fmt.Fprintf(out, "Selecting newest version of %s that satisfies constraint %q.\n", imp.Name, imp.Version)
		return versions.FindBestSemver(constraint)
	} else if imp.Branch != "" {
		fmt.Fprintf(out, "Selecting version of %s with shortest references ending with %q.\n", imp.Name, imp.Branch)
		return versions.FindReference(imp.Branch)
	}
	fmt.Fprintf(out, "Selecting newest version or most recent timestamp of %s.\n", imp.Name)
	return versions.FindBestVersion()
}
