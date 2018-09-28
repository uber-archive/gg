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
	"strings"
)

// NewPatterns creates a glob-like pattern match and replace tool.
func NewPatterns(tuples [][2]string) Patterns {
	patterns := make(Patterns, 0, len(tuples))
	for _, tuple := range tuples {
		match := PatternSplit(tuple[0])
		replace := PatternSplit(tuple[1])
		patterns = append(patterns, Pattern{match, replace})
	}
	return patterns
}

// Patterns is an ordered list of pattern replacements.
type Patterns []Pattern

// Pattern is a directive to replace some terms, including some "*" or "..."
// wildcards, with the corresponding replacement terms.
type Pattern struct {
	Match   []string
	Replace []string
}

// Replace takes a string and returns the part of the string that matched a
// pattern and its corresponding replacement, or just a pair of empty strings
// if none of the patterns matched.
// Returns the index of the applied rule, or -1.
func (patterns Patterns) Replace(str string) (string, string, int) {
	parts := PatternSplit(str)
	for index, pattern := range patterns {
		if matched, replaced := pattern.replace(parts); replaced != nil {
			return strings.Join(matched, ""), strings.Join(replaced, ""), index
		}
	}
	return "", "", -1
}

// PatternSplit divides a string into "/", ":", "*", and "..." delimited
// components, including components for each delimiter.
func PatternSplit(str string) []string {
	// Capacity is a guess based on a pessimisstic average word
	// length of four.
	parts := make([]string, 0, len(str)/4)
	// Capacity is a pessimistic guess that the longest word
	// part is going to be about 32 runes long.
	// We perpetually recycle the allocation using the part[0:0]
	// notation.
	part := make([]rune, 0, 32)
	flush := func() {
		if len(part) > 0 {
			parts = append(parts, string(part))
			part = part[0:0]
		}
	}
	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '*' {
			flush()
			parts = append(parts, "*")
		} else if len(runes) > i+2 && runes[i] == '.' && runes[i+1] == '.' && runes[i+2] == '.' {
			flush()
			parts = append(parts, "...")
			i += 2
		} else if runes[i] == '/' || runes[i] == ':' {
			flush()
			parts = append(parts, string(runes[i]))
		} else {
			part = append(part, runes[i])
		}
	}
	flush()
	return parts
}

func (pattern Pattern) replace(parts []string) ([]string, []string) {
	if len(parts) < len(pattern.Match) {
		return nil, nil
	}
	matched := make([]string, 0, len(parts))
	wild := make([]string, 0, len(parts))
	for i, part := range pattern.Match {
		if part == "*" {
			matched = append(matched, parts[i])
			wild = append(wild, parts[i])
			continue
		} else if part == "..." {
			matched = append(matched, parts[i:]...)
			wild = append(wild, parts[i:]...)
			break
		} else if part != parts[i] {
			return nil, nil
		}
		matched = append(matched, part)
	}
	replaced := make([]string, 0, len(parts))
	for _, part := range pattern.Replace {
		if part == "*" {
			if len(wild) == 0 {
				return nil, nil
			}
			replaced = append(replaced, wild[0])
			wild = wild[1:]
		} else if part == "..." {
			replaced = append(replaced, wild...)
			wild = nil
		} else {
			replaced = append(replaced, part)
		}
	}
	return matched, replaced
}
