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

const writeOnlyUsage UsageError = `Usage: gg write-only/wo
Usage: gg write-glide-lock/wgl
Example: gg read upgrade write-only exec 'go test .' checkout

Writes the staged dependencies solution to glide.lock.  This is an expedient
for "gg write" or "gg w" if you have already previously checked out the
solution's vendor tree with "gg checkout" or "gg co".  Checking out the vendor
tree, running tests, and checking for conflicts are good practices before
writing a new glide.lock.
`

func writeOnlyCommand() Command {
	return Command{
		Names: []string{
			"write-only",
			"wo",
			"write-glide-lock",
			"wgl",
		},
		Usage: writeOnlyUsage,
		Read:  true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			modules := driver.next.Modules()
			if err := driver.memo.FinishPackages(ctx, driver.err, modules); err != nil {
				return err
			}

			msg := "Writing glide.lock"
			driver.err.Start(msg)
			driver.prev = driver.next
			err := WriteOwnModules(modules)
			driver.err.Stop(msg)
			return err
		},
	}
}
