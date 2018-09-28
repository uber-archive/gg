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
	"time"
)

const metricsUsage UsageError = `Usage: metrics

Shows a report of performance metrics.
`

func metricsCommand() Command {
	return Command{
		Names: []string{
			"metrics",
		},
		Usage: metricsUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			ShowMetrics(driver.out, driver.memo)
			return nil
		},
	}
}

// ShowMetrics writes a report about the metrics gathered by the module loader
// for the dependency solver.
func ShowMetrics(out io.Writer, memo *Memo) {
	fmt.Fprintf(out, "GitFetchCalls: %v\n", memo.GitFetchCalls)
	fmt.Fprintf(out, "GitFetchTotalDuration: %v\n", memo.GitFetchDuration)
	if memo.GitFetchCalls > 0 {
		fmt.Fprintf(out, "GitFetchMeanDuration: %v\n", memo.GitFetchDuration/time.Duration(memo.GitFetchCalls))
	}

	fmt.Fprintf(out, "RemoteForPackageCalls: %v\n", memo.RemoteForPackageCalls)
	fmt.Fprintf(out, "RemoteForPackageTotalDuration: %v\n", memo.RemoteForPackageDuration)
	if memo.RemoteForPackageCalls > 0 {
		fmt.Fprintf(out, "RemoteForPackageMeanDuration: %v\n", memo.RemoteForPackageDuration/time.Duration(memo.RemoteForPackageCalls))
	}

	fmt.Fprintf(out, "GitCommitMemoHits: %d\n", memo.GitCommitMemoHits)
	fmt.Fprintf(out, "GitResolveCommitCalls: %v\n", memo.GitResolveCommitCalls)
	fmt.Fprintf(out, "GitResolveCommitTotalDuration: %v\n", memo.GitResolveCommitDuration)
	if memo.GitResolveCommitCalls > 1 {
		fmt.Fprintf(out, "GitResolveCommitMeanDuration: %v\n", memo.GitResolveCommitDuration/time.Duration(memo.GitResolveCommitCalls))
	}
	fmt.Fprintf(out, "GitDigestRefsCalls: %v\n", memo.GitDigestRefsCalls)
	fmt.Fprintf(out, "GitDigestRefsTotalDuration: %v\n", memo.GitDigestRefsDuration)
	if memo.GitDigestRefsCalls > 1 {
		fmt.Fprintf(out, "GitDigestRefsMeanDuration: %v\n", memo.GitDigestRefsDuration/time.Duration(memo.GitDigestRefsCalls))
	}
	fmt.Fprintf(out, "ReadOwnModulesCalls: %d\n", memo.ReadOwnModulesCalls)
	fmt.Fprintf(out, "ReadOwnModulesTotalDuration: %v\n", memo.ReadOwnModulesDuration)
	if memo.ReadOwnModulesCalls > 1 {
		fmt.Fprintf(out, "ReadOwnModulesMeanDuration: %v\n", memo.ReadOwnModulesDuration/time.Duration(memo.ReadOwnModulesCalls))
	}
	fmt.Fprintf(out, "ReadGitPackagesCalls: %d\n", memo.ReadGitPackagesCalls)
	fmt.Fprintf(out, "ReadGitPackagesTotalDuration: %v\n", memo.ReadGitPackagesDuration)
	if memo.ReadGitPackagesCalls > 1 {
		fmt.Fprintf(out, "ReadGitPackagesMeanDuration: %v\n", memo.ReadGitPackagesDuration/time.Duration(memo.ReadGitPackagesCalls))
	}
	fmt.Fprintf(out, "References: %d\n", len(memo.Refs))
	fmt.Fprintf(out, "PrimedRemotes: %d\n", memo.PrimedRemotes)
	fmt.Fprintf(out, "TotalRemotes: %d\n", len(memo.Remotes))
}
