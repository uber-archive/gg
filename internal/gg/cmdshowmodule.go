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

const showModuleUsage UsageError = `Usage: gg show-module/sm <module>
Example: gg show-module go.uber.org/fx

Shows the dependencies described in the given module's lock file.  Also shows
the version of that module in the staged dependency solution if it is different
from the exact version in the lock file.  Many modules do not have a lock file.

Shows any warning encountered when analyzing that module.
`

func showModuleCommand() Command {
	return Command{
		Names: []string{
			"show-module",
			"sm",
		},
		Usage:         showModuleUsage,
		Read:          true,
		SuggestModule: true,
		Monadic: func(ctx context.Context, driver *Driver, pkg string) error {
			module, err := driver.FindSolutionOrExpressModule(ctx, pkg, false)
			if err != nil {
				return err
			}
			if err := driver.memo.FinishModules(ctx, driver.err, module.Modules); err != nil {
				fmt.Fprintf(driver.err, "Failed to normalize the requirements of %s: %s\n", module.Summary(), err)
			}
			state := driver.next
			ShowModule(driver.out, state, module)
			return nil
		},
	}
}

// ShowModule writes a report about a particular revision of a module.
func ShowModule(out io.Writer, state *State, module Module) {
	fmt.Fprintf(out, "Package:    %s\n", module.Name)
	fmt.Fprintf(out, "Remote:     %s\n", module.Remote)
	fmt.Fprintf(out, "Hash:       %s\n", module.Hash)
	fmt.Fprintf(out, "Timestamp:  %s\n", module.Time)
	fmt.Fprintf(out, "Reference:  %s\n", module.Ref)
	if module.Changelog != NoHash {
		fmt.Fprintf(out, "CHANGELOG:  %s\n", module.Changelog)
	}
	if module.Glidelock != NoHash {
		fmt.Fprintf(out, "glide.lock: %s\n", module.Glidelock)
	}
	if module.Deplock != NoHash {
		fmt.Fprintf(out, "Gopkg.lock: %s\n", module.Deplock)
	}
	if module.GitoliteMirror {
		created := ""
		if module.GitoliteMirrorCreated {
			created = " (created)"
		}
		fmt.Fprintf(out, "Gitolite:   remote is a mirror %s\n", created)
	}
	if len(module.Warnings) > 0 {
		fmt.Fprintf(out, yellow+"Warnings:"+clear+"\n")
		for _, warning := range module.Warnings {
			fmt.Fprintf(out, yellow+"* %s"+clear+"\n", warning)
		}
	}
	if module.FetchError != nil {
		fmt.Printf(red+"Fetch error: %s"+clear+"\n", module.FetchError)
	}
	fmt.Fprintf(out, "Dependencies:\n")
	if len(module.Modules) == 0 {
		fmt.Fprintf(out, "* No locked modules.\n")
	}
	for _, want := range module.Modules {
		solution, ok := state.Solution[want.Name]
		got := solution.Module
		if !ok {
			fmt.Fprintf(out, "\n")
			fmt.Fprintf(out, red+"-   %s "+red+"(missing)"+clear+"\n", want)
		} else {
			showDependency(out, module, solution.Module, got)
		}
	}
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Dependees:\n")
	dependencies := state.Modules().FilterDependencies(module)
	if len(dependencies) == 0 {
		fmt.Fprintf(out, "* No dependees\n")
	}
	for _, dependency := range dependencies {
		showDependency(out, dependency.Module, dependency.Want, dependency.Got)
	}
}

func showDependency(out io.Writer, module, want, got Module) {
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "  %s depends on %s\n", module.Summary(), want.Summary())
	diffModule(out, want, got)
}

func diffModule(out io.Writer, want, got Module) {
	if got.Hash != want.Hash {
		fmt.Fprintf(out, gray+"- %s"+clear+"\n", want)
		if !want.CanUpgradeTo(got) {
			fmt.Fprintf(out, yellow+"+ %s "+yellow+"(conflict)"+clear+"\n", got)
		} else if want.Before(got) {
			fmt.Fprintf(out, green+"+ %s "+green+"(upgrade)"+clear+"\n", got)
		} else {
			fmt.Fprintf(out, red+"+ %s "+red+"(downgrade)"+clear+"\n", got)
		}
	} else {
		fmt.Fprintf(out, "= %s (same)\n", got)
	}
}
