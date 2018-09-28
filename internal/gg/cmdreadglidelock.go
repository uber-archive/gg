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
	"time"
)

const readOnlyUsage UsageError = `Usage: gg read-only/ro
Usage: read-glide-lock/rgl
Example: gg ro co

Reads glide.lock into the stage, but does not run the constraint solver to
verify that the cached solution is consistent or correct.  This is a common
expedient for checking out the locked versions.  The command "gg read-only
checkout" or "gg ro co" is equivalent to "gg install" or effectively "glide
install".
`

func readOnlyCommand() Command {
	return Command{
		Names: []string{
			"read-only",
			"ro",
			"read-glide-lock",
			"rgl",
		},
		Usage: readOnlyUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			memo := driver.memo
			state := NewState()
			out := driver.err

			out.Start("Reading glide.lock")
			defer out.Stop("Reading glide.lock")
			modules, err := memo.ReadOwnModules(ctx, out)
			if err != nil {
				return err
			}

			start := time.Now()
			out.Start("Reading git references")
			for i, module := range modules {
				err := memo.DigestRefs(ctx, out, module)
				if err != nil {
					fmt.Fprintf(driver.err, "Error digesting references for %s: %s\n", module.Summary(), err)
				}
				out.Progress("Reading git references", i, len(modules), start, time.Now())
			}
			out.Stop("Reading git references")

			out.Start("Merging constraints")
			next, err := state.Constrain(ctx, memo, out, modules, false)
			out.Stop("Merging constraints")
			if err != nil {
				return err
			}
			state = next

			driver.prev = state
			driver.push(state)
			return nil
		},
	}
}
