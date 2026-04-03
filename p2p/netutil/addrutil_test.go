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

package netutil

import (
	"net"
	"testing"
)

func TestAddrIP(t *testing.T) {
	ip4 := net.ParseIP("192.168.1.1")
	ip6 := net.ParseIP("2001:db8::1")

	tests := []struct {
		name string
		addr net.Addr
		want net.IP
	}{
		{
			name: "IPAddr IPv4",
			addr: &net.IPAddr{IP: ip4},
			want: ip4,
		},
		{
			name: "IPAddr IPv6",
			addr: &net.IPAddr{IP: ip6},
			want: ip6,
		},
		{
			name: "TCPAddr IPv4",
			addr: &net.TCPAddr{IP: ip4, Port: 8080},
			want: ip4,
		},
		{
			name: "TCPAddr IPv6",
			addr: &net.TCPAddr{IP: ip6, Port: 30303},
			want: ip6,
		},
		{
			name: "UDPAddr IPv4",
			addr: &net.UDPAddr{IP: ip4, Port: 30303},
			want: ip4,
		},
		{
			name: "UDPAddr IPv6",
			addr: &net.UDPAddr{IP: ip6, Port: 30303},
			want: ip6,
		},
		{
			name: "nil addr returns nil",
			addr: nil,
			want: nil,
		},
		{
			name: "unknown addr type returns nil",
			addr: &net.UnixAddr{Name: "/tmp/test.sock", Net: "unix"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddrIP(tt.addr)
			if !got.Equal(tt.want) {
				t.Errorf("AddrIP(%v) = %v, want %v", tt.addr, got, tt.want)
			}
		})
	}
}
