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

import "unicode"

// NextToken scans through white space to return the first token, and all that
// remains after any following space.
// This is for the readline mode.
func NextToken(str string) (string, string) {
	for index, r := range str[0:] {
		if !unicode.IsSpace(r) {
			start := index
			for index, r := range str[start:] {
				if unicode.IsSpace(r) {
					end := start + index
					for index, r := range str[end:] {
						if !unicode.IsSpace(r) {
							return str[start:end], str[end+index:]
						}
					}
					return str[start:end], ""
				}
			}
			return str[start:], ""
		}
	}
	return "", str
}
