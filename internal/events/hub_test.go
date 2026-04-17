package events

import (
	"testing"
	"time"
)

func recv(t *testing.T, ch <-chan Envelope) Envelope {
	t.Helper()
	select {
	case e := <-ch:
		return e
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for event")
		return Envelope{}
	}
}

func noRecv(t *testing.T, ch <-chan Envelope) {
	t.Helper()
	select {
	case e := <-ch:
		t.Fatalf("unexpected event: %+v", e)
	case <-time.After(20 * time.Millisecond):
	}
}

func TestEmit_DeliversToBothSubscribers(t *testing.T) {
	h := NewHub()
	c1 := &client{ch: make(chan Envelope, 16)}
	c2 := &client{ch: make(chan Envelope, 16)}
	h.add(sessionKeyPrefix+"sess1", c1)
	h.add(sessionKeyPrefix+"sess1", c2)

	h.Emit("sess1", Envelope{Type: EventRoundUpdated})

	e1 := recv(t, c1.ch)
	e2 := recv(t, c2.ch)
	if e1.Type != EventRoundUpdated || e2.Type != EventRoundUpdated {
		t.Fatalf("unexpected types: %q %q", e1.Type, e2.Type)
	}
}

func TestEmit_DoesNotCrossSession(t *testing.T) {
	h := NewHub()
	c1 := &client{ch: make(chan Envelope, 16)}
	c2 := &client{ch: make(chan Envelope, 16)}
	h.add(sessionKeyPrefix+"sess1", c1)
	h.add(sessionKeyPrefix+"sess2", c2)

	h.Emit("sess1", Envelope{Type: EventSessionUpdated})

	recv(t, c1.ch)
	noRecv(t, c2.ch)
}

func TestEmit_SlowClientDropsWithoutBlocking(t *testing.T) {
	h := NewHub()
	slow := &client{ch: make(chan Envelope, 1)} // buffer of 1 — fills after first event
	h.add(sessionKeyPrefix+"sess1", slow)

	start := time.Now()
	for i := 0; i < 10; i++ {
		h.Emit("sess1", Envelope{Type: EventLiveScore})
	}
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Fatalf("Emit blocked for %v", elapsed)
	}
}

func TestRemove_CleansUpSession(t *testing.T) {
	h := NewHub()
	c := &client{ch: make(chan Envelope, 16)}
	h.add(sessionKeyPrefix+"sess1", c)
	h.remove(sessionKeyPrefix+"sess1", c)

	h.mu.RLock()
	_, exists := h.subs[sessionKeyPrefix+"sess1"]
	h.mu.RUnlock()
	if exists {
		t.Fatal("session entry should be removed when last client leaves")
	}
}

func TestEmitToUser_DeliversToMultipleTabs(t *testing.T) {
	h := NewHub()
	c1 := &client{ch: make(chan Envelope, 16)}
	c2 := &client{ch: make(chan Envelope, 16)}
	h.add(userKeyPrefix+"user1", c1)
	h.add(userKeyPrefix+"user1", c2)

	h.EmitToUser("user1", Envelope{Type: EventInviteReceived})

	e1 := recv(t, c1.ch)
	e2 := recv(t, c2.ch)
	if e1.Type != EventInviteReceived || e2.Type != EventInviteReceived {
		t.Fatalf("unexpected types: %q %q", e1.Type, e2.Type)
	}
}

func TestEmitToUser_DoesNotCrossUser(t *testing.T) {
	h := NewHub()
	c1 := &client{ch: make(chan Envelope, 16)}
	c2 := &client{ch: make(chan Envelope, 16)}
	h.add(userKeyPrefix+"user1", c1)
	h.add(userKeyPrefix+"user2", c2)

	h.EmitToUser("user1", Envelope{Type: EventInviteReceived})

	recv(t, c1.ch)
	noRecv(t, c2.ch)
}

func TestEmitToUser_DoesNotCrossSession(t *testing.T) {
	h := NewHub()
	cs := &client{ch: make(chan Envelope, 16)}
	cu := &client{ch: make(chan Envelope, 16)}
	h.add(sessionKeyPrefix+"sess1", cs)
	h.add(userKeyPrefix+"sess1", cu) // same bare ID, different namespace

	h.Emit("sess1", Envelope{Type: EventSessionUpdated})

	recv(t, cs.ch)
	noRecv(t, cu.ch)
}

func TestRemove_ClosesChannel(t *testing.T) {
	h := NewHub()
	c := &client{ch: make(chan Envelope, 16)}
	h.add(sessionKeyPrefix+"sess1", c)
	h.remove(sessionKeyPrefix+"sess1", c)

	select {
	case _, open := <-c.ch:
		if open {
			t.Fatal("channel should be closed after remove")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("channel was not closed")
	}
}
