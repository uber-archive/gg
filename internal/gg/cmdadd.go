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

const addUsage UsageError = `Usage: gg add/a <module>[@<spec>]
Usage: gg add-test/at <module>[@<spec>]
Example: gg add go.uber.org/fx@1

Reads glide.lock, adds or upgrades the version of the given module, checks out
vendor, and writes the new glide.lock.

The test variant adds a dependency for tests.  Test dependencies are only
needed to run tests in this package, so they don't get subsumed when a module
adds this module as a dependency.

The specifier can be a version like "v1.0.0" or merely "1", or it can be a
reference like "tags/v1.0.0" or "heads/master".  The specifier can be a hash or
hash prefix.  In the absence of a specifier, gg will choose the highest
version, and if there are no version references, will choose "heads/master".

An add commands alone on the command line implies reading glide.lock in before,
writing glide.lock out after, and checking out the new vendor.
`

func addCommand() Command {
	return Command{
		Names: []string{
			"add",
			"a",
		},
		Usage:         addUsage,
		SuggestModule: true,
		Write:         true,
		Monadic: func(ctx context.Context, driver *Driver, spec string) error {
			return Add(ctx, driver, spec, false)
		},
	}
}

func addTestCommand() Command {
	return Command{
		Names: []string{
			"add-test",
			"at",
		},
		Usage:         addUsage,
		SuggestModule: true,
		Monadic: func(ctx context.Context, driver *Driver, spec string) error {
			return Add(ctx, driver, spec, true)
		},
	}
}

// Add adds a module or test module to the solution, executing either the "add"
// or "add-test" command.
func Add(ctx context.Context, driver *Driver, spec string, test bool) error {
	memo := driver.memo
	state := driver.next

	module, err := memo.FindModule(ctx, driver.err, spec, test)
	if err != nil {
		return fmt.Errorf("Unable to add module %s: %s", spec, err)
	}

	state, err = state.Add(ctx, memo, driver.err, module)
	if err != nil {
		return fmt.Errorf("unable to add module %s: %s", module.Summary(), err)
	}
	driver.push(state)

	ShowDiff(driver.out, driver.prev.Modules(), driver.next.Modules())
	return nil
}
