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
	"os/signal"
	"strings"
)

const consoleUsage UsageError = `Usage: gg console/c
Example: gg read console

From the command arguments, opens an interactive gg command console.  The
console can interrupt a sequence of commands.

Example: gg read console write

gg> continue/c
gg> quit/q

During an interactive session, the "continue" or "c" command will resume the
sequence of commands on the command line, but "quit" or "q" will abort.
`

func consoleCommand() Command {
	return Command{
		Names: []string{
			"console",
			"c",
		},
		Usage: consoleUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			// Pre-cache own modules so they don't have to be read during auto-suggest.
			_, _, _ = driver.memo.ReadOwnPackages(ctx, driver.err)

			// Trap SIGINT
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt)
			defer func() {
				signal.Stop(ch)
			}()

			for {
				line, err := driver.rl.Readline()
				if err == io.EOF {
					return nil
				}
				if err != nil {
					return err
				}

				cmd, rest := NextToken(line)
				rest = strings.TrimSpace(rest)
				command := driver.commands[cmd]

				if command.OptionallyMonadic {
					// show-packages, push, and pull take an optional argument
					// in the readline mode.
					err := driverShowPackages(ctx, driver, rest)
					if err != nil {
						return err
					}

				} else if cmd == "quit" || cmd == "q" || cmd == "exit" || (cmd == "x" && rest == "") {
					// The quit command and its aliases terminate readline and abort
					// the workflow on the command line.
					// As an affordance for newcomer's expectations, the "x" command, which
					// is an alias for "execute", is also aliased to quit when there is no
					// argument in readline mode only.
					return fmt.Errorf("quit from console")

				} else if cmd == "continue" || cmd == "c" {
					// The continue command is a form of quit, but does not abort the
					// command line workflow.
					// Since the "c" alias for console is not useful when already in a
					// console, we repurpose it as an alias for "continue".
					return nil

				} else if cmd == "console" {
					// Already there.

				} else if command.Niladic != nil {
					if rest != "" {
						_ = driver.ExecuteArguments(context.TODO(), "help", cmd)
					} else {
						graceful(ctx, ch, func(ctx context.Context) {
							err := command.Niladic(ctx, driver)
							if err != nil {
								fmt.Fprintf(driver.err, "Error executing %s: %s.\n", cmd, err)
							}
						})
					}

				} else if command.Monadic != nil {
					if rest == "" {
						_ = driver.ExecuteArguments(context.TODO(), "help", cmd)
					} else {
						graceful(ctx, ch, func(ctx context.Context) {
							err := command.Monadic(ctx, driver, rest)
							if err != nil {
								fmt.Fprintf(driver.err, "Error executing %s: %s.\n", cmd, err)
							}
						})
					}

				} else {
					fmt.Fprintf(driver.err, "%s", ggUsage)
					fmt.Fprintf(driver.err, "Unrecognized command %q.\n", cmd)
				}
			}
		},
	}
}

func graceful(ctx context.Context, ch chan os.Signal, f func(context.Context)) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()
	f(ctx)
}
