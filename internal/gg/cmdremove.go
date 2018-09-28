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
)

const removeUsage UsageError = `Usage: gg remove/rm <module>
Example: gg read remove go.uber.org/fx

Removes a module from the dependency solution.  Removing a module also
eliminates any module that depends on it, but does not necessarily remove the
modules it depends upon, so removing a package is not sufficient to revert an
add.

A remove command alone on the command line implies reading glide.lock in
before, writing glide.lock out after, and checking out the new vendor.
`

func removeCommand() Command {
	return Command{
		Names: []string{
			"remove",
			"rm",
		},
		Usage:         removeUsage,
		SuggestModule: true,
		Write:         true,
		Monadic: func(ctx context.Context, driver *Driver, name string) error {
			state := driver.next

			msg := fmt.Sprintf("Removing %s", name)
			driver.err.Start(msg)
			next, err := state.Remove(ctx, driver.memo, driver.err, name)
			driver.err.Stop(msg)
			if err != nil {
				return fmt.Errorf("unable to remove module %s: %s", name, err)
			}
			state = next
			driver.push(state)

			ShowDiff(driver.out, driver.prev.Modules(), driver.next.Modules())
			return nil
		},
	}
}
