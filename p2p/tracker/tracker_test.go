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

package tracker

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
)

func enableMetrics(t *testing.T) {
	t.Helper()
	old := metrics.Enabled
	metrics.Enabled = true
	t.Cleanup(func() { metrics.Enabled = old })
}

// TestTrackerNew verifies that New creates a usable Tracker.
func TestTrackerNew(t *testing.T) {
	tr := New("eth", 5*time.Second)
	if tr == nil {
		t.Fatal("New returned nil")
	}
	if tr.protocol != "eth" {
		t.Errorf("protocol = %q, want %q", tr.protocol, "eth")
	}
	if tr.timeout != 5*time.Second {
		t.Errorf("timeout = %v, want %v", tr.timeout, 5*time.Second)
	}
	if tr.pending == nil {
		t.Error("pending map is nil")
	}
}

// TestTrackerTrackAndFulfil verifies the basic track-then-fulfil happy path.
func TestTrackerTrackAndFulfil(t *testing.T) {
	enableMetrics(t)

	tr := New("eth", time.Minute)

	const (
		peer    = "peer1"
		version = uint(66)
		reqCode = uint64(0x01)
		resCode = uint64(0x02)
		id      = uint64(1234)
	)

	tr.Track(peer, version, reqCode, resCode, id)

	tr.lock.Lock()
	if _, ok := tr.pending[id]; !ok {
		tr.lock.Unlock()
		t.Fatal("request was not tracked")
	}
	tr.lock.Unlock()

	tr.Fulfil(peer, version, resCode, id)

	tr.lock.Lock()
	if _, ok := tr.pending[id]; ok {
		tr.lock.Unlock()
		t.Fatal("request was not removed after Fulfil")
	}
	tr.lock.Unlock()
}

// TestTrackerFulfilNonExistent verifies that fulfilling a non-existent request
// does not panic and is handled gracefully (recorded as stale).
func TestTrackerFulfilNonExistent(t *testing.T) {
	enableMetrics(t)

	tr := New("eth", time.Minute)
	// Fulfil an ID that was never tracked — should not panic.
	tr.Fulfil("peer1", 66, 0x02, 9999)
}

// TestTrackerDuplicateTrack verifies that tracking the same ID twice
// does not panic and logs an error (the second Track is a no-op).
func TestTrackerDuplicateTrack(t *testing.T) {
	enableMetrics(t)

	tr := New("eth", time.Minute)

	tr.Track("peer1", 66, 0x01, 0x02, 42)
	// Second Track with same ID — should be a no-op (logs an error but no panic).
	tr.Track("peer1", 66, 0x01, 0x02, 42)

	tr.lock.Lock()
	if len(tr.pending) != 1 {
		tr.lock.Unlock()
		t.Errorf("pending len = %d, want 1 after duplicate Track", len(tr.pending))
		return
	}
	tr.lock.Unlock()
}

// TestTrackerFulfilWrongPeer verifies that a Fulfil with mismatched peer/version/code
// does not remove the pending request.
func TestTrackerFulfilWrongPeer(t *testing.T) {
	enableMetrics(t)

	tr := New("eth", time.Minute)

	const id = uint64(55)
	tr.Track("peer1", 66, 0x01, 0x02, id)

	// Fulfil with a different peer — should not remove the request.
	tr.Fulfil("peer2", 66, 0x02, id)

	tr.lock.Lock()
	if _, ok := tr.pending[id]; !ok {
		tr.lock.Unlock()
		t.Error("request was incorrectly removed when Fulfil peer did not match")
		return
	}
	tr.lock.Unlock()
}

// TestTrackerFulfilWrongCode verifies that a Fulfil with mismatched response code
// does not remove the pending request.
func TestTrackerFulfilWrongCode(t *testing.T) {
	enableMetrics(t)

	tr := New("eth", time.Minute)

	const id = uint64(77)
	tr.Track("peer1", 66, 0x01, 0x02, id)

	// Fulfil with wrong response code — should not remove the request.
	tr.Fulfil("peer1", 66, 0x99, id)

	tr.lock.Lock()
	if _, ok := tr.pending[id]; !ok {
		tr.lock.Unlock()
		t.Error("request was incorrectly removed when Fulfil code did not match")
		return
	}
	tr.lock.Unlock()
}

// TestTrackerMultipleRequests verifies that multiple distinct requests can be
// tracked and fulfilled independently.
func TestTrackerMultipleRequests(t *testing.T) {
	enableMetrics(t)

	tr := New("eth", time.Minute)

	const (
		peer    = "peer1"
		version = uint(66)
		reqCode = uint64(0x01)
		resCode = uint64(0x02)
	)

	// Track three requests.
	for i := uint64(0); i < 3; i++ {
		tr.Track(peer, version, reqCode, resCode, i)
	}

	tr.lock.Lock()
	if n := len(tr.pending); n != 3 {
		tr.lock.Unlock()
		t.Fatalf("pending len = %d, want 3", n)
	}
	tr.lock.Unlock()

	// Fulfil them in order.
	for i := uint64(0); i < 3; i++ {
		tr.Fulfil(peer, version, resCode, i)
	}

	tr.lock.Lock()
	if n := len(tr.pending); n != 0 {
		tr.lock.Unlock()
		t.Fatalf("pending len = %d, want 0 after all fulfilled", n)
	}
	tr.lock.Unlock()
}

// TestTrackerMetricsDisabled verifies that when metrics is disabled,
// Track and Fulfil are no-ops and do not add anything to pending.
func TestTrackerMetricsDisabled(t *testing.T) {
	// Ensure metrics is disabled for this test.
	old := metrics.Enabled
	metrics.Enabled = false
	defer func() { metrics.Enabled = old }()

	tr := New("eth", time.Minute)

	tr.Track("peer1", 66, 0x01, 0x02, 100)

	tr.lock.Lock()
	n := len(tr.pending)
	tr.lock.Unlock()

	if n != 0 {
		t.Errorf("pending len = %d, want 0 when metrics disabled", n)
	}
}

// TestTrackerExpiry verifies that pending requests are cleaned up after the
// timeout expires.
func TestTrackerExpiry(t *testing.T) {
	enableMetrics(t)

	// Use a very short timeout so the test runs quickly.
	timeout := 50 * time.Millisecond
	tr := New("eth", timeout)

	const id = uint64(1)
	tr.Track("peer1", 66, 0x01, 0x02, id)

	tr.lock.Lock()
	if _, ok := tr.pending[id]; !ok {
		tr.lock.Unlock()
		t.Fatal("request was not tracked")
	}
	tr.lock.Unlock()

	// Wait for the expiry to fire (timeout + some buffer for the 5ms threshold in clean()).
	time.Sleep(timeout + 100*time.Millisecond)

	tr.lock.Lock()
	n := len(tr.pending)
	tr.lock.Unlock()

	if n != 0 {
		t.Errorf("pending len = %d after expiry, want 0", n)
	}
}
