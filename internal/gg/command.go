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
)

// NiladicFunc defines the behavior of a command that takes no arguments.
type NiladicFunc func(ctx context.Context, driver *Driver) error

// MonadicFunc defines the behavior of a command that takes one argument.
type MonadicFunc func(ctx context.Context, driver *Driver, arg string) error

// Command is a definition of a command. A command may be niladic, monadic, or
// neither if it is just a hook for a help page.
type Command struct {
	Names             []string
	Usage             UsageError
	Niladic           NiladicFunc
	Monadic           MonadicFunc
	OptionallyMonadic bool
	SuggestModule     bool
	SuggestPackage    bool
	// Read means that, if this command is executed alone at the command line,
	// we must implicitly read (but not solve) first.
	Read bool
	// Write means that, if this command is executed alone at the command line,
	// we must implicitly read and solve before and write and checkout afterward.
	Write bool
}

// UsageError is a usage string that can be used as an error.
type UsageError string

// Error returns the usage as a string.
func (u UsageError) Error() string {
	return string(u)
}
