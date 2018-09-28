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
	"time"
)

// ProgressWriter is an io.Writer with a Progress method for printing a
// progress indicator.
type ProgressWriter interface {
	io.Writer

	Progress(msg string, num, tot int, start, now time.Time)
	Start(msg string)
	Stop(msg string)
}
