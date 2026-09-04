// Copyright (c) 2019, David Kitchen <david@buro9.com>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// * Neither the name of the organisation (Microcosm) nor the names of its
//   contributors may be used to endorse or promote products derived from
//   this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package css

import (
	"strings"
	"testing"
	"time"
)

// Regression test for https://github.com/microcosm-cc/bluemonday/issues/225.
// Whitespace-only tokens give recursiveCheck many equally valid split points;
// unmemoized it re-explored each one, taking ~4s for the value below. The
// verdict is unchanged (rejected either way), only the time to reach it.
func TestIssue225(t *testing.T) {
	// 20 blank tokens, as the issue #225 payload produced them.
	pathological := "normal\n" + strings.Repeat(" ", 20) + "16px\nroboto"

	const deadline = 3 * time.Second
	const want = false

	result := make(chan bool, 1)
	go func() {
		result <- FontHandler(pathological)
	}()

	select {
	case got := <-result:
		if got != want {
			t.Errorf("FontHandler(%q) = %v, want %v", pathological, got, want)
		}
	case <-time.After(deadline):
		t.Fatalf("FontHandler did not return within %v for a 20-blank-token "+
			"value; recursiveCheck may have regressed to exponential-time "+
			"behaviour (see issue #225)", deadline)
	}
}

// Memoization must not change which values FontHandler accepts, only how
// fast it decides. Expectations captured from the pre-fix behaviour.
func TestIssue225AcceptanceUnchanged(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		// Valid "font" shorthand values.
		{"italic bold 12px/30px Georgia, serif", true},
		{"icon", true},
		{"caption", true},
		{"normal 16px Roboto", true},
		{"bold 16px/1.5 Arial, sans-serif", true},
		// Invalid/unsafe values that must stay rejected.
		{"expression(alert(1))", false},
		{"javascript:alert(1)", false},
		{"url(evil.com)", false},
	}
	for _, tt := range tests {
		if got := FontHandler(tt.value); got != tt.want {
			t.Errorf("FontHandler(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}
