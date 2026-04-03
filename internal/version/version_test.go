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

package version

import (
	"runtime/debug"
	"strings"
	"testing"
)

// TestBuildInfoVCS checks extraction of VCS information from debug.BuildInfo.
func TestBuildInfoVCS(t *testing.T) {
	tests := []struct {
		name     string
		settings []debug.BuildSetting
		wantOK   bool
		want     VCSInfo
	}{
		{
			name: "all fields present",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abcdef1234567890"},
				{Key: "vcs.time", Value: "2023-06-15T12:00:00Z"},
				{Key: "vcs.modified", Value: "false"},
			},
			wantOK: true,
			want: VCSInfo{
				Commit: "abcdef1234567890",
				Date:   "20230615",
				Dirty:  false,
			},
		},
		{
			name: "dirty build",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "deadbeef"},
				{Key: "vcs.time", Value: "2022-01-01T00:00:00Z"},
				{Key: "vcs.modified", Value: "true"},
			},
			wantOK: true,
			want: VCSInfo{
				Commit: "deadbeef",
				Date:   "20220101",
				Dirty:  true,
			},
		},
		{
			name: "missing revision returns not ok",
			settings: []debug.BuildSetting{
				{Key: "vcs.time", Value: "2023-06-15T12:00:00Z"},
			},
			wantOK: false,
		},
		{
			name: "missing date returns not ok",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc123"},
			},
			wantOK: false,
		},
		{
			name:     "empty settings",
			settings: nil,
			wantOK:   false,
		},
		{
			name: "invalid time format is skipped",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc123"},
				{Key: "vcs.time", Value: "not-a-time"},
			},
			// Date will be empty because the time didn't parse, so ok=false
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &debug.BuildInfo{Settings: tt.settings}
			got, ok := buildInfoVCS(info)
			if ok != tt.wantOK {
				t.Errorf("buildInfoVCS ok = %v, want %v", ok, tt.wantOK)
				return
			}
			if !tt.wantOK {
				return
			}
			if got.Commit != tt.want.Commit {
				t.Errorf("Commit = %q, want %q", got.Commit, tt.want.Commit)
			}
			if got.Date != tt.want.Date {
				t.Errorf("Date = %q, want %q", got.Date, tt.want.Date)
			}
			if got.Dirty != tt.want.Dirty {
				t.Errorf("Dirty = %v, want %v", got.Dirty, tt.want.Dirty)
			}
		})
	}
}

// TestVCS verifies that VCS returns a result without panicking.
// The exact result depends on whether the binary was built with VCS info,
// so we just verify it doesn't panic and returns a consistent type.
func TestVCS(t *testing.T) {
	info, ok := VCS()
	if ok {
		// If VCS info is available, commit should be non-empty.
		if info.Commit == "" {
			t.Error("VCS returned ok=true but Commit is empty")
		}
		if info.Date == "" {
			t.Error("VCS returned ok=true but Date is empty")
		}
	}
	// ok=false is also valid when VCS info is not embedded.
	_ = info
}

// TestClientName verifies that ClientName returns a properly formatted string.
func TestClientName(t *testing.T) {
	name := ClientName("geth")
	// ClientName should include the identifier (title-cased), a version, OS/arch, and Go version.
	if !strings.HasPrefix(name, "Geth/") {
		t.Errorf("ClientName(%q) = %q, want prefix %q", "geth", name, "Geth/")
	}
	// Should contain at least three "/" separators
	parts := strings.Split(name, "/")
	if len(parts) < 4 {
		t.Errorf("ClientName(%q) = %q, want at least 4 slash-separated parts", "geth", name)
	}
}

// TestInfo verifies that Info returns non-empty version and consistent vcs strings.
func TestInfo(t *testing.T) {
	version, vcs := Info()
	if version == "" {
		t.Error("Info() version is empty")
	}
	// vcs may be empty if VCS info was not embedded; that's fine.
	// If it is non-empty, it should contain a commit hash segment.
	if vcs != "" && !strings.Contains(vcs, "-") {
		t.Errorf("Info() vcs = %q, expected a dash-separated commit-date string", vcs)
	}
}
