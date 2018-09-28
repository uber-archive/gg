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

const writeDepLockUsage UsageError = `Usage: gg write-dep-lock/wdl
Example: gg r wdl

Writes Gopkg.lock, the lock file used by dep, from the staged solution.
`

func writeDepLockCommand() Command {
	return Command{
		Names: []string{
			"write-dep-lock",
			"wdl",
		},
		Usage: writeDepLockUsage,
		Read:  true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			msg := "Writing Gopkg.lock"
			driver.err.Start(msg)

			modules := driver.next.Modules()
			lock := DepLockFromModules(modules)
			err := WriteOwnDepLock(lock)

			driver.err.Stop(msg)
			return err
		},
	}
}
