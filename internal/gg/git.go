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
	"io"
	"os"
	"os/exec"
	"time"
)

// GitFetchRootRemote fetches all of the tags and branches corresponding to a
// dependency with the given remote URL and "root" (remote repository cache
// key).
func GitFetchRootRemote(out io.Writer, gitDir string, root, remoteURL string) error {
	cmd := exec.Command(
		"git", "fetch", remoteURL,
		"+refs/heads/*:refs/vendor/"+root+"/heads/*",
		"+refs/tags/*:refs/vendor/"+root+"/tags/*",
		"-f", "--no-tags",
		"--recurse-submodules",
	)
	cmd.Env = GitEnv(gitDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

// GitPullVendorCache fetches all of the vendor references from the remote refs
// cache.
func GitPullVendorCache(out io.Writer, gitDir string, remoteURL string) error {
	cmd := exec.Command(
		"git", "fetch", remoteURL,
		"+refs/vendor/*:refs/vendor/*",
		"-f", "--no-tags",
		"--recurse-submodules",
	)
	cmd.Env = GitEnv(gitDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

// GitPushVendorCache pushes the vendor references from the cache to the
// remote, except those that cannot fast-forward.
func GitPushVendorCache(out io.Writer, gitDir string, remoteURL string) error {
	cmd := exec.Command("git", "push", remoteURL, "refs/vendor/*")
	cmd.Env = GitEnv(gitDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

// Checkout builds a vendor directory from the given modules in the git commit
// stage and then checks it out.
func Checkout(out ProgressWriter, gitDir string, modules Modules) error {
	env := GitEnv(gitDir)

	cmd := exec.Command("git", "read-tree", "--empty")
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Env = env
	err := cmd.Run()
	if err != nil {
		return err
	}

	start := time.Now()
	for i, mod := range modules {
		cmd := exec.Command(
			"git",
			"read-tree",
			"--prefix", "vendor/"+mod.Name,
			mod.Hash.String(),
		)
		cmd.Env = env
		cmd.Stdin = os.Stdin
		cmd.Stdout = out
		cmd.Stderr = out
		err = cmd.Run()
		if err != nil {
			return err
		}
		out.Progress("Staging modules", i+1, len(modules), start, time.Now())
	}

	// TODO look into using checkout-index flags to preserve vendor/.git and
	// avoid nuking it outright.
	out.Start("Removing stale vendor")
	cmd = exec.Command("rm", "-rf", "vendor")
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = out
	err = cmd.Run()
	out.Stop("Removing stale vendor")
	if err != nil {
		return err
	}

	out.Start("Writing staged vendor")
	cmd = exec.Command("git", "checkout-index", "-af")
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = out
	err = cmd.Run()
	out.Stop("Writing staged vendor")
	if err != nil {
		return err
	}

	return nil
}

// GitEnv creates an os environment for git commands that manipulate the .gg
// bare repository cache of dependency repositories.
func GitEnv(path string) []string {
	env := os.Environ()
	env = append(env, "GIT_DIR="+path)
	env = append(env, "GIT_WORK_TREE=.")
	env = append(env, "GIT_INDEX_FILE="+path+"/INDEX")
	return env
}
