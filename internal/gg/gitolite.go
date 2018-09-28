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
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GitoliteCreateMirror attempts to create a gitolite mirror.
// Do or do not; there is no try.
func GitoliteCreateMirror(out ProgressWriter, mirror string) error {
	status := fmt.Sprintf("Attempting to create gitolite mirror (may already exist): %s", mirror)
	out.Start(status)
	defer out.Stop(status)

	parts := strings.SplitN(mirror, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("Remote location for gitolite mirror must have one colon to separate remote from path: %s", mirror)
	}
	remote := parts[0]
	path := parts[1]

	cmd := exec.Command("ssh", remote, "create", path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = out
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(out, "Warning while attempting to create gitolite mirror: %s\n", err)
	}

	return nil
}
