// Copyright (c) 2018 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
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

import "time"

const (
	// GGVersion is the version of the gg command.
	GGVersion = "1.0.0-pre"
	// Stamp is a stamp for the "gg version" command to write.
	Stamp = "gg " + GGVersion
	// GGCachePath is the location of the ".gg" git bare repository in the
	// working copy.
	GGCachePath = ".gg"
	// FetchMaxAttempts is the maximum number of attempts that the memo will
	// execute to attempt to fetch new versions of a module.
	FetchMaxAttempts = 5
	// FetchFirstAttemptWait is the maximum duration of the time to wait
	// between the first and second attempt to fetch a package, followed with
	// exponential back-off and full jitter.
	FetchFirstAttemptWait = 5 * time.Second
	// FetchMaxAttemptWait is the maximum time to wait between attempts to
	// fetch a package, after exponential back-off has run its course.
	FetchMaxAttemptWait = time.Minute
)
