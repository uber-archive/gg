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
	"os"
	"os/exec"
	"runtime/pprof"
	"strings"

	"github.com/chzyer/readline"
)

// Driver captures the state of a command session for manipulating
// dependencies, including the next and previous states of the dependency
// solution, the module loader memo, a readline, and the associated IO streams.
type Driver struct {
	rl      *readline.Instance
	memo    *Memo
	prev    *State
	next    *State
	history []*State
	future  []*State
	in      io.Reader
	out     io.Writer
	err     *Progress

	commands  map[string]Command
	help      map[string]UsageError
	completer readline.AutoCompleter
}

// NewDriver returns a driver for both command line flags and readline mode.
func NewDriver(memo *Memo, in io.Reader, out, errout io.Writer) (*Driver, error) {
	pout, perr := NewProgress(out, errout)
	driver := &Driver{
		memo: memo,
		prev: NewState(),
		next: NewState(),
		in:   in,
		out:  pout,
		err:  perr,
	}

	commands, help, completer := AssembleCommands(driver, commands())
	driver.commands = commands
	driver.help = help

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       green + "gg>" + clear + " ",
		HistoryFile:  HistoryFile(),
		AutoComplete: completer,
	})
	if err != nil {
		return nil, err
	}
	driver.rl = rl

	return driver, nil
}

// HistoryFile returns the location of the readline history file, based on the
// XDG_CONFIG_DIR and HOME environment variables, whichever is present.
func HistoryFile() string {
	history := ""

	xdg := os.Getenv("XDG_CONFIG_DIR")
	if xdg != "" {
		return xdg + "/gg_history"
	}

	home := os.Getenv("HOME")
	if home != "" {
		return home + "/.gg_history"
	}

	return history
}

// ExecuteArguments executes command line arguments for the gg command.
func (driver *Driver) ExecuteArguments(ctx context.Context, args ...string) error {

	if len(args) == 0 {
		return nil
	}

	// Imply a read and write for some commands.
	// For example, "up", as a sole argument, is an alias for the read, update,
	// and rewrite workflow.  "up" as a part of a workflow, only upgrades the
	// state.
	command := driver.commands[args[0]]
	if (len(args) == 1 && command.Niladic != nil) || (len(args) == 2 && command.Monadic != nil) {
		if command.Write {
			args = append([]string{"read"}, append(args, "write")...)
		} else if command.Read {
			args = append([]string{"read-only"}, args...)
		}
	}

	// Execute command line flags in order, capturing the next flag as an
	// argument if the flagged command takes an argument.
	for len(args) > 0 {
		cmd := args[0]
		command := driver.commands[cmd]
		args = args[1:]
		if cmd == "cpuprofile" {
			profile := args[0]
			args = args[1:]
			if profile == "" {
				return cpuProfileUsage
			}

			f, err := os.Create(profile)
			if err != nil {
				return err
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				return err
			}
			defer pprof.StopCPUProfile()

		} else if command.OptionallyMonadic {
			// show-packages on the command line is niladic instead of monadic,
			// as an affordance for common usage, and to be less surprising
			// since its argument is optional in readline mode.
			err := command.Monadic(ctx, driver, "")
			if err != nil {
				return err
			}

		} else if command.Monadic != nil {
			if len(args) == 0 {
				return driver.ExecuteArguments(ctx, "help", cmd)
			}
			arg := args[0]
			args = args[1:]
			err := command.Monadic(ctx, driver, arg)
			if err != nil {
				return err
			}

		} else if command.Niladic != nil {
			err := command.Niladic(ctx, driver)
			if err != nil {
				return err
			}

		} else {
			fmt.Fprintf(driver.err, "%s\n", ggUsage)
			return fmt.Errorf("unrecognized command: %s", cmd)
		}
	}
	return nil
}

// ShellExec is a utility of Shell and Exec that runs a command in a subshell.
func (driver *Driver) ShellExec(command string, args ...string) error {
	env := os.Environ()

	env, stop := driver.StartServer(env)
	defer stop()

	quoted := []string{
		fmt.Sprintf("%q", command),
	}
	for _, arg := range args {
		quoted = append(quoted, fmt.Sprintf("%q", arg))
	}

	fmt.Fprintf(driver.err, "Executing %s.\n", strings.Join(quoted, " "))
	cmd := exec.Command(command, args...)
	cmd.Env = env
	cmd.Stdin = driver.in
	cmd.Stdout = driver.out
	cmd.Stderr = driver.err
	return cmd.Run()
}

// SuggestPackages auto-completes package names from known packages
// in the memo.
func (driver *Driver) SuggestPackages(_ string) []string {
	return driver.Packages(context.Background()).All.Keys()
}

// SuggestModules auto-completes module names from known modules in the memo.
func (driver *Driver) SuggestModules(_ string) []string {
	packages := make([]string, 0, len(driver.memo.Remotes))
	for name := range driver.memo.Remotes {
		packages = append(packages, name)
	}
	return packages
}

// Packages returns the union of packages in the solution and the working copy
func (driver *Driver) Packages(ctx context.Context) Packages {
	_, ownPackages, _ := driver.memo.ReadOwnPackages(ctx, DiscardProgress)
	modules := driver.next.Modules()
	_ = driver.memo.FinishPackages(ctx, DiscardProgress, modules)
	packages := modules.Packages()
	packages.Include(ownPackages)
	return packages
}

// FindSolutionOrExpressModule is a utility function for workflows that need to
// infer the version of a module the user expressed or implied, either by the
// versioned expressed after the "@" symbol in a package name, or implied by
// the version present in the next solution.
func (driver *Driver) FindSolutionOrExpressModule(ctx context.Context, name string, test bool) (Module, error) {
	memo := driver.memo
	state := driver.next

	var module Module
	if strings.Index(name, "@") >= 0 {
		found, err := memo.FindModule(ctx, driver.err, name, false)
		if err != nil {
			return module, err
		}
		module = found
	} else {
		solution, ok := state.Solution[name]
		if !ok {
			return module, fmt.Errorf("there is no version of package %s in the current solution", name)
		}
		module = solution.Module
	}

	return module, nil
}

func (driver *Driver) push(state *State) {
	driver.next = state
	driver.history = append(driver.history, state)
}
