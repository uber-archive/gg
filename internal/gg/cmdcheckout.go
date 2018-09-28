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

const checkoutUsage UsageError = `Usage: gg checkout/co
Example: gg read-only checkout

Checks out the staged solution's vendor directory, replacing whatever was
previously there.  This is typically a preamble to writing a new glide.lock
with "gg write-only" or "gg wo" and is implied in the command "gg write" or
"gg w".
`

func checkoutCommand() Command {
	return Command{
		Names: []string{
			"checkout",
			"co",
		},
		Usage: checkoutUsage,
		Read:  true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			msg := "Checking out vendor"
			driver.err.Start(msg)
			err := Checkout(driver.err, driver.memo.GitDir, driver.next.Modules())
			driver.err.Stop(msg)
			return err
		},
	}
}
