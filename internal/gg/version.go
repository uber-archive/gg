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

import "strconv"

// Version represents a major, minor, patch version number.
type Version [3]int

// NoVersion is the zero value of a version, an invalid version, indicating a
// parsed version that did not match any of the acceptable three number version
// patterns.
var NoVersion Version

// String returns a string representation of a version number or an empty
// string if the version is zero.
func (v Version) String() string {
	if v == NoVersion {
		return ""
	}
	return strconv.Itoa(v[0]) + "." + strconv.Itoa(v[1]) + "." + strconv.Itoa(v[2])
}

// Before returns whether a version precedes another version.
func (v Version) Before(w Version) bool {
	if v[0] != w[0] {
		return v[0] < w[0]
	}
	if v[1] != w[1] {
		return v[1] < w[1]
	}
	return v[2] < w[2]
}

// CanUpgradeTo returns whether a version can be validly upgraded to another
// version without breaking changes according to semantic versioning.
// Consequently, versions can only be upgraded to higher versions that
// are in the same compatibility window.
// Versions are compatible only if their major version number is the same, or
// if the major versions are zero and their minor version numbers are the same.
func (v Version) CanUpgradeTo(w Version) bool {
	return v.Before(w) && w.Before(v.nextBreak())
}

// nextBreak returns the next version number that allows for a breaking change
// according to semantic versioning.
func (v Version) nextBreak() Version {
	if v[0] == 0 {
		return Version{v[0], v[1] + 1, 0}
	}
	return Version{v[0] + 1, 0, 0}
}

// ParseVersion recognizes versions of the forms:
//  1
//  1.2
//  1.2.3
//  v1
//  v1.2
//  v1.2.3
// All other patterns return NoVersion including versions with beta or release
// candidate trailers, since these need to be compared by timestamp and do
// not imply compatibility ranges.
func ParseVersion(str string) (version Version) {
	runes := []rune(str)
	length := len(runes)
	index := 0
	digit := 0
	if index >= length {
		return NoVersion
	}
	if runes[index] == 'v' {
		index++
		if index >= length {
			return NoVersion
		}
	}
	for {
		// First digit
		if runes[index] == '0' {
			index++
			if index >= length {
				return version
			}
			if runes[index] != '.' {
				return NoVersion
			}
			// Expect "."
		} else if runes[index] >= '1' && runes[index] <= '9' {
			version[digit] = int(runes[index] - '0')
			index++
			// Subsequent digits
			for {
				if index >= length || runes[index] < '0' || runes[index] > '9' {
					break
				}
				version[digit] = version[digit]*10 + int(runes[index]-'0')
				index++
			}
		} else {
			return NoVersion
		}
		// Advance to next digit, until we have parsed three digits.
		digit++
		if digit >= 3 {
			// Trailing characters
			if index != length {
				return NoVersion
			}
			// Three digits parsed
			return version
		}
		// Possible termination before three digits accumulated.
		if index >= length {
			return version
		}
		// Expect dot
		if runes[index] != '.' {
			return NoVersion
		}
		// Consume dot
		index++
		if index >= length {
			return NoVersion
		}
	}
}
