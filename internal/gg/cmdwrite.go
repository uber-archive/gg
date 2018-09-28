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

const writeUsage UsageError = `Usage: gg write/w
Example: gg read upgrade write

Checks out the staged dependency solution into the vendor directory, replacing
whatever was previously there and writes a new glide.lock.  This is equivalent
to "gg checkout write-only" or "gg co wo".
`

func writeCommand() Command {
	return Command{
		Names: []string{
			"write",
			"w",
		},
		Usage: writeUsage,
		Read:  true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			return driver.ExecuteArguments(ctx, "checkout", "write-glide-lock")
		},
	}
}
