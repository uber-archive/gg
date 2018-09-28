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

const showRemotesCacheUsage UsageError = `Usage: gg show-remotes-cache/src
Example: gg read src

Shows the known package remote URL cache.  The source of truth for this is an
HTTP request to the domain in the first term of the package name.  These get
captured in glide.lock, but can become stale.

See: gg help clear-remotes-cache
`

func showRemotesCacheCommand() Command {
	return Command{
		Names: []string{
			"show-remotes-cache",
			"src",
		},
		Usage: showRemotesCacheUsage,
		Read:  true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			remotes := driver.memo.Remotes
			fmt.Fprintf(driver.out, "Package remotes cache:\n")
			if len(remotes) == 0 {
				fmt.Fprintf(driver.out, "* Cache empty\n")
			}
			for _, name := range StringMapKeys(remotes) {
				fmt.Fprintf(driver.out, "* %s: %s\n", name, remotes[name])
			}
			return nil
		},
	}
}
