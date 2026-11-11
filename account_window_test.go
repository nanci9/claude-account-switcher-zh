// 中文社区版修改文件；变更概述见 docs/MODIFICATIONS.md。
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAccountWindowClosesWhenParentExits(t *testing.T) {
	reader, writer := io.Pipe()
	closed := make(chan struct{})
	go closeWhenParentExits(reader, func() { close(closed) })
	select {
	case <-closed:
		t.Fatal("window closed while parent was running")
	default:
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("window did not close after parent exited")
	}
}

func TestAccountWindowSecurityAndState(t *testing.T) {
	s := newSwitcher(t.TempDir())
	s.st.Accounts = []*Account{{UUID: "one", Name: "First"}, {UUID: "two", Name: "Second"}}
	s.active = "one"
	if err := os.MkdirAll(s.profileDir("one"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.profileDir("one"), "config.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	w := &accountWindow{s: s, token: "secret", address: "127.0.0.1:12345"}
	request := func(method, path, token, origin, host string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://"+w.address+path, strings.NewReader(`{"action":"switch","id":"unknown"}`))
		r.Host = host
		r.Header.Set("X-Window-Token", token)
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		out := httptest.NewRecorder()
		w.ServeHTTP(out, r)
		return out
	}
	if got := request("GET", "/api/state", "", "", w.address).Code; got != http.StatusForbidden {
		t.Fatalf("missing token: %d", got)
	}
	if got := request("GET", "/api/state", "secret", "", "evil.example").Code; got != http.StatusForbidden {
		t.Fatalf("invalid host: %d", got)
	}
	if got := request("POST", "/api/action", "secret", "https://evil.example", w.address).Code; got != http.StatusForbidden {
		t.Fatalf("cross origin: %d", got)
	}
	if got := request("POST", "/api/action", "secret", "http://"+w.address, w.address).Code; got != http.StatusBadRequest {
		t.Fatalf("unknown account: %d", got)
	}
	out := request("GET", "/api/state", "secret", "", w.address)
	if out.Code != http.StatusOK {
		t.Fatalf("state: %d", out.Code)
	}
	var state struct {
		Accounts []windowAccount `json:"accounts"`
	}
	if err := json.NewDecoder(out.Body).Decode(&state); err != nil {
		t.Fatal(err)
	}
	if len(state.Accounts) != 2 || !state.Accounts[0].Current || !state.Accounts[0].Saved || state.Accounts[1].Saved {
		t.Fatalf("wrong state: %+v", state.Accounts)
	}
}

func TestOrderedWindowAccounts(t *testing.T) {
	now := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	usage := func(n float64, reset time.Time) *Usage {
		return &Usage{Session: n, SessionReset: reset, CheckedAt: now}
	}
	s := newSwitcher(t.TempDir())
	snap := snapshot{Active: "current", Accounts: []Account{
		{UUID: "full-late", Usage: usage(100, now.Add(3*time.Hour))},
		{UUID: "unknown"},
		{UUID: "current", Usage: usage(90, now.Add(time.Hour))},
		{UUID: "full-unknown", Usage: usage(100, time.Time{})},
		{UUID: "half", Usage: usage(50, now.Add(time.Hour))},
		{UUID: "full-soon", Usage: usage(100, now.Add(time.Hour))},
		{UUID: "zero", Usage: usage(0, now.Add(time.Hour))},
	}}
	got := orderedWindowAccounts(s, snap, now)
	want := []string{"current", "zero", "half", "full-soon", "full-late", "full-unknown", "unknown"}
	if len(got) != len(want) {
		t.Fatalf("wrong count: %d", len(got))
	}
	for i, id := range want {
		if got[i].ID != id {
			t.Fatalf("position %d: got %q, want %q", i, got[i].ID, id)
		}
	}
	if got[len(got)-1].Session != nil {
		t.Fatal("missing usage must not appear as zero")
	}
}

func TestAccountWindowReuse(t *testing.T) {
	w := &accountWindow{}
	if !w.beginOpen(func(uintptr) bool { t.Fatal("no window should be focused yet"); return false }) {
		t.Fatal("initial open must create a window")
	}
	if w.beginOpen(func(uintptr) bool { t.Fatal("opening window must not be focused"); return false }) {
		t.Fatal("concurrent open must wait")
	}
	w.finishOpen(42)
	if w.beginOpen(func(h uintptr) bool {
		if h != 42 {
			t.Fatalf("wrong handle: %d", h)
		}
		return true
	}) {
		t.Fatal("existing window must be reused")
	}
	if !w.beginOpen(func(h uintptr) bool { return false }) {
		t.Fatal("closed window must be recreated")
	}
	w.finishOpen(0)
	if !w.beginOpen(func(uintptr) bool { return false }) {
		t.Fatal("failed creation must be retryable")
	}
}
