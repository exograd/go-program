// Copyright (c) 2021 Nicolas Martyanoff <khaelin@gmail.com>
// Copyright (c) 2022 Exograd SAS.
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionParse(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		s  string
		id BuildId
	}{
		{"v0.0.0",
			BuildId{Major: 0, Minor: 0, Patch: 0}},
		{"v1.2.3",
			BuildId{Major: 1, Minor: 2, Patch: 3}},
		{"v10.2.314",
			BuildId{Major: 10, Minor: 2, Patch: 314}},
		{"v1.2.3-17-f1d2d2f",
			BuildId{Major: 1, Minor: 2, Patch: 3,
				NbCommits: optionalInt(17),
				Revision:  optionalString("f1d2d2f")}},
	}

	for _, test := range tests {
		var id BuildId
		if err := id.Parse(test.s); err != nil {
			t.Errorf("cannot parse %q: %v", test.s, err)
			continue
		}

		if assert.Equal(test.id, id, test.s) {
			assert.Equal(test.s, id.String(), test.s)
		}
	}
}

func optionalInt(i int) *int {
	return &i
}

func optionalString(s string) *string {
	return &s
}
