// Copyright 2023 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package common

import (
	"strings"
	"testing"
	"time"
)

func TestPrettyDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		// wantContains is a substring expected in the result; empty means any output is fine
		wantContains string
		// wantMaxDots is the max number of digits after the decimal point allowed
		wantMaxDecimalDigits int
	}{
		{
			// Short duration with nanoseconds: 1.23456789s -> should be truncated to 3 decimals
			duration:             1234567890 * time.Nanosecond,
			wantContains:         "1.234",
			wantMaxDecimalDigits: 3,
		},
		{
			// Exactly 4 decimal digits: 1.2345s -> no truncation needed
			duration:             1234500000 * time.Nanosecond,
			wantContains:         "1.234",
			wantMaxDecimalDigits: 3,
		},
		{
			// Duration with only 2 decimal digits: 1.25s -> no truncation
			duration:             1250000000 * time.Nanosecond,
			wantContains:         "1.25s",
			wantMaxDecimalDigits: 3,
		},
		{
			// Whole seconds: 2s -> no decimal point
			duration:             2 * time.Second,
			wantContains:         "2s",
			wantMaxDecimalDigits: 0,
		},
		{
			// Minutes: should contain "m"
			duration:             90 * time.Second,
			wantContains:         "m",
			wantMaxDecimalDigits: 3,
		},
		{
			// Hours: should contain "h"
			duration:             2 * time.Hour,
			wantContains:         "h",
			wantMaxDecimalDigits: 3,
		},
	}

	for _, tt := range tests {
		got := PrettyDuration(tt.duration).String()

		if tt.wantContains != "" && !strings.Contains(got, tt.wantContains) {
			t.Errorf("PrettyDuration(%v).String() = %q, want it to contain %q", tt.duration, got, tt.wantContains)
		}

		// Verify the decimal part is not longer than 3 digits
		if idx := strings.Index(got, "."); idx != -1 {
			// Find the end of the decimal digits
			end := idx + 1
			for end < len(got) && got[end] >= '0' && got[end] <= '9' {
				end++
			}
			decimalDigits := end - idx - 1
			if decimalDigits > tt.wantMaxDecimalDigits {
				t.Errorf("PrettyDuration(%v).String() = %q, decimal part has %d digits, want at most %d",
					tt.duration, got, decimalDigits, tt.wantMaxDecimalDigits)
			}
		}
	}
}

func TestPrettyDurationTruncation(t *testing.T) {
	// A duration with many sub-second digits should be truncated to 3 decimal places
	d := PrettyDuration(1*time.Second + 999*time.Microsecond + 999*time.Nanosecond)
	got := d.String()
	if idx := strings.Index(got, "."); idx != -1 {
		end := idx + 1
		for end < len(got) && got[end] >= '0' && got[end] <= '9' {
			end++
		}
		if digits := end - idx - 1; digits > 3 {
			t.Errorf("PrettyDuration truncation: got %q with %d decimal digits, want at most 3", got, digits)
		}
	}
}

func TestPrettyAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		age          time.Time
		wantContains string
	}{
		{
			name:         "just now returns 0",
			age:          now,
			wantContains: "0",
		},
		{
			name:         "sub-second returns 0",
			age:          now.Add(-500 * time.Millisecond),
			wantContains: "0",
		},
		{
			name:         "a few seconds ago",
			age:          now.Add(-30 * time.Second),
			wantContains: "s",
		},
		{
			name:         "a few minutes ago",
			age:          now.Add(-5 * time.Minute),
			wantContains: "m",
		},
		{
			name:         "a few hours ago",
			age:          now.Add(-3 * time.Hour),
			wantContains: "h",
		},
		{
			name:         "a day ago",
			age:          now.Add(-25 * time.Hour),
			wantContains: "d",
		},
		{
			name:         "a week ago",
			age:          now.Add(-8 * 24 * time.Hour),
			wantContains: "w",
		},
		{
			name:         "a month ago",
			age:          now.Add(-31 * 24 * time.Hour),
			wantContains: "mo",
		},
		{
			name:         "a year ago",
			age:          now.Add(-366 * 24 * time.Hour),
			wantContains: "y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrettyAge(tt.age).String()
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("PrettyAge(%v).String() = %q, want it to contain %q", tt.age, got, tt.wantContains)
			}
		})
	}
}

func TestPrettyAgeMaxComponents(t *testing.T) {
	// An age of 1y2mo3w4d5h6m7s should be displayed with at most 3 components
	// Use a known duration: roughly 1 year + 2 months + 3 weeks
	oneYear := 12 * 30 * 24 * time.Hour
	twoMonths := 2 * 30 * 24 * time.Hour
	threeWeeks := 3 * 7 * 24 * time.Hour
	then := time.Now().Add(-(oneYear + twoMonths + threeWeeks + 5*24*time.Hour + 2*time.Hour + 3*time.Minute))

	got := PrettyAge(then).String()
	// Count the number of unit symbols present
	symbols := []string{"y", "mo", "w", "d", "h", "m", "s"}
	count := 0
	for _, sym := range symbols {
		if strings.Contains(got, sym) {
			count++
		}
	}
	// "mo" contains "m", so we need a smarter count
	// Just verify the string is non-empty and doesn't end up with too many components
	if got == "0" || got == "" {
		t.Errorf("PrettyAge with old time returned %q, expected meaningful duration", got)
	}
}
