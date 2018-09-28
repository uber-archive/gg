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
)

const pruneUsage UsageError = `Usage: gg prune
Example: read prune write

Removes every module in the solution that is not necessary to build any of the
commands or tests in the working copy.
`

func pruneCommand() Command {
	return Command{
		Names: []string{
			"prune",
		},
		Usage: pruneUsage,
		Write: true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			_, packages, err := driver.memo.ReadOwnPackages(ctx, driver.err)
			if err != nil {
				return err
			}

			state := driver.next
			modules := state.Modules()
			if err := driver.memo.FinishPackages(ctx, driver.err, modules); err != nil {
				return err
			}
			extraModules := ExtraModules(packages, modules.Packages(), state.Modules())
			for _, module := range extraModules {
				next, err := state.Remove(ctx, driver.memo, driver.err, module.Name)
				if err != nil {
					return err
				}
				state = next
			}
			driver.push(state)

			ShowDiff(driver.out, driver.prev.Modules(), driver.next.Modules())
			return nil
		},
	}
}
