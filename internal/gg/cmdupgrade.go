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

const upgradeUsage UsageError = `Usage: gg upgrade/update/up/u
Example: gg up
Example: gg read up show-solution console

Upgrade runs an upgrade workflow that uses only the information in the staged
solution: the current modules, their version numbers, their branch names, their
timestamps, and all new references in the remote repository.

	Alternative upgrade approaches:
	
	To emulate a glide up:
		gg read-glide-yaml write

	To emulate a dep ensure -update:
		gg read-gopkg-toml write

	Both of these approaches create a lockfile from scratch and an empty cache.
	To prime the cache and write back to the cache, prefix these commands with
	read and new.

		gg read new rgy write

Upgrades dependencies. The upgrade command can be used stand-alone command line
argument or as a part of a sequence of commands.  If upgrade is used alone, it
will read the existing glide.lock, upgrade, checkout vendor, and write a new
glide.lock.

Scans all of the modules in the staged dependency solution, fetches upgrades
from their remote git repository, and upgrades each dependency if a qualified
version is available.

For modules that have a version tag, upgrade will use the most recent semantic
version in the same implied version range.  For example, a module with version
2.1.1 can be upgraded to any newer version that is less than 3.0.0.  A module
with version 0.1.2 can be upgraded to any version less than 0.2.0.

For modules that were previously added or upgraded by gg, we track the branch
name in glide.lock.  If there is a newer commit with the same branch name,
it will be ugpraded.

Otherwise, especially for modules written by glide.lock, we do not know the
original reference, so we assume that the git reference was "heads/master" and
upgrade accordingly.

gg uses the dependency solver to add any new modules or ugprade transitive
dependencies, by the rules defined by the solver.  The solver prefers the
higher semantic version.  In the absence of a semantic version, it uses the
module with the newer git commit timestamp.

An upgrade command alone on the command line implies reading glide.lock in
before, writing glide.lock out after, and checking out the new vendor.
`

func upgradeCommand() Command {
	return Command{
		Names: []string{
			"upgrade",
			"update",
			"up",
			"u",
		},
		Usage: upgradeUsage,
		Write: true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			memo := driver.memo
			state := driver.next

			driver.err.Start("Upgrading")
			defer driver.err.Stop("Upgrading")
			next, err := Upgrade(ctx, memo, driver.err, state)
			if err != nil {
				return err
			}
			state = next
			driver.push(state)

			ShowDiff(driver.out, driver.prev.Modules(), driver.next.Modules())
			return nil
		},
	}
}
