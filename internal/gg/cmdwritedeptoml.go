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

const writeDepManifestUsage UsageError = `Usage: gg write-dep-toml/wdt
Example: gg wdt

Writes a Gopkg.toml file using a heursitic.  The Gopkg.toml will have
an override for every module in the solution for which there is a version
conflict.  There will be a constraint for all remaining modules that
are direct dependencies of the packages in the working copy.
The constraints will prefer to produce a semver range from the module version,
will fall back to a branch name if the module's reference is known and starts
with heads/, and will then fall back to pinning a hash.

Pinning to a hash is supremely undesirable, and avoidable in most cases if gg
upgrades the solution.  This typically finds a reference or version number for
every module, if possible.

Alone on the command line, write-dep-toml will implicitly read the existing
solution then translate that to a Gopkg.toml.
`

func writeDepManifestCommand() Command {
	return Command{
		Names: []string{
			"write-dep-toml",
			"wdt",
		},
		Usage: writeDepManifestUsage,
		Read:  true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			_, packages, err := driver.memo.ReadOwnPackages(ctx, driver.err)
			if err != nil {
				return err
			}

			manifest, _ := ReadOwnDepManifest()

			state := driver.next
			modules := state.Modules()
			if err := driver.memo.FinishPackages(ctx, driver.err, modules); err != nil {
				return err
			}

			msg := "Writing Gopkg.toml"
			driver.err.Start(msg)

			DepManifestFromModules(manifest, modules, packages)

			err = WriteOwnDepManifest(manifest)
			driver.err.Stop(msg)
			return err
		},
	}
}
