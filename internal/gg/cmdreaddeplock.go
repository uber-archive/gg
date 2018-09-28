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

const readDepLockUsage UsageError = `Usage: gg read-dep-lock/read-gopkg-lock/rdl
Example: gg rdl w

Reads Gopkg.lock, replacing the staged solution.

This does not however automatically run the constraint solver.
Follow-up with a "solve" command to ensure that the dependencies read are
complete and consistent.

Reading the dep lockfile is an alternative to the "gg read" command when
bootstrapping a project to use "gg" in a project that has a Gopkg.lock but no
glide.lock.  However, gg caches a lot more extra information in glide.lock to
speed up future runs, so writing and using a glide.lock as the source of truth
will perform better than reading Gopkg.lock in general.
`

func readDepLockCommand() Command {
	return Command{
		Names: []string{
			"read-dep-lock",
			"read-gopkg-lock",
			"rdl",
		},
		Usage: readDepLockUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			memo := driver.memo
			state := NewState()

			msg := "Reading Gopkg.lock"
			driver.err.Start(msg)
			lock, err := ReadOwnDepLock()
			driver.err.Stop(msg)
			if err != nil {
				return err
			}
			modules, err := ModulesFromDepLock(lock)
			if err != nil {
				return err
			}
			if err := memo.FinishModules(ctx, driver.err, modules); err != nil {
				return err
			}

			next, err := state.Constrain(ctx, memo, driver.err, modules, false)
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
