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
	"io"
	"io/ioutil"
	"sort"
	"time"

	"github.com/chzyer/readline"
)

const (
	// ProgressBarWidth is the number of cells in a rendered progress bar,
	// except for the "[", "@", and "]" glyphs.
	ProgressBarWidth = 20
	// ProgressFirstDelay is the minimum time between starting a workflow and
	// the first progress notification.
	ProgressFirstDelay = 1 * time.Second
	// ProgressBetweenDelay is the minimum time between progress notifications.
	ProgressBetweenDelay = time.Second
)

// DiscardProgress sends progress reports to /dev/null
var DiscardProgress = &Progress{writer: ioutil.Discard}

// NewProgress produces an output and progress/error writer pair, such that
// output written to either will redraw progress bars after the cursor.
func NewProgress(out, err io.Writer) (io.Writer, *Progress) {
	perr := &Progress{writer: err}
	pout := &progressOut{err: perr, out: out}
	return pout, perr
}

// Progress wraps an io.Writer that satisfies all of the progress indication
// methods of the upgrade, add missing, and solve workflows, and debounces
// progress bar indications.
type Progress struct {
	writer io.Writer
	last   time.Time
	etas   ETAs
}

// ETA represents a progress bar
type ETA struct {
	msg   string
	num   int
	tot   int
	start time.Time
}

// String draws a progress bar with a message, ratio of completed subtasks,
// estimated time of completion, and estimated time remaining.
func (eta ETA) String(now time.Time) string {
	var line string
	if eta.num == 0 {
		line = fmt.Sprintf("[.....................] %s", eta.msg)
	} else {
		i := 0
		bar := "["
		for ; i < ProgressBarWidth*eta.num/eta.tot; i++ {
			bar += "_"
		}
		bar += "@"
		for ; i < ProgressBarWidth; i++ {
			bar += "."
		}
		bar += "]"

		circa := "~"
		if eta.num == eta.tot {
			circa = ""
		}

		line = fmt.Sprintf("%s %d/%s%d %s", bar, eta.num, circa, eta.tot, eta.msg)
	}

	width := readline.GetScreenWidth()
	if len(line) > width {
		line = line[:width]
	}

	return line
}

// ETAs are an orderable list of ETAs.
type ETAs []ETA

// Len is the length of the slice.
func (etas ETAs) Len() int {
	return len(etas)
}

// Less returns whether ETAs at respective indicies are out of order.
func (etas ETAs) Less(i, j int) bool {
	return etas[i].start.Before(etas[j].start)
}

// Swap swaps the ETAs at given indices.
func (etas ETAs) Swap(i, j int) {
	etas[i], etas[j] = etas[j], etas[i]
}

// Update either adds or updates an ETA with the same message.
func (etas ETAs) Update(out ETA) ETAs {
	outs := make(ETAs, 0, len(etas))
	found := false
	for _, in := range etas {
		if in.msg == out.msg {
			outs = append(outs, out)
			found = true
		} else {
			outs = append(outs, in)
		}
	}
	if !found {
		outs = append(outs, out)
		sort.Sort(outs)
	}
	return outs
}

// Remove removes all ETAs with the same message.
func (etas ETAs) Remove(msg string) ETAs {
	outs := make(ETAs, 0, len(etas))
	for _, in := range etas {
		if in.msg != msg {
			outs = append(outs, in)
		}
	}
	return outs
}

// progressOut is a wrapper around an output writer, which hooks into a
// progress writer to ensure that it redraws after the cursor advances.
type progressOut struct {
	out io.Writer
	err *Progress
}

// ShowState is a solver hook, but does nothing for runtime progress bars.
func (p Progress) ShowState(_ *State) {}

// Consider is a solver hook, but does nothing for runtime progress bars.
func (p Progress) Consider(_ *State, module Module) {
	// fmt.Fprintf(p, "Consider:   %s\n", module)
}

// Constrain is a solver hook, but does nothing for runtime progress bars.
func (p Progress) Constrain(_ *State, module Module) {
	// fmt.Fprintf(p, "Constraint: %s\n", module)
}

// Backtrack is a solver hook, but does nothing for runtime progress bars.
func (p Progress) Backtrack(_ *State, prev, next Module) {
	// fmt.Fprintf(p, "Back tracking: %v %v\n", prev.Before(next), HashBefore(prev.Hash, next.Hash))
	// fmt.Fprintf(p, "- %s\n", prev)
	// fmt.Fprintf(p, "+ %s\n", next)
}

// Start indicates that progress has begun for a process of indeterminate duration.
func (p *Progress) Start(msg string) {
	p.erase()
	now := time.Now()
	p.etas = p.etas.Update(ETA{msg: msg, start: now})
	p.draw(now)
}

// Stop indicates that progress has stopped for a process.
func (p *Progress) Stop(msg string) {
	p.erase()
	p.etas = p.etas.Remove(msg)
	p.draw(time.Now())
}

// Progress draws an ETA and progress bar for the given process, number of
// completed steps, total steps, start time, and current time, if a progress
// indicator has not recently been drawn.
func (p *Progress) Progress(msg string, num, tot int, start, now time.Time) {
	p.erase()

	if num == 0 || tot == 0 || num == tot {
		p.etas = p.etas.Remove(msg)
	} else {
		p.etas = p.etas.Update(ETA{
			msg:   msg,
			num:   num,
			tot:   tot,
			start: start,
		})
	}

	p.draw(now)
}

func (p *Progress) erase() {
	for range p.etas {
		_, _ = p.writer.Write([]byte(lineUp + eraseLine))
	}
}

func (p *Progress) draw(now time.Time) {
	for _, eta := range p.etas {
		fmt.Fprintf(p.writer, "%s\n", eta.String(now))
	}
}

func (p progressOut) Write(bytes []byte) (int, error) {
	p.err.erase()
	count, err := p.out.Write(bytes)
	p.err.draw(time.Now())
	return count, err
}

func (p *Progress) Write(bytes []byte) (int, error) {
	p.erase()
	count, err := p.writer.Write(bytes)
	p.draw(time.Now())
	return count, err
}
