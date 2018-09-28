// Copyright (c) 2018 Uber Technologies, Inc.
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
	"io"
)

const showConflictsUsage UsageError = `Usage: gg show-conflicts/sc
Example: gg read upgrade show-solution show-conflicts

This constraint solver can consistently solve any dependency graph by resolving
conflicts between dependencies by always choosing the newer of versions, either
by their implied sementic version range, or by the timestamp of the
corresponding git commit.  However, this algorithm can reveal inconsistencies,
where one of your dependencies may need a major upgrade to coexist in the
solution.  Building your package and running tests may reveal the
inconsistency.  The show-conflicts command can produce a list of suspects,
by revealing dependencies that depend on incompatible semantic versions of the
same package.

Also, glide produces glide.lock files that do not mention the version or branch
that was used to produce the hash for the locked hash.
In this state, the "gg upgrade" command will promote these hashes to the current
"master" branch, if newer than the existing commit hash by timestamp, but these
commits might not be in the history of the "master" branch.
The "gg show-conflicts" or "gg sc" command will reveal these stray dependencies,
which you can heal by explicitly adding them to the solution with "gg add" /
"gg a", "gg add-test" / "gg at", "gg ensure" / "gg e", or "gg ensure-test" /
"gg et".
`

func showConflictsCommand() Command {
	return Command{
		Names: []string{
			"show-conflicts",
			"sc",
		},
		Usage: showConflictsUsage,
		Read:  true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			ShowConflicts(driver.out, driver.next.Modules())
			return nil
		},
	}
}

// ShowConflicts writes a report about semantic version conflicts that have
// emerged in a dependency solution, where some modules depend on
// non-overlapping semantic version ranges of the same package.
func ShowConflicts(out io.Writer, modules Modules) {
	fmt.Fprintf(out, "Conflicts:\n")
	conflicts := false
	index := Modules(modules).Index()
	// Missing
	for _, module := range modules {
		for _, desired := range module.Modules {
			_, ok := index[desired.Name]
			if !ok {
				fmt.Fprintf(out, "* Found a missing dependency.\n")
				fmt.Fprintf(out, "  %s depends on %s\n", module.Name, desired.Name)
				fmt.Fprintf(out, "  FROM %s\n", module)
				fmt.Fprintf(out, "  LACK %s\n", desired)
				conflicts = true
			}
		}
	}
	// Absent
	for _, module := range modules {
		for _, desired := range module.Modules {
			required, ok := index[desired.Name]
			if ok && desired.Ref != "" && desired.Version != NoVersion && desired.Hash != required.Hash && !desired.CanUpgradeTo(required) {
				fmt.Fprintf(out, "* Found a potential version conflict.\n")
				fmt.Fprintf(out, "  %s is locked to a version of %s that may not be compatible with the completed solution, based on their versions.\n", module.Name, desired.Name)
				fmt.Fprintf(out, "  from %s\n", module)
				fmt.Fprintf(out, "  want %s\n", desired)
				fmt.Fprintf(out, "  got  %s\n", required)
				conflicts = true
			}
		}
	}
	if !conflicts {
		fmt.Fprintf(out, "* No conflicts.\n")
	}
}
