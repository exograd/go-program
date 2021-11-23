// Copyright (c) 2021 Nicolas Martyanoff <khaelin@gmail.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY
// SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR
// IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package program

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	buildIdRE *regexp.Regexp
)

func init() {
	digit := `(0|(?:[1-9][0-9]*))`
	version := `v` + digit + `.` + digit + `.` + digit
	nbCommits := `([1-9][0-9]*)`
	revision := `([a-z0-9]+)`

	buildIdRE =
		regexp.MustCompile(`^` + version +
			`(?:-` + nbCommits + `-` + revision + `)?$`)
}

type BuildId struct {
	Major int
	Minor int
	Patch int

	NbCommits *int
	Revision  *string
}

func (id BuildId) IsStable() bool {
	return id.NbCommits == nil && id.Revision == nil
}

func (id BuildId) String() string {
	s := fmt.Sprintf("v%d.%d.%d", id.Major, id.Minor, id.Patch)

	if !id.IsStable() {
		s += fmt.Sprintf("-%d-%s", *id.NbCommits, *id.Revision)
	}

	return s
}

func (id *BuildId) Parse(s string) error {
	matches := buildIdRE.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 || len(matches[0]) == 0 {
		return fmt.Errorf("invalid format")
	}

	id.Major, _ = strconv.Atoi(matches[0][1])
	id.Minor, _ = strconv.Atoi(matches[0][2])
	id.Patch, _ = strconv.Atoi(matches[0][3])

	if len(matches[0][4]) > 0 {
		n, _ := strconv.Atoi(matches[0][4])
		id.NbCommits = &n

		id.Revision = &matches[0][5]
	}

	return nil
}

func (id1 BuildId) Equal(id2 BuildId) bool {
	return id1.Major == id2.Major &&
		id1.Minor == id2.Minor &&
		id1.Patch == id2.Patch &&
		id1.NbCommits == id2.NbCommits &&
		id1.Revision == id2.Revision
}

func (id1 BuildId) LowerOrEqualTo(id2 BuildId) bool {
	if id1.Major < id2.Major {
		return true
	} else if id1.Major > id2.Major {
		return false
	}

	if id1.Minor < id2.Minor {
		return true
	} else if id1.Minor > id2.Minor {
		return false
	}

	if id1.Patch < id2.Patch {
		return true
	} else if id1.Patch > id2.Patch {
		return false
	}

	n1 := 0
	if id1.NbCommits != nil {
		n1 = *id1.NbCommits
	}

	n2 := 0
	if id2.NbCommits != nil {
		n2 = *id2.NbCommits
	}

	return n1 <= n2
}
