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

// file hash.go provides extensions for the Git hash plumbing type.

import "gopkg.in/src-d/go-git.v4/plumbing"

// NoHash is the zero value for a hash, used to distinguish whether a hash is
// present.
var NoHash = plumbing.Hash{}

// MaxHash is the maximum git hash value.
var MaxHash = plumbing.Hash{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

// HashString returns the hexadecimal string representation of a hash,
// or an empty string if the hash is not set.
// This is useful for ensuring that hashes are omitted from YAML files if they
// are not present.
func HashString(hash plumbing.Hash) string {
	if hash == NoHash {
		return ""
	}
	return hash.String()
}

// HashBefore compares two hashes numerically, to make hashes sortable.
func HashBefore(a, b plumbing.Hash) bool {
	for i := 0; i < 20; i++ {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}
	return false
}

// HashAfter returns whether one hash is more than another.
func HashAfter(a, b plumbing.Hash) bool {
	for i := 0; i < 20; i++ {
		if a[i] > b[i] {
			return true
		}
		if a[i] < b[i] {
			return false
		}
	}
	return false
}

// HashBetween returns whether a hash is between inclusive bounds.
func HashBetween(min, hash, max plumbing.Hash) bool {
	return !HashAfter(hash, max) && !HashAfter(min, hash)
}

// HashSub subtracts a hash from another.
func HashSub(a, b plumbing.Hash) plumbing.Hash {
	var (
		hash plumbing.Hash
		rem  int16
	)
	for i := 19; i >= 0; i-- {
		diff := int16(a[i]) - int16(b[i]) + rem
		hash[i] = byte(diff)
		rem = diff >> 8
	}
	return hash
}

// HashDiff computes the positive difference between two hashes.
func HashDiff(a, b plumbing.Hash) plumbing.Hash {
	if HashBefore(a, b) {
		return HashSub(b, a)
	}
	return HashSub(a, b)
}

// ParseHashPrefix parses a hash prefix, taking the given digits as the most
// significant digits of the resulting hash, or returns NoHash if the string is
// not a hash prefix.
//
// Example: for the string "abcde", the min is
// "abcde00000000000000000000000000000000000" and max is
// "abcdefffffffffffffffffffffffffffffffffff".
func ParseHashPrefix(str string) (min, max plumbing.Hash) {
	max = MaxHash

	runes := []rune(str)
	length := 40
	if len(runes) > 40 {
		return NoHash, NoHash
	}
	if len(runes) < length {
		length = len(runes)
	}
	for i := 0; i < length; i++ {
		r := runes[i]
		if r >= '0' && r <= '9' {
			val := byte(16 * (r - '0'))
			min[i/2] = val
			max[i/2] = val + 0xf
		} else if r >= 'a' && r <= 'f' {
			val := byte(16 * (r - 'a' + 10))
			min[i/2] = val
			max[i/2] = val + 0xf
		} else {
			return NoHash, NoHash
		}

		i++
		if i >= length {
			break
		}

		r = runes[i]
		if r >= '0' && r <= '9' {
			val := byte(r - '0')
			min[i/2] += val
			max[i/2] += val - 0xf
		} else if r >= 'a' && r <= 'f' {
			val := byte(r - 'a' + 10)
			min[i/2] += val
			max[i/2] += val - 0xf
		} else {
			return NoHash, NoHash
		}
	}
	return
}
