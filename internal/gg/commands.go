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
	"github.com/chzyer/readline"
)

// Commands returns the entire gg command vocabulary.
func commands() []Command {
	return []Command{
		// Sorted
		addCommand(),
		addMissingCommand(),
		backCommand(),
		cpuProfileCommand(),
		changelogCommand(),
		checkoutCommand(),
		clearRemotesCacheCommand(),
		configCommand(),
		consoleCommand(),
		deplockCommand(),
		execCommand(),
		fetchCommand(),
		foreCommand(),
		gitCommand(),
		glidelockCommand(),
		helpCommand(),
		initCommand(),
		installCommand(),
		markCommand(),
		metricsCommand(),
		newCommand(),
		offlineCommand(),
		onlineCommand(),
		pruneCommand(),
		pullCommand(),
		pushCommand(),
		quietCommand(),
		readCommand(),
		readDepLockCommand(),
		readDepManifestCommand(),
		readGlideManifestCommand(),
		readOnlyCommand(),
		removeCommand(),
		resetCommand(),
		shellCommand(),
		showConflictsCommand(),
		showDiffCommand(),
		showExtraModulesCommand(),
		showImportsCommand(),
		showMissingPackagesCommand(),
		showModuleCommand(),
		showOwnPackagesCommand(),
		showPackagesCommand(),
		showRemotesCacheCommand(),
		showShallowSolutionCommand(),
		showSolutionCommand(),
		showVersionsCommand(),
		solveCommand(),
		traceCommand(),
		upgradeCommand(),
		versionCommand(),
		writeCommand(),
		writeDepLockCommand(),
		writeDepManifestCommand(),
		writeGlideManifestCommand(),
		writeOnlyCommand(),
	}
}

// AssembleCommands takes commands and builds tables for help, auto-complete, and routing.
func AssembleCommands(driver *Driver, commands []Command) (map[string]Command, map[string]UsageError, readline.AutoCompleter) {
	index := make(map[string]Command)
	usage := make(map[string]UsageError)
	// Triple the capacities over commands on the assumption that we'll have at most three aliases for each.
	items := make([]readline.PrefixCompleterInterface, 0, 3*len(commands))
	helpItems := make([]readline.PrefixCompleterInterface, 0, 3*len(commands))
	suggestModules := driver.SuggestModules
	suggestPackages := driver.SuggestPackages

	for _, command := range commands {
		for _, name := range command.Names {
			index[name] = command
			usage[name] = command.Usage

			var item readline.PrefixCompleterInterface
			if command.SuggestModule {
				item = readline.PcItem(name, readline.PcItemDynamic(suggestModules))
			} else if command.SuggestPackage {
				item = readline.PcItem(name, readline.PcItemDynamic(suggestPackages))
			} else {
				item = readline.PcItem(name)
			}
			items = append(items, item)
			helpItems = append(helpItems, readline.PcItem(name))
		}
	}

	items = append(items, readline.PcItem("help", helpItems...))
	completer := readline.NewPrefixCompleter(items...)

	return index, usage, completer
}
