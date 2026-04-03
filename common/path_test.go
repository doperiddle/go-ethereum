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
	"os"
	"path/filepath"
	"testing"
)

func TestFileExist(t *testing.T) {
	// Create a temporary file that exists
	tmpFile, err := os.CreateTemp(t.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	if !FileExist(tmpFile.Name()) {
		t.Errorf("FileExist(%q) = false, want true for an existing file", tmpFile.Name())
	}
}

func TestFileExistNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent_file_xyz")
	if FileExist(path) {
		t.Errorf("FileExist(%q) = true, want false for a non-existent file", path)
	}
}

func TestFileExistDirectory(t *testing.T) {
	// A directory also "exists", so FileExist should return true for it
	dir := t.TempDir()
	if !FileExist(dir) {
		t.Errorf("FileExist(%q) = false, want true for an existing directory", dir)
	}
}

func TestAbsolutePath(t *testing.T) {
	tests := []struct {
		datadir  string
		filename string
		want     string
	}{
		{
			// Relative filename is joined with datadir
			datadir:  "/home/user/data",
			filename: "keystore",
			want:     "/home/user/data/keystore",
		},
		{
			// Absolute filename is returned as-is
			datadir:  "/home/user/data",
			filename: "/etc/keystore",
			want:     "/etc/keystore",
		},
		{
			// Empty filename joined with datadir
			datadir:  "/home/user/data",
			filename: "",
			want:     "/home/user/data",
		},
		{
			// Nested relative path
			datadir:  "/data",
			filename: "subdir/file.json",
			want:     "/data/subdir/file.json",
		},
	}

	for _, tt := range tests {
		got := AbsolutePath(tt.datadir, tt.filename)
		if got != tt.want {
			t.Errorf("AbsolutePath(%q, %q) = %q, want %q", tt.datadir, tt.filename, got, tt.want)
		}
	}
}
