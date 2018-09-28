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
	"io"
)

const showDiffUsage UsageError = `Usage: gg diff/d
Example: gg read upgrade diff

Shows the differences between the current proposed solution and the most
recently read lockfile.  Each module with a "+" prefix is part of the new
solution but absent in the old.  Each module with a "-" prefix is absent in the
new solution but present in the old.  Modules with a "-" followed by a "+" have
changed.
`

func showDiffCommand() Command {
	return Command{
		Names: []string{
			"diff",
			"d",
		},
		Usage: showDiffUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			ShowDiff(driver.out, driver.prev.Modules(), driver.next.Modules())
			return nil
		},
	}
}

// ShowDiff writes a colorized report of what modules have been added, removed,
// or changed revisions between two sets of modules.
func ShowDiff(out io.Writer, before Modules, after Modules) {
	fmt.Fprintf(out, "Differences:\n")
	b := before.Index()
	a := after.Index()
	names := before.Names()
	names.Include(after.Names())
	diff := false
	for _, name := range names.Keys() {
		if !a[name].Equal(b[name]) {
			diff = true
			if module, ok := b[name]; ok {
				fmt.Fprintf(out, "\x1b[31m- %s\x1b[0m\n", module)
			}
			if module, ok := a[name]; ok {
				fmt.Fprintf(out, "\x1b[32m+ %s\x1b[0m\n", module)
			}
		}
	}
	if !diff {
		fmt.Fprintf(out, "* No differences.\n")
	}
}
